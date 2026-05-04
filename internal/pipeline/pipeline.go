package pipeline

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/BlueSkyXN/SKY-Socks5/internal/config"
	"github.com/BlueSkyXN/SKY-Socks5/internal/fetch"
	"github.com/BlueSkyXN/SKY-Socks5/internal/output"
	"github.com/BlueSkyXN/SKY-Socks5/internal/proxy"
	"github.com/BlueSkyXN/SKY-Socks5/internal/report"
	"github.com/BlueSkyXN/SKY-Socks5/internal/source"
	"github.com/BlueSkyXN/SKY-Socks5/internal/validate"
)

type Dependencies struct {
	LoadSources func(path string) ([]string, error)
	FetchAll    func(ctx context.Context, urls []string, opts fetch.Options) []fetch.Result
	Validate    func(ctx context.Context, addresses []string, opts validate.Options) ([]validate.Result, error)
	WriteLines  func(path string, lines []string, dedupe bool) error
	WriteJSON   func(path string, payload any) error
	Now         func() time.Time
}

type Result struct {
	Raw       []string
	Unique    []string
	Validated []string
	Metadata  report.Metadata
}

func RunCLI(ctx context.Context, args []string, stdout io.Writer) error {
	cfg, err := config.Parse(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, _ = io.WriteString(stdout, config.Usage())
			return nil
		}
		return err
	}
	result, err := Run(ctx, cfg, DefaultDependencies())
	if err != nil {
		return err
	}
	fmt.Fprintf(
		stdout,
		"sources=%d raw=%d unique=%d valid=%d raw_output=%s unique_output=%s validated_output=%s report=%s\n",
		result.Metadata.Totals.SourceURLs,
		len(result.Raw),
		len(result.Unique),
		len(result.Validated),
		cfg.RawOutputPath,
		cfg.UniqueOutputPath,
		cfg.ValidatedOutputPath,
		cfg.ReportOutputPath,
	)
	return nil
}

func Run(ctx context.Context, cfg config.Config, deps Dependencies) (Result, error) {
	deps = deps.withDefaults()
	startedAt := deps.Now()
	meta := report.Metadata{
		ProbeURL:          cfg.ProbeURL,
		FetchProxy:        cfg.FetchProxy,
		ValidStatuses:     config.StatusesString(cfg.ValidStatuses),
		FetchTimeout:      cfg.FetchTimeout.String(),
		ValidationTimeout: cfg.ValidateTimeout.String(),
		ValidateWorkers:   cfg.ValidateConcurrency,
		Outputs: report.Outputs{
			RawPath:       cfg.RawOutputPath,
			UniquePath:    cfg.UniqueOutputPath,
			ValidatedPath: cfg.ValidatedOutputPath,
			ReportPath:    cfg.ReportOutputPath,
		},
	}

	sourceURLs, err := deps.LoadSources(cfg.SourceFile)
	if err != nil {
		return Result{}, fmt.Errorf("load sources: %w", err)
	}
	meta.Totals.SourceURLs = len(sourceURLs)

	fetchResults := deps.FetchAll(ctx, sourceURLs, fetch.Options{
		Timeout: cfg.FetchTimeout,
		Proxy:   cfg.FetchProxy,
	})

	raw := make([]string, 0)
	unique := make([]string, 0)
	seen := make(map[string]struct{})
	for _, fetched := range fetchResults {
		sourceMeta := report.Source{
			URL:            fetched.URL,
			StatusCode:     fetched.StatusCode,
			Error:          fetched.Error,
			DurationMillis: fetched.DurationMillis,
		}
		if fetched.Error != "" {
			meta.Totals.FetchErrors++
		}
		for _, line := range fetched.Lines {
			sourceMeta.Candidates++
			meta.Totals.RawCandidates++
			normalized, err := proxy.Normalize(line)
			if err != nil {
				meta.Totals.ParseErrors++
				continue
			}
			meta.Totals.ParsedProxies++
			sourceMeta.Accepted++
			raw = append(raw, normalized)
			if _, ok := seen[normalized]; ok {
				continue
			}
			seen[normalized] = struct{}{}
			unique = append(unique, normalized)
		}
		meta.Sources = append(meta.Sources, sourceMeta)
	}
	proxy.SortStrings(raw)
	proxy.SortStrings(unique)
	meta.Totals.UniqueProxies = len(unique)

	validationResults, err := deps.Validate(ctx, unique, validate.Options{
		ProbeURL:      cfg.ProbeURL,
		Timeout:       cfg.ValidateTimeout,
		Concurrency:   cfg.ValidateConcurrency,
		ValidStatuses: cfg.ValidStatuses,
	})
	if err != nil {
		return Result{}, fmt.Errorf("validate proxies: %w", err)
	}
	validated := make([]string, 0, len(validationResults))
	for _, validationResult := range validationResults {
		if validationResult.Reachable {
			validated = append(validated, validationResult.Address)
		}
	}
	proxy.SortStrings(validated)
	meta.Totals.ValidProxies = len(validated)
	meta.Finalize(startedAt, deps.Now())

	if err := deps.WriteLines(cfg.RawOutputPath, raw, false); err != nil {
		return Result{}, fmt.Errorf("write raw output: %w", err)
	}
	if err := deps.WriteLines(cfg.UniqueOutputPath, unique, true); err != nil {
		return Result{}, fmt.Errorf("write unique output: %w", err)
	}
	if err := deps.WriteLines(cfg.ValidatedOutputPath, validated, true); err != nil {
		return Result{}, fmt.Errorf("write validated output: %w", err)
	}
	if err := deps.WriteJSON(cfg.ReportOutputPath, meta); err != nil {
		return Result{}, fmt.Errorf("write report output: %w", err)
	}

	return Result{Raw: raw, Unique: unique, Validated: validated, Metadata: meta}, nil
}

func DefaultDependencies() Dependencies {
	return Dependencies{
		LoadSources: source.LoadFile,
		FetchAll:    fetch.All,
		Validate:    validate.Check,
		WriteLines:  output.WriteLines,
		WriteJSON:   output.WriteJSON,
		Now:         time.Now,
	}
}

func (d Dependencies) withDefaults() Dependencies {
	defaults := DefaultDependencies()
	if d.LoadSources == nil {
		d.LoadSources = defaults.LoadSources
	}
	if d.FetchAll == nil {
		d.FetchAll = defaults.FetchAll
	}
	if d.Validate == nil {
		d.Validate = defaults.Validate
	}
	if d.WriteLines == nil {
		d.WriteLines = defaults.WriteLines
	}
	if d.WriteJSON == nil {
		d.WriteJSON = defaults.WriteJSON
	}
	if d.Now == nil {
		d.Now = defaults.Now
	}
	return d
}
