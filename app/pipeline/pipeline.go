package pipeline

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/BlueSkyXN/SKY-Socks5/app/config"
	"github.com/BlueSkyXN/SKY-Socks5/app/fetch"
	"github.com/BlueSkyXN/SKY-Socks5/app/output"
	"github.com/BlueSkyXN/SKY-Socks5/app/proxy"
	"github.com/BlueSkyXN/SKY-Socks5/app/report"
	"github.com/BlueSkyXN/SKY-Socks5/app/source"
	"github.com/BlueSkyXN/SKY-Socks5/app/validate"
)

type Dependencies struct {
	LoadSources func(path string) ([]string, error)
	FetchAll    func(ctx context.Context, urls []string, opts fetch.Options) []fetch.Result
	Validate    func(ctx context.Context, addresses []string, opts validate.Options) ([]validate.Result, error)
	WriteLines  func(path string, lines []string, dedupe bool) error
	WriteCSV    func(path string, rows [][]string) error
	WriteJSON   func(path string, payload any) error
	Now         func() time.Time
}

type Result struct {
	Raw          []string
	Unique       []string
	Validated    []string
	ValidatedCSV [][]string
	Metadata     report.Metadata
}

type candidateInfo struct {
	Address       string
	Host          string
	Port          int
	SourceURLs    []string
	SourceCountry string
	SourceCity    string
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
		"sources=%d raw=%d unique=%d valid=%d raw_output=%s unique_output=%s validated_output=%s validated_csv=%s report=%s\n",
		result.Metadata.Totals.SourceURLs,
		len(result.Raw),
		len(result.Unique),
		len(result.Validated),
		cfg.RawOutputPath,
		cfg.UniqueOutputPath,
		cfg.ValidatedOutputPath,
		cfg.ValidatedCSVPath,
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
			RawPath:          cfg.RawOutputPath,
			UniquePath:       cfg.UniqueOutputPath,
			ValidatedPath:    cfg.ValidatedOutputPath,
			ValidatedCSVPath: cfg.ValidatedCSVPath,
			ReportPath:       cfg.ReportOutputPath,
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
	candidateIndex := make(map[string]*candidateInfo)
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
			candidates, errs := proxy.ExtractCandidates(line)
			sourceMeta.Candidates += len(candidates) + len(errs)
			meta.Totals.RawCandidates += len(candidates) + len(errs)
			meta.Totals.ParseErrors += len(errs)
			for _, candidate := range candidates {
				meta.Totals.ParsedProxies++
				sourceMeta.Accepted++
				raw = append(raw, candidate.Address)
				trackCandidate(candidateIndex, fetched.URL, candidate)
				if _, ok := seen[candidate.Address]; ok {
					continue
				}
				seen[candidate.Address] = struct{}{}
				unique = append(unique, candidate.Address)
			}
			if len(candidates) > 0 || len(errs) > 0 {
				continue
			}
			meta.Totals.ParseErrors++
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
	validatedCSV := validatedCSVHeader()
	for _, validationResult := range validationResults {
		if validationResult.Reachable {
			validated = append(validated, validationResult.Address)
			validatedCSV = append(validatedCSV, validatedCSVRow(validationResult, candidateIndex[validationResult.Address], cfg.ProbeURL))
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
	if err := deps.WriteCSV(cfg.ValidatedCSVPath, validatedCSV); err != nil {
		return Result{}, fmt.Errorf("write validated CSV output: %w", err)
	}
	if err := deps.WriteJSON(cfg.ReportOutputPath, meta); err != nil {
		return Result{}, fmt.Errorf("write report output: %w", err)
	}

	return Result{Raw: raw, Unique: unique, Validated: validated, ValidatedCSV: validatedCSV, Metadata: meta}, nil
}

func DefaultDependencies() Dependencies {
	return Dependencies{
		LoadSources: source.LoadFile,
		FetchAll:    fetch.All,
		Validate:    validate.Check,
		WriteLines:  output.WriteLines,
		WriteCSV:    output.WriteCSV,
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
	if d.WriteCSV == nil {
		d.WriteCSV = defaults.WriteCSV
	}
	if d.WriteJSON == nil {
		d.WriteJSON = defaults.WriteJSON
	}
	if d.Now == nil {
		d.Now = defaults.Now
	}
	return d
}

func trackCandidate(index map[string]*candidateInfo, sourceURL string, candidate proxy.Candidate) {
	info, ok := index[candidate.Address]
	if !ok {
		info = &candidateInfo{
			Address:       candidate.Address,
			Host:          candidate.Host,
			Port:          candidate.Port,
			SourceCountry: candidate.SourceCountry,
			SourceCity:    candidate.SourceCity,
		}
		index[candidate.Address] = info
	}
	info.SourceURLs = appendUnique(info.SourceURLs, sourceURL)
	if info.SourceCountry == "" {
		info.SourceCountry = candidate.SourceCountry
	}
	if info.SourceCity == "" {
		info.SourceCity = candidate.SourceCity
	}
}

func appendUnique(values []string, value string) []string {
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func validatedCSVHeader() [][]string {
	return [][]string{{
		"proxy_address",
		"entry_host",
		"entry_port",
		"source_urls",
		"source_country",
		"source_city",
		"exit_ip",
		"exit_country",
		"cloudflare_colo",
		"cloudflare_http",
		"cloudflare_tls",
		"status_code",
		"duration_ms",
		"probe_url",
	}}
}

func validatedCSVRow(result validate.Result, info *candidateInfo, probeURL string) []string {
	entryHost := ""
	entryPort := ""
	sourceURLs := ""
	sourceCountry := ""
	sourceCity := ""
	if info != nil {
		entryHost = info.Host
		entryPort = fmt.Sprintf("%d", info.Port)
		sourceURLs = joinCSVList(info.SourceURLs)
		sourceCountry = info.SourceCountry
		sourceCity = info.SourceCity
	}
	return []string{
		result.Address,
		entryHost,
		entryPort,
		sourceURLs,
		sourceCountry,
		sourceCity,
		result.ExitIP,
		result.ExitCountry,
		result.CloudflareColo,
		result.CloudflareHTTP,
		result.CloudflareTLS,
		fmt.Sprintf("%d", result.StatusCode),
		fmt.Sprintf("%d", result.DurationMillis),
		probeURL,
	}
}

func joinCSVList(values []string) string {
	return join(values, " | ")
}

func join(values []string, sep string) string {
	if len(values) == 0 {
		return ""
	}
	out := values[0]
	for _, value := range values[1:] {
		out += sep + value
	}
	return out
}
