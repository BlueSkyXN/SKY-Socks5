package report

import (
	"sort"
	"time"
)

type Source struct {
	URL            string `json:"url"`
	StatusCode     int    `json:"status_code,omitempty"`
	Candidates     int    `json:"candidates"`
	Accepted       int    `json:"accepted"`
	UniqueProxies  int    `json:"unique_proxies"`
	ValidProxies   int    `json:"valid_proxies"`
	Error          string `json:"error,omitempty"`
	DurationMillis int64  `json:"duration_ms"`
}

type Outputs struct {
	RawPath          string `json:"raw_path"`
	UniquePath       string `json:"unique_path"`
	ValidatedPath    string `json:"validated_path"`
	ValidatedCSVPath string `json:"validated_csv_path"`
	ReportPath       string `json:"report_path"`
}

type Totals struct {
	SourceURLs     int `json:"source_urls"`
	FetchErrors    int `json:"fetch_errors"`
	RawCandidates  int `json:"raw_candidates"`
	ParseErrors    int `json:"parse_errors"`
	ParsedProxies  int `json:"parsed_proxies"`
	UniqueProxies  int `json:"unique_proxies"`
	ValidProxies   int `json:"valid_proxies"`
	InvalidProxies int `json:"invalid_proxies"`
}

type ValidatedProxy struct {
	ProxyAddress             string            `json:"proxy_address"`
	EntryHost                string            `json:"entry_host"`
	EntryPort                int               `json:"entry_port"`
	SourceCount              int               `json:"source_count"`
	DuplicateCount           int               `json:"duplicate_count"`
	SourceURLs               []string          `json:"source_urls"`
	SourceCountry            string            `json:"source_country,omitempty"`
	SourceCity               string            `json:"source_city,omitempty"`
	ExitIP                   string            `json:"exit_ip,omitempty"`
	ExitCountry              string            `json:"exit_country,omitempty"`
	EntryExitSameIP          bool              `json:"entry_exit_same_ip"`
	SourceCountryMatchesExit string            `json:"source_country_matches_exit,omitempty"`
	CloudflareColo           string            `json:"cloudflare_colo,omitempty"`
	CloudflareHTTP           string            `json:"cloudflare_http,omitempty"`
	CloudflareFlags          []string          `json:"cloudflare_flags,omitempty"`
	StatusCode               int               `json:"status_code"`
	DurationMillis           int64             `json:"duration_ms"`
	ValidatedAt              string            `json:"validated_at"`
	ProbeURL                 string            `json:"probe_url"`
	ValidationTimeout        string            `json:"validation_timeout"`
	CloudflareTrace          map[string]string `json:"cloudflare_trace,omitempty"`
}

type Metadata struct {
	StartedAt         time.Time        `json:"started_at"`
	FinishedAt        time.Time        `json:"finished_at"`
	Duration          string           `json:"duration"`
	ProbeURL          string           `json:"probe_url"`
	FetchProxy        string           `json:"fetch_proxy,omitempty"`
	ValidStatuses     string           `json:"valid_statuses"`
	FetchTimeout      string           `json:"fetch_timeout"`
	ValidationTimeout string           `json:"validation_timeout"`
	ValidateWorkers   int              `json:"validate_workers"`
	Totals            Totals           `json:"totals"`
	Outputs           Outputs          `json:"outputs"`
	Sources           []Source         `json:"sources"`
	ValidatedProxies  []ValidatedProxy `json:"validated_proxies,omitempty"`
}

func (m *Metadata) Finalize(startedAt, finishedAt time.Time) {
	m.StartedAt = startedAt.UTC()
	m.FinishedAt = finishedAt.UTC()
	if !finishedAt.Before(startedAt) {
		m.Duration = finishedAt.Sub(startedAt).String()
	}
	m.Totals.InvalidProxies = m.Totals.UniqueProxies - m.Totals.ValidProxies
	sort.Slice(m.Sources, func(i, j int) bool {
		return m.Sources[i].URL < m.Sources[j].URL
	})
}
