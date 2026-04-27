package pipeline

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/BlueSkyXN/SKY-Socks5/internal/config"
	"github.com/BlueSkyXN/SKY-Socks5/internal/report"
	"github.com/BlueSkyXN/SKY-Socks5/internal/validate"
)

const (
	defaultSourceFile         = config.DefaultSourceFile
	defaultRawOutputPath      = config.DefaultRawOutputFile
	defaultUniqueOutputPath   = config.DefaultUniqueOutputFile
	defaultValidatedOutput    = config.DefaultValidatedOutputFile
	defaultMetadataOutputPath = config.DefaultMetaOutputFile
	defaultFetchTimeout       = config.DefaultFetchTimeout
	defaultProbeURL           = config.DefaultProbeURL
	defaultConcurrency        = config.DefaultValidateConcurrency
	defaultValidateTimeout    = config.DefaultValidateTimeout
)

type Options struct {
	RawOutputPath       string
	SourceFile          string
	UniqueOutputPath    string
	ValidatedOutputPath string
	MetadataOutputPath  string
	FetchTimeout        time.Duration
	ValidateTimeout     time.Duration
	ValidateConcurrency int
	ProbeURL            string
}

type SourceResult struct {
	URL     string
	Entries []string
	Error   error
}

type Dependencies struct {
	LoadSources func(path string) ([]string, error)
	FetchAll    func(ctx context.Context, urls []string, timeout time.Duration) ([]SourceResult, error)
	Normalize   func(raw string) (string, error)
	Validate    func(ctx context.Context, addresses []string, opts validate.Options) ([]validate.Result, error)
	WriteRaw    func(path string, addresses []string) error
	WriteText   func(path string, addresses []string) error
	WriteJSON   func(path string, payload any) error
	Now         func() time.Time
}

type Result struct {
	Raw      []string
	Unique   []string
	Valid    []string
	Metadata report.Metadata
}

func Run(ctx context.Context, opts Options, deps Dependencies) (result Result, err error) {
	opts = opts.withDefaults()
	deps = deps.withDefaults()
	if err = deps.validate(); err != nil {
		return result, err
	}

	startedAt := deps.Now()
	meta := report.Metadata{
		ProbeURL:          opts.ProbeURL,
		ValidateWorkers:   opts.ValidateConcurrency,
		FetchTimeout:      opts.FetchTimeout.String(),
		ValidationTimeout: opts.ValidateTimeout.String(),
		Outputs: report.Outputs{
			RawPath:       opts.RawOutputPath,
			UniquePath:    opts.UniqueOutputPath,
			ValidatedPath: opts.ValidatedOutputPath,
			MetadataPath:  opts.MetadataOutputPath,
		},
	}

	finish := func() {
		if result.Metadata.GeneratedAt.IsZero() {
			meta.Finalize(startedAt, deps.Now())
			result.Metadata = meta
		}
	}
	defer finish()

	sourceURLs, loadErr := deps.LoadSources(opts.SourceFile)
	if loadErr != nil {
		err = fmt.Errorf("load sources: %w", loadErr)
		return result, err
	}
	meta.Totals.SourceURLs = len(sourceURLs)

	sourceResults, fetchErr := deps.FetchAll(ctx, sourceURLs, opts.FetchTimeout)
	if fetchErr != nil {
		meta.Errors = append(meta.Errors, fetchErr.Error())
	}

	seen := make(map[string]struct{})
	parsed := make([]string, 0)
	unique := make([]string, 0)
	meta.Sources = make([]report.Source, 0, len(sourceResults))

	for _, sourceResult := range sourceResults {
		sourceMeta := report.Source{URL: sourceResult.URL}
		if sourceResult.Error != nil {
			sourceMeta.Error = sourceResult.Error.Error()
			meta.Totals.FetchErrors++
		}

		for _, entry := range sourceResult.Entries {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}

			sourceMeta.Candidates++
			meta.Totals.RawCandidates++

			normalized, normalizeErr := deps.Normalize(entry)
			if normalizeErr != nil {
				meta.Totals.ParseErrors++
				continue
			}

			meta.Totals.ParsedProxies++
			parsed = append(parsed, normalized)
			sourceMeta.Accepted++
			if _, ok := seen[normalized]; ok {
				continue
			}
			seen[normalized] = struct{}{}
			unique = append(unique, normalized)
		}

		meta.Sources = append(meta.Sources, sourceMeta)
	}

	sort.Strings(parsed)
	result.Raw = parsed
	sort.Strings(unique)
	result.Unique = unique
	meta.Totals.UniqueProxies = len(unique)

	validationResults, validateErr := deps.Validate(ctx, unique, validate.Options{
		ProbeURL:    opts.ProbeURL,
		Timeout:     opts.ValidateTimeout,
		Concurrency: opts.ValidateConcurrency,
	})
	if validateErr != nil {
		err = fmt.Errorf("validate proxies: %w", validateErr)
		return result, err
	}

	valid := make([]string, 0, len(validationResults))
	for _, validationResult := range validationResults {
		if validationResult.Reachable {
			valid = append(valid, validationResult.Address)
		}
	}
	sort.Strings(valid)

	result.Valid = valid
	meta.Totals.ValidProxies = len(valid)
	meta.Totals.InvalidProxies = meta.Totals.UniqueProxies - meta.Totals.ValidProxies

	if writeErr := deps.WriteRaw(opts.RawOutputPath, parsed); writeErr != nil {
		err = fmt.Errorf("write raw output: %w", writeErr)
		return result, err
	}
	if writeErr := deps.WriteText(opts.UniqueOutputPath, unique); writeErr != nil {
		err = fmt.Errorf("write unique output: %w", writeErr)
		return result, err
	}
	if writeErr := deps.WriteText(opts.ValidatedOutputPath, valid); writeErr != nil {
		err = fmt.Errorf("write validated output: %w", writeErr)
		return result, err
	}

	finish()
	if writeErr := deps.WriteJSON(opts.MetadataOutputPath, result.Metadata); writeErr != nil {
		err = fmt.Errorf("write metadata output: %w", writeErr)
		return result, err
	}

	return result, nil
}

func (o Options) withDefaults() Options {
	if o.SourceFile == "" {
		o.SourceFile = defaultSourceFile
	}
	if o.RawOutputPath == "" {
		o.RawOutputPath = defaultRawOutputPath
	}
	if o.UniqueOutputPath == "" {
		o.UniqueOutputPath = defaultUniqueOutputPath
	}
	if o.ValidatedOutputPath == "" {
		o.ValidatedOutputPath = defaultValidatedOutput
	}
	if o.MetadataOutputPath == "" {
		o.MetadataOutputPath = defaultMetadataOutputPath
	}
	if o.FetchTimeout <= 0 {
		o.FetchTimeout = defaultFetchTimeout
	}
	if o.ValidateTimeout <= 0 {
		o.ValidateTimeout = defaultValidateTimeout
	}
	if o.ValidateConcurrency <= 0 {
		o.ValidateConcurrency = defaultConcurrency
	}
	if o.ProbeURL == "" {
		o.ProbeURL = defaultProbeURL
	}
	return o
}

func (d Dependencies) withDefaults() Dependencies {
	if d.Validate == nil {
		d.Validate = validate.Check
	}
	if d.Now == nil {
		d.Now = time.Now
	}
	return d
}

func (d Dependencies) validate() error {
	switch {
	case d.LoadSources == nil:
		return fmt.Errorf("pipeline dependency LoadSources is required")
	case d.FetchAll == nil:
		return fmt.Errorf("pipeline dependency FetchAll is required")
	case d.Normalize == nil:
		return fmt.Errorf("pipeline dependency Normalize is required")
	case d.WriteRaw == nil:
		return fmt.Errorf("pipeline dependency WriteRaw is required")
	case d.WriteText == nil:
		return fmt.Errorf("pipeline dependency WriteText is required")
	case d.WriteJSON == nil:
		return fmt.Errorf("pipeline dependency WriteJSON is required")
	}
	return nil
}
