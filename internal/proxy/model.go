package proxy

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
)

// Proxy is the normalized host:port representation used across the pipeline.
type Proxy struct {
	Host    string   `json:"host"`
	Port    int      `json:"port"`
	Sources []string `json:"sources,omitempty"`
}

// NormalizationStats summarizes parsing and deduplication work over fetched sources.
type NormalizationStats struct {
	SourceCount       int `json:"source_count"`
	FailedSourceCount int `json:"failed_source_count"`
	TotalRawLines     int `json:"total_raw_lines"`
	ParsedCount       int `json:"parsed_count"`
	UniqueCount       int `json:"unique_count"`
	DuplicateCount    int `json:"duplicate_count"`
	InvalidCount      int `json:"invalid_count"`
}

// String returns the canonical host:port form for the proxy.
func (p Proxy) String() string {
	return net.JoinHostPort(p.Host, strconv.Itoa(p.Port))
}

// Validate ensures the proxy model is usable.
func (p Proxy) Validate() error {
	if strings.TrimSpace(p.Host) == "" {
		return errors.New("host is required")
	}
	if p.Port < 1 || p.Port > 65535 {
		return fmt.Errorf("port %d is out of range", p.Port)
	}
	return nil
}

// Sort normalizes source lists and sorts proxies in place for deterministic output.
func Sort(proxies []Proxy) {
	for i := range proxies {
		proxies[i].Sources = uniqueSortedStrings(proxies[i].Sources)
	}

	sort.Slice(proxies, func(i, j int) bool {
		leftHost := strings.ToLower(proxies[i].Host)
		rightHost := strings.ToLower(proxies[j].Host)
		if leftHost == rightHost {
			return proxies[i].Port < proxies[j].Port
		}
		return leftHost < rightHost
	})
}

func uniqueSortedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	sort.Strings(result)
	if len(result) == 0 {
		return nil
	}
	return result
}
