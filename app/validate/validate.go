package validate

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Options struct {
	ProbeURL      string
	Timeout       time.Duration
	Concurrency   int
	ValidStatuses map[int]struct{}
	Probe         ProbeFunc
}

type Result struct {
	Address        string        `json:"address"`
	Reachable      bool          `json:"reachable"`
	StatusCode     int           `json:"status_code,omitempty"`
	Error          string        `json:"error,omitempty"`
	ExitIP         string        `json:"exit_ip,omitempty"`
	ExitCountry    string        `json:"exit_country,omitempty"`
	CloudflareColo string        `json:"cloudflare_colo,omitempty"`
	CloudflareHTTP string        `json:"cloudflare_http,omitempty"`
	CloudflareTLS  string        `json:"cloudflare_tls,omitempty"`
	Duration       time.Duration `json:"-"`
	DurationMillis int64         `json:"duration_ms"`
}

type ProbeFunc func(ctx context.Context, address string, opts Options) Result

func Check(ctx context.Context, addresses []string, opts Options) ([]Result, error) {
	if opts.Concurrency <= 0 {
		return nil, fmt.Errorf("concurrency must be greater than zero")
	}
	if opts.ProbeURL == "" {
		return nil, fmt.Errorf("probe URL is required")
	}
	if len(opts.ValidStatuses) == 0 {
		return nil, fmt.Errorf("valid statuses are required")
	}
	if opts.Probe == nil {
		opts.Probe = probeHTTP
	}

	results := make([]Result, len(addresses))
	type job struct {
		index   int
		address string
	}
	jobs := make(chan job)
	var wg sync.WaitGroup

	for worker := 0; worker < opts.Concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range jobs {
				results[item.index] = opts.Probe(ctx, item.address, opts)
			}
		}()
	}
	for index, address := range addresses {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return results, ctx.Err()
		case jobs <- job{index: index, address: address}:
		}
	}
	close(jobs)
	wg.Wait()
	return results, nil
}

func probeHTTP(ctx context.Context, address string, opts Options) (result Result) {
	result.Address = address
	startedAt := time.Now()
	defer func() {
		result.Duration = time.Since(startedAt)
		result.DurationMillis = result.Duration.Milliseconds()
	}()

	proxyURL, err := url.Parse("socks5://" + address)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	defer transport.CloseIdleConnections()

	client := &http.Client{
		Timeout:   opts.Timeout,
		Transport: transport,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, opts.ProbeURL, nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	resp, err := client.Do(req)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.StatusCode = resp.StatusCode
	result.applyCloudflareTrace(body)
	if _, ok := opts.ValidStatuses[resp.StatusCode]; ok {
		result.Reachable = true
		return result
	}
	result.Error = fmt.Sprintf("unexpected status: %d", resp.StatusCode)
	return result
}

func (r *Result) applyCloudflareTrace(body []byte) {
	trace := parseCloudflareTrace(string(body))
	r.ExitIP = trace["ip"]
	r.ExitCountry = trace["loc"]
	r.CloudflareColo = trace["colo"]
	r.CloudflareHTTP = trace["http"]
	r.CloudflareTLS = trace["tls"]
}

func parseCloudflareTrace(body string) map[string]string {
	trace := make(map[string]string)
	for _, line := range strings.Split(body, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		trace[key] = value
	}
	return trace
}
