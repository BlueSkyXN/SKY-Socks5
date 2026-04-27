package validate

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	defaultProbeURL    = "https://one.one.one.one"
	defaultConcurrency = 10
	defaultTimeout     = 10 * time.Second
)

type Result struct {
	Address    string        `json:"address"`
	Reachable  bool          `json:"reachable"`
	StatusCode int           `json:"status_code,omitempty"`
	Duration   time.Duration `json:"duration"`
	Error      string        `json:"error,omitempty"`
}

type ProbeFunc func(ctx context.Context, address, probeURL string, timeout time.Duration) Result

type Options struct {
	ProbeURL    string
	Timeout     time.Duration
	Concurrency int
	Probe       ProbeFunc
}

func Check(ctx context.Context, addresses []string, opts Options) ([]Result, error) {
	opts = opts.withDefaults()
	if opts.Concurrency <= 0 {
		return nil, fmt.Errorf("concurrency must be greater than zero")
	}
	if opts.ProbeURL == "" {
		return nil, fmt.Errorf("probe URL must not be empty")
	}

	results := make([]Result, len(addresses))
	if len(addresses) == 0 {
		return results, nil
	}

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
				results[item.index] = opts.Probe(ctx, item.address, opts.ProbeURL, opts.Timeout)
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

func (o Options) withDefaults() Options {
	if o.ProbeURL == "" {
		o.ProbeURL = defaultProbeURL
	}
	if o.Timeout <= 0 {
		o.Timeout = defaultTimeout
	}
	if o.Concurrency <= 0 {
		o.Concurrency = defaultConcurrency
	}
	if o.Probe == nil {
		o.Probe = probeHTTP
	}
	return o
}

func probeHTTP(ctx context.Context, address, probeURL string, timeout time.Duration) Result {
	result := Result{Address: address}
	startedAt := time.Now()
	defer func() {
		result.Duration = time.Since(startedAt)
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
		Timeout:   timeout,
		Transport: transport,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, probeURL, nil)
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

	result.StatusCode = resp.StatusCode
	result.Reachable = resp.StatusCode == http.StatusOK
	if !result.Reachable {
		result.Error = fmt.Sprintf("unexpected status: %d", resp.StatusCode)
	}
	return result
}
