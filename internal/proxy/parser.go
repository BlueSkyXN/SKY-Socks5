package proxy

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/BlueSkyXN/SKY-Socks5/internal/fetch"
)

// Parse converts a raw line into a normalized host:port proxy model.
func Parse(raw string) (Proxy, error) {
	token := normalizedToken(raw)
	if token == "" {
		return Proxy{}, errors.New("proxy line is empty")
	}
	if strings.HasPrefix(token, "#") {
		return Proxy{}, errors.New("proxy line is a comment")
	}

	host, portText, err := parseHostPort(token)
	if err != nil {
		return Proxy{}, err
	}

	port, err := strconv.Atoi(portText)
	if err != nil {
		return Proxy{}, fmt.Errorf("invalid port %q: %w", portText, err)
	}

	proxy := Proxy{
		Host: strings.ToLower(strings.TrimSpace(host)),
		Port: port,
	}
	if err := proxy.Validate(); err != nil {
		return Proxy{}, err
	}

	return proxy, nil
}

// Normalize returns the canonical host:port representation for a raw proxy line.
func Normalize(raw string) (string, error) {
	parsed, err := Parse(raw)
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}

// UniqueFromSourceResults parses, normalizes, and deduplicates all successful source results.
func UniqueFromSourceResults(results []fetch.SourceResult) ([]Proxy, NormalizationStats) {
	stats := NormalizationStats{
		SourceCount: len(results),
	}

	unique := make(map[string]Proxy)
	sourceIndex := make(map[string]map[string]struct{})

	for _, result := range results {
		if result.Err != nil || result.Error != "" {
			stats.FailedSourceCount++
			continue
		}

		for _, rawLine := range result.RawLines {
			stats.TotalRawLines++

			proxy, err := Parse(rawLine)
			if err != nil {
				stats.InvalidCount++
				continue
			}
			stats.ParsedCount++

			key := proxy.String()
			if _, exists := unique[key]; exists {
				stats.DuplicateCount++
			} else {
				unique[key] = proxy
			}

			if result.URL != "" {
				if _, exists := sourceIndex[key]; !exists {
					sourceIndex[key] = make(map[string]struct{})
				}
				sourceIndex[key][result.URL] = struct{}{}
			}
		}
	}

	proxies := make([]Proxy, 0, len(unique))
	for key, parsedProxy := range unique {
		if sources, exists := sourceIndex[key]; exists {
			parsedProxy.Sources = mapKeys(sources)
		}
		proxies = append(proxies, parsedProxy)
	}
	Sort(proxies)

	stats.UniqueCount = len(proxies)
	return proxies, stats
}

func normalizedToken(raw string) string {
	trimmed := strings.TrimSpace(strings.TrimPrefix(raw, "\uFEFF"))
	if trimmed == "" {
		return ""
	}

	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return ""
	}

	return strings.TrimRight(fields[0], ",;")
}

func parseHostPort(token string) (string, string, error) {
	if strings.Contains(token, "://") {
		parsedURL, err := url.Parse(token)
		if err != nil {
			return "", "", fmt.Errorf("invalid proxy URL %q: %w", token, err)
		}
		host := parsedURL.Hostname()
		port := parsedURL.Port()
		if host == "" || port == "" {
			return "", "", fmt.Errorf("proxy %q must include host and port", token)
		}
		return host, port, nil
	}

	if host, port, err := net.SplitHostPort(token); err == nil {
		return host, port, nil
	}

	if strings.Count(token, ":") > 1 && !strings.HasPrefix(token, "[") {
		return "", "", fmt.Errorf("IPv6 proxy %q must use bracket notation", token)
	}

	lastColon := strings.LastIndex(token, ":")
	if lastColon <= 0 || lastColon == len(token)-1 {
		return "", "", fmt.Errorf("proxy %q must be in host:port format", token)
	}

	return token[:lastColon], token[lastColon+1:], nil
}

func mapKeys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	return uniqueSortedStrings(result)
}
