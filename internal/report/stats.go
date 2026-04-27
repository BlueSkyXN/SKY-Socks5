package report

import (
	"sort"
	"time"
)

type Source struct {
	URL        string `json:"url"`
	Candidates int    `json:"candidates"`
	Accepted   int    `json:"accepted"`
	Error      string `json:"error,omitempty"`
}

type Outputs struct {
	RawPath       string `json:"raw_path"`
	UniquePath    string `json:"unique_path"`
	ValidatedPath string `json:"validated_path"`
	MetadataPath  string `json:"metadata_path"`
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

type Metadata struct {
	GeneratedAt       time.Time `json:"generated_at"`
	StartedAt         time.Time `json:"started_at"`
	FinishedAt        time.Time `json:"finished_at"`
	Duration          string    `json:"duration"`
	ProbeURL          string    `json:"probe_url"`
	ValidateWorkers   int       `json:"validate_workers"`
	FetchTimeout      string    `json:"fetch_timeout"`
	ValidationTimeout string    `json:"validation_timeout"`
	Totals            Totals    `json:"totals"`
	Outputs           Outputs   `json:"outputs"`
	Sources           []Source  `json:"sources,omitempty"`
	Errors            []string  `json:"errors,omitempty"`
}

func (m *Metadata) Finalize(startedAt, finishedAt time.Time) {
	m.StartedAt = startedAt.UTC()
	m.FinishedAt = finishedAt.UTC()
	m.GeneratedAt = m.FinishedAt
	if !finishedAt.Before(startedAt) {
		m.Duration = finishedAt.Sub(startedAt).String()
	}
	if m.Totals.InvalidProxies == 0 && m.Totals.UniqueProxies >= m.Totals.ValidProxies {
		m.Totals.InvalidProxies = m.Totals.UniqueProxies - m.Totals.ValidProxies
	}
	sort.Slice(m.Sources, func(i, j int) bool {
		return m.Sources[i].URL < m.Sources[j].URL
	})
	sort.Strings(m.Errors)
}
