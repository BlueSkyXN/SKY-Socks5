package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultSourceFile          = "urls.txt"
	DefaultOutputDir           = "."
	DefaultRawOutputFile       = "raw_proxies.txt"
	DefaultUniqueOutputFile    = "unique_proxies.txt"
	DefaultValidatedOutputFile = "validated_proxies.txt"
	DefaultMetaOutputFile      = "proxy_report.json"
	DefaultProbeURL            = "https://one.one.one.one"
	DefaultFetchTimeout        = 15 * time.Second
	DefaultValidateTimeout     = 10 * time.Second
	DefaultValidateConcurrency = 10
)

// Config contains the runtime settings shared across the pipeline.
type Config struct {
	SourceFile          string
	OutputDir           string
	RawOutputPath       string
	UniqueOutputPath    string
	ValidatedOutputPath string
	MetaOutputPath      string
	ProbeURL            string
	FetchTimeout        time.Duration
	ValidateTimeout     time.Duration
	ValidateConcurrency int
}

// Default returns the default runtime configuration.
func Default() Config {
	return Config{
		SourceFile:          DefaultSourceFile,
		OutputDir:           DefaultOutputDir,
		RawOutputPath:       DefaultRawOutputFile,
		UniqueOutputPath:    DefaultUniqueOutputFile,
		ValidatedOutputPath: DefaultValidatedOutputFile,
		MetaOutputPath:      DefaultMetaOutputFile,
		ProbeURL:            DefaultProbeURL,
		FetchTimeout:        DefaultFetchTimeout,
		ValidateTimeout:     DefaultValidateTimeout,
		ValidateConcurrency: DefaultValidateConcurrency,
	}
}

// Load parses CLI arguments into a validated Config without exiting the process.
func Load(args []string) (Config, error) {
	defaults := Default()

	fs := flag.NewFlagSet("sky-socks5", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var (
		sourceFile          string
		outputDir           string
		rawOutput           string
		uniqueOutput        string
		validatedOutput     string
		metaOutput          string
		probeURL            string
		fetchTimeout        time.Duration
		validateTimeout     time.Duration
		validateConcurrency int
	)

	fs.StringVar(&sourceFile, "source-file", defaults.SourceFile, "Path to the source URL file")
	fs.StringVar(&outputDir, "output-dir", defaults.OutputDir, "Directory used for generated artifacts")
	fs.StringVar(&rawOutput, "raw-output", DefaultRawOutputFile, "Filename or path for aggregated raw proxies")
	fs.StringVar(&uniqueOutput, "unique-output", DefaultUniqueOutputFile, "Filename or path for unique proxies")
	fs.StringVar(&validatedOutput, "validated-output", DefaultValidatedOutputFile, "Filename or path for validated proxies")
	fs.StringVar(&metaOutput, "meta-output", DefaultMetaOutputFile, "Filename or path for metadata JSON")
	fs.StringVar(&metaOutput, "metadata-output", DefaultMetaOutputFile, "Filename or path for metadata JSON")
	fs.StringVar(&probeURL, "probe-url", defaults.ProbeURL, "Probe URL used during validation")
	fs.DurationVar(&fetchTimeout, "fetch-timeout", defaults.FetchTimeout, "HTTP timeout for source fetching")
	fs.DurationVar(&validateTimeout, "validate-timeout", defaults.ValidateTimeout, "Timeout for validating a proxy")
	fs.IntVar(&validateConcurrency, "validate-concurrency", defaults.ValidateConcurrency, "Maximum number of concurrent proxy validations")
	fs.IntVar(&validateConcurrency, "concurrency", defaults.ValidateConcurrency, "Maximum number of concurrent proxy validations")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	if extra := fs.Args(); len(extra) > 0 {
		return Config{}, fmt.Errorf("unexpected positional arguments: %s", strings.Join(extra, " "))
	}

	sourceFile = strings.TrimSpace(sourceFile)
	if sourceFile == "" {
		return Config{}, errors.New("source file path is required")
	}
	outputDir = strings.TrimSpace(outputDir)
	if outputDir == "" {
		return Config{}, errors.New("output directory is required")
	}

	cfg := Config{
		SourceFile:          filepath.Clean(sourceFile),
		OutputDir:           filepath.Clean(outputDir),
		ProbeURL:            strings.TrimSpace(probeURL),
		FetchTimeout:        fetchTimeout,
		ValidateTimeout:     validateTimeout,
		ValidateConcurrency: validateConcurrency,
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
	if err := validateProbeURL(cfg.ProbeURL); err != nil {
		return Config{}, err
	}

	var err error
	if cfg.RawOutputPath, err = resolveOutputPath(cfg.OutputDir, rawOutput); err != nil {
		return Config{}, fmt.Errorf("raw output: %w", err)
	}
	if cfg.UniqueOutputPath, err = resolveOutputPath(cfg.OutputDir, uniqueOutput); err != nil {
		return Config{}, fmt.Errorf("unique output: %w", err)
	}
	if cfg.ValidatedOutputPath, err = resolveOutputPath(cfg.OutputDir, validatedOutput); err != nil {
		return Config{}, fmt.Errorf("validated output: %w", err)
	}
	if cfg.MetaOutputPath, err = resolveOutputPath(cfg.OutputDir, metaOutput); err != nil {
		return Config{}, fmt.Errorf("meta output: %w", err)
	}

	if err := validateDistinctPaths(map[string]string{
		"raw output":       cfg.RawOutputPath,
		"unique output":    cfg.UniqueOutputPath,
		"validated output": cfg.ValidatedOutputPath,
		"meta output":      cfg.MetaOutputPath,
	}); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func validateProbeURL(raw string) error {
	if raw == "" {
		return errors.New("probe URL is required")
	}

	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("invalid probe URL %q: %w", raw, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("probe URL must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("probe URL %q is missing a host", raw)
	}

	return nil
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

func validateDistinctPaths(namedPaths map[string]string) error {
	seen := make(map[string]string, len(namedPaths))
	for name, path := range namedPaths {
		key := strings.ToLower(filepath.Clean(path))
		if previous, exists := seen[key]; exists {
			return fmt.Errorf("%s and %s cannot point to the same file (%s)", previous, name, path)
		}
		seen[key] = name
	}
	return nil
}
