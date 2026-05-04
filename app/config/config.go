package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultSourceFile          = "configs/sources.txt"
	DefaultOutputDir           = "."
	DefaultRawOutputFile       = "raw_proxies.txt"
	DefaultUniqueOutputFile    = "unique_proxies.txt"
	DefaultValidatedOutputFile = "validated_proxies.txt"
	DefaultReportOutputFile    = "proxy_report.json"
	DefaultProbeURL            = "https://one.one.one.one"
	DefaultFetchTimeout        = 15 * time.Second
	DefaultValidateTimeout     = 10 * time.Second
	DefaultValidateConcurrency = 100
	DefaultValidStatuses       = "200"
)

type Config struct {
	SourceFile          string
	OutputDir           string
	RawOutputPath       string
	UniqueOutputPath    string
	ValidatedOutputPath string
	ReportOutputPath    string
	FetchProxy          string
	ProbeURL            string
	ValidStatuses       map[int]struct{}
	FetchTimeout        time.Duration
	ValidateTimeout     time.Duration
	ValidateConcurrency int
}

func Default() Config {
	statuses, _ := ParseStatuses(DefaultValidStatuses)
	return Config{
		SourceFile:          DefaultSourceFile,
		OutputDir:           DefaultOutputDir,
		RawOutputPath:       DefaultRawOutputFile,
		UniqueOutputPath:    DefaultUniqueOutputFile,
		ValidatedOutputPath: DefaultValidatedOutputFile,
		ReportOutputPath:    DefaultReportOutputFile,
		ProbeURL:            DefaultProbeURL,
		ValidStatuses:       statuses,
		FetchTimeout:        DefaultFetchTimeout,
		ValidateTimeout:     DefaultValidateTimeout,
		ValidateConcurrency: DefaultValidateConcurrency,
	}
}

func Parse(args []string) (Config, error) {
	defaults := Default()
	fs := flag.NewFlagSet("sky-socks5", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var (
		sourceFile          string
		outputDir           string
		rawOutput           string
		uniqueOutput        string
		validatedOutput     string
		reportOutput        string
		fetchProxy          string
		probeURL            string
		validStatuses       string
		fetchTimeout        time.Duration
		validateTimeout     time.Duration
		validateConcurrency int
	)

	fs.StringVar(&sourceFile, "source-file", defaults.SourceFile, "Path to newline-delimited source URL file")
	fs.StringVar(&outputDir, "output-dir", defaults.OutputDir, "Directory used for generated artifacts")
	fs.StringVar(&rawOutput, "raw-output", DefaultRawOutputFile, "Output file for parsed proxies before deduplication")
	fs.StringVar(&uniqueOutput, "unique-output", DefaultUniqueOutputFile, "Output file for deduplicated proxies")
	fs.StringVar(&validatedOutput, "validated-output", DefaultValidatedOutputFile, "Output file for reachable proxies")
	fs.StringVar(&reportOutput, "report-output", DefaultReportOutputFile, "Output file for JSON run report")
	fs.StringVar(&reportOutput, "meta-output", DefaultReportOutputFile, "Alias for -report-output")
	fs.StringVar(&fetchProxy, "fetch-proxy", "", "Optional proxy used when fetching source lists")
	fs.StringVar(&fetchProxy, "proxy", "", "Backward-compatible alias for -fetch-proxy")
	fs.StringVar(&probeURL, "probe-url", defaults.ProbeURL, "HTTP or HTTPS URL used to validate proxies")
	fs.StringVar(&validStatuses, "valid-statuses", DefaultValidStatuses, "Comma-separated HTTP statuses accepted during validation")
	fs.DurationVar(&fetchTimeout, "fetch-timeout", defaults.FetchTimeout, "Timeout for fetching each source list")
	fs.DurationVar(&validateTimeout, "validate-timeout", defaults.ValidateTimeout, "Timeout for validating each proxy")
	fs.IntVar(&validateConcurrency, "validate-concurrency", defaults.ValidateConcurrency, "Maximum concurrent proxy validation workers")
	fs.IntVar(&validateConcurrency, "concurrency", defaults.ValidateConcurrency, "Alias for -validate-concurrency")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	if extra := fs.Args(); len(extra) > 0 {
		return Config{}, fmt.Errorf("unexpected positional arguments: %s", strings.Join(extra, " "))
	}

	statusSet, err := ParseStatuses(validStatuses)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		SourceFile:          cleanRequiredPath(sourceFile),
		OutputDir:           cleanRequiredPath(outputDir),
		FetchProxy:          strings.TrimSpace(fetchProxy),
		ProbeURL:            strings.TrimSpace(probeURL),
		ValidStatuses:       statusSet,
		FetchTimeout:        fetchTimeout,
		ValidateTimeout:     validateTimeout,
		ValidateConcurrency: validateConcurrency,
	}
	if cfg.SourceFile == "" {
		return Config{}, errors.New("source file path is required")
	}
	if cfg.OutputDir == "" {
		return Config{}, errors.New("output directory is required")
	}
	if cfg.FetchTimeout <= 0 {
		return Config{}, errors.New("fetch timeout must be greater than zero")
	}
	if cfg.ValidateTimeout <= 0 {
		return Config{}, errors.New("validate timeout must be greater than zero")
	}
	if cfg.ValidateConcurrency <= 0 {
		return Config{}, errors.New("validate concurrency must be greater than zero")
	}
	if err := validateHTTPURL(cfg.ProbeURL, "probe URL"); err != nil {
		return Config{}, err
	}
	if cfg.FetchProxy != "" {
		if err := validateProxyURL(cfg.FetchProxy); err != nil {
			return Config{}, err
		}
	}

	if cfg.RawOutputPath, err = resolveOutputPath(cfg.OutputDir, rawOutput); err != nil {
		return Config{}, fmt.Errorf("raw output: %w", err)
	}
	if cfg.UniqueOutputPath, err = resolveOutputPath(cfg.OutputDir, uniqueOutput); err != nil {
		return Config{}, fmt.Errorf("unique output: %w", err)
	}
	if cfg.ValidatedOutputPath, err = resolveOutputPath(cfg.OutputDir, validatedOutput); err != nil {
		return Config{}, fmt.Errorf("validated output: %w", err)
	}
	if cfg.ReportOutputPath, err = resolveOutputPath(cfg.OutputDir, reportOutput); err != nil {
		return Config{}, fmt.Errorf("report output: %w", err)
	}
	if err := validateDistinctPaths(cfg.RawOutputPath, cfg.UniqueOutputPath, cfg.ValidatedOutputPath, cfg.ReportOutputPath); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Usage() string {
	return fmt.Sprintf(`Usage: sky-socks5 [flags]

Flags:
  -source-file string
        Path to newline-delimited source URL file (default %q)
  -output-dir string
        Directory used for generated artifacts (default %q)
  -raw-output string
        Output file for parsed proxies before deduplication (default %q)
  -unique-output string
        Output file for deduplicated proxies (default %q)
  -validated-output string
        Output file for reachable proxies (default %q)
  -report-output string
        Output file for JSON run report (default %q)
  -fetch-proxy string
        Optional proxy used when fetching source lists; -proxy is accepted as an alias
  -probe-url string
        HTTP or HTTPS URL used to validate proxies (default %q)
  -valid-statuses string
        Comma-separated HTTP statuses accepted during validation (default %q)
  -fetch-timeout duration
        Timeout for fetching each source list (default %s)
  -validate-timeout duration
        Timeout for validating each proxy (default %s)
  -validate-concurrency int
        Maximum concurrent proxy validation workers (default %d)
`, DefaultSourceFile, DefaultOutputDir, DefaultRawOutputFile, DefaultUniqueOutputFile, DefaultValidatedOutputFile, DefaultReportOutputFile, DefaultProbeURL, DefaultValidStatuses, DefaultFetchTimeout, DefaultValidateTimeout, DefaultValidateConcurrency)
}

func ParseStatuses(raw string) (map[int]struct{}, error) {
	statuses := make(map[int]struct{})
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		status, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid HTTP status %q: %w", part, err)
		}
		if status < 100 || status > 599 {
			return nil, fmt.Errorf("HTTP status %d is out of range", status)
		}
		statuses[status] = struct{}{}
	}
	if len(statuses) == 0 {
		return nil, errors.New("at least one valid status is required")
	}
	return statuses, nil
}

func StatusesString(statuses map[int]struct{}) string {
	values := make([]int, 0, len(statuses))
	for status := range statuses {
		values = append(values, status)
	}
	sort.Ints(values)
	parts := make([]string, 0, len(values))
	for _, status := range values {
		parts = append(parts, strconv.Itoa(status))
	}
	return strings.Join(parts, ",")
}

func cleanRequiredPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	return filepath.Clean(path)
}

func resolveOutputPath(outputDir, rawPath string) (string, error) {
	cleaned := strings.TrimSpace(rawPath)
	if cleaned == "" {
		return "", errors.New("path is required")
	}
	if filepath.IsAbs(cleaned) {
		cleaned = filepath.Clean(cleaned)
	} else {
		cleaned = filepath.Join(outputDir, cleaned)
	}
	if cleaned == "." || cleaned == string(filepath.Separator) {
		return "", fmt.Errorf("path %q must include a file name", rawPath)
	}
	return filepath.Clean(cleaned), nil
}

func validateHTTPURL(raw, label string) error {
	if raw == "" {
		return fmt.Errorf("%s is required", label)
	}
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("invalid %s %q: %w", label, raw, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("%s must use http or https, got %q", label, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("%s %q is missing a host", label, raw)
	}
	return nil
}

func validateProxyURL(raw string) error {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("invalid fetch proxy %q: %w", raw, err)
	}
	switch parsed.Scheme {
	case "http", "https", "socks5":
	default:
		return fmt.Errorf("fetch proxy must use http, https, or socks5, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("fetch proxy %q is missing a host", raw)
	}
	return nil
}

func validateDistinctPaths(paths ...string) error {
	seen := make(map[string]string, len(paths))
	for _, path := range paths {
		key := strings.ToLower(filepath.Clean(path))
		if previous, exists := seen[key]; exists {
			return fmt.Errorf("output paths must be distinct: %s and %s", previous, path)
		}
		seen[key] = path
	}
	return nil
}
