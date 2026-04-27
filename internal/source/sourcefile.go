package source

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
)

// LoadFile reads, validates, deduplicates, and sorts source URLs from a file.
func LoadFile(path string) ([]string, error) {
	cleanPath := strings.TrimSpace(path)
	if cleanPath == "" {
		return nil, errors.New("source file path is required")
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("open source file %q: %w", cleanPath, err)
	}
	defer file.Close()

	unique := make(map[string]struct{})
	scanner := bufio.NewScanner(file)

	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\uFEFF"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		rawURL := fields[0]
		if err := validateSourceURL(rawURL); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNumber, err)
		}

		unique[rawURL] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read source file %q: %w", cleanPath, err)
	}

	urls := make([]string, 0, len(unique))
	for rawURL := range unique {
		urls = append(urls, rawURL)
	}
	sort.Strings(urls)
	if len(urls) == 0 {
		return nil, fmt.Errorf("no source URLs found in %s", cleanPath)
	}

	return urls, nil
}

func validateSourceURL(rawURL string) error {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid source URL %q: %w", rawURL, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("invalid source URL %q: unsupported scheme %q", rawURL, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("invalid source URL %q: missing host", rawURL)
	}
	return nil
}
