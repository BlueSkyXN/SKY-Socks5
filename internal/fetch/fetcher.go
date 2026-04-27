package fetch

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// SourceResult contains the raw proxy lines returned by a single source URL.
type SourceResult struct {
	URL            string        `json:"url"`
	StatusCode     int           `json:"status_code,omitempty"`
	RawLines       []string      `json:"raw_lines,omitempty"`
	Error          string        `json:"error,omitempty"`
	Duration       time.Duration `json:"-"`
	DurationMillis int64         `json:"duration_ms,omitempty"`
	Err            error         `json:"-"`
}

// Successful reports whether the source returned a successful HTTP response.
func (r SourceResult) Successful() bool {
	return r.Err == nil && r.Error == "" && r.StatusCode >= http.StatusOK && r.StatusCode < http.StatusMultipleChoices
}

// FetchAll fetches all sources concurrently and returns a deterministic result order.
func FetchAll(ctx context.Context, urls []string, timeout time.Duration) ([]SourceResult, error) {
	if ctx == nil {
		return nil, errors.New("context is required")
	}
	if timeout <= 0 {
		return nil, errors.New("timeout must be greater than zero")
	}
	if len(urls) == 0 {
		return []SourceResult{}, nil
	}

	client := &http.Client{Timeout: timeout}
	results := make([]SourceResult, len(urls))

	var wg sync.WaitGroup
	for i, rawURL := range urls {
		i := i
		rawURL := strings.TrimSpace(rawURL)
		if rawURL == "" {
			results[i] = failedResult(rawURL, errors.New("source URL is empty"))
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = fetchOne(ctx, client, rawURL)
		}()
	}
	wg.Wait()

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].URL < results[j].URL
	})

	if err := ctx.Err(); err != nil {
		return results, err
	}
	return results, nil
}

// CombinedError joins source-specific failures into one error value.
func CombinedError(results []SourceResult) error {
	var errs []error
	for _, result := range results {
		if result.Err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", result.URL, result.Err))
			continue
		}
		if result.Error != "" {
			errs = append(errs, fmt.Errorf("%s: %s", result.URL, result.Error))
		}
	}
	return errors.Join(errs...)
}

func fetchOne(ctx context.Context, client *http.Client, rawURL string) (result SourceResult) {
	started := time.Now()
	result.URL = rawURL
	defer func() {
		result.Duration = time.Since(started)
		result.DurationMillis = result.Duration.Milliseconds()
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return failedResult(rawURL, fmt.Errorf("build request: %w", err))
	}

	resp, err := client.Do(req)
	if err != nil {
		return failedResult(rawURL, fmt.Errorf("fetch source: %w", err))
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return failedResultWithStatus(rawURL, resp.StatusCode, fmt.Errorf("unexpected HTTP status: %s", resp.Status))
	}

	lines, err := readLines(resp.Body)
	if err != nil {
		return failedResultWithStatus(rawURL, resp.StatusCode, fmt.Errorf("read response body: %w", err))
	}
	result.RawLines = lines
	return result
}

func failedResult(rawURL string, err error) SourceResult {
	return failedResultWithStatus(rawURL, 0, err)
}

func failedResultWithStatus(rawURL string, statusCode int, err error) SourceResult {
	return SourceResult{
		URL:        rawURL,
		StatusCode: statusCode,
		Error:      err.Error(),
		Err:        err,
	}
}

func readLines(reader io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)

	lines := make([]string, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\uFEFF"))
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
