package fetch

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

type Options struct {
	Timeout time.Duration
	Proxy   string
}

type Result struct {
	URL            string        `json:"url"`
	StatusCode     int           `json:"status_code,omitempty"`
	Lines          []string      `json:"-"`
	Error          string        `json:"error,omitempty"`
	Duration       time.Duration `json:"-"`
	DurationMillis int64         `json:"duration_ms"`
}

func All(ctx context.Context, urls []string, opts Options) []Result {
	if len(urls) == 0 {
		return nil
	}
	client := clientFor(opts)
	results := make([]Result, len(urls))

	var wg sync.WaitGroup
	for i, sourceURL := range urls {
		i := i
		sourceURL := strings.TrimSpace(sourceURL)
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = one(ctx, client, sourceURL)
		}()
	}
	wg.Wait()

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].URL < results[j].URL
	})
	return results
}

func clientFor(opts Options) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if opts.Proxy != "" {
		if parsed, err := url.Parse(opts.Proxy); err == nil {
			transport.Proxy = http.ProxyURL(parsed)
		}
	}
	return &http.Client{
		Timeout:   opts.Timeout,
		Transport: transport,
	}
}

func one(ctx context.Context, client *http.Client, sourceURL string) (result Result) {
	startedAt := time.Now()
	result.URL = sourceURL
	defer func() {
		result.Duration = time.Since(startedAt)
		result.DurationMillis = result.Duration.Milliseconds()
	}()

	if sourceURL == "" {
		result.Error = "source URL is empty"
		return result
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("build request: %v", err)
		return result
	}
	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("fetch source: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		result.Error = fmt.Sprintf("unexpected HTTP status: %s", resp.Status)
		return result
	}
	lines, err := readLines(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("read response body: %v", err)
		return result
	}
	result.Lines = lines
	return result
}

func readLines(reader io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)
	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\uFEFF"))
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines, scanner.Err()
}
