package proxy

import (
	"errors"
	"reflect"
	"testing"

	"github.com/BlueSkyXN/SKY-Socks5/internal/fetch"
)

func TestParseNormalizesProxies(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want Proxy
	}{
		{
			name: "plain host port",
			raw:  "127.0.0.1:1080",
			want: Proxy{Host: "127.0.0.1", Port: 1080},
		},
		{
			name: "scheme and mixed case host",
			raw:  "socks5://Example.COM:1081",
			want: Proxy{Host: "example.com", Port: 1081},
		},
		{
			name: "ipv6",
			raw:  "[2001:db8::1]:1080",
			want: Proxy{Host: "2001:db8::1", Port: 1080},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.raw)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", tt.raw, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Parse(%q) = %#v, want %#v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestParseRejectsInvalidProxy(t *testing.T) {
	tests := []string{
		"",
		"# comment",
		"example.com",
		"example.com:70000",
	}

	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			if _, err := Parse(raw); err == nil {
				t.Fatalf("Parse(%q) error = nil, want non-nil", raw)
			}
		})
	}
}

func TestUniqueFromSourceResults(t *testing.T) {
	results := []fetch.SourceResult{
		{
			URL:      "https://example.com/a.txt",
			RawLines: []string{"1.1.1.1:1080", "2.2.2.2:1080", "bad-line"},
		},
		{
			URL:      "https://example.com/b.txt",
			RawLines: []string{"1.1.1.1:1080", "SOCKS5://Example.com:1081"},
		},
		{
			URL:   "https://example.com/c.txt",
			Err:   errors.New("request failed"),
			Error: "request failed",
		},
	}

	proxies, stats := UniqueFromSourceResults(results)

	wantProxies := []Proxy{
		{Host: "1.1.1.1", Port: 1080, Sources: []string{"https://example.com/a.txt", "https://example.com/b.txt"}},
		{Host: "2.2.2.2", Port: 1080, Sources: []string{"https://example.com/a.txt"}},
		{Host: "example.com", Port: 1081, Sources: []string{"https://example.com/b.txt"}},
	}
	if !reflect.DeepEqual(proxies, wantProxies) {
		t.Fatalf("UniqueFromSourceResults() proxies = %#v, want %#v", proxies, wantProxies)
	}

	wantStats := NormalizationStats{
		SourceCount:       3,
		FailedSourceCount: 1,
		TotalRawLines:     5,
		ParsedCount:       4,
		UniqueCount:       3,
		DuplicateCount:    1,
		InvalidCount:      1,
	}
	if !reflect.DeepEqual(stats, wantStats) {
		t.Fatalf("UniqueFromSourceResults() stats = %#v, want %#v", stats, wantStats)
	}
}
