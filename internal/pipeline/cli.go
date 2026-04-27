package pipeline

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/BlueSkyXN/SKY-Socks5/internal/config"
	"github.com/BlueSkyXN/SKY-Socks5/internal/fetch"
	"github.com/BlueSkyXN/SKY-Socks5/internal/output"
	"github.com/BlueSkyXN/SKY-Socks5/internal/proxy"
	"github.com/BlueSkyXN/SKY-Socks5/internal/source"
	"github.com/BlueSkyXN/SKY-Socks5/internal/validate"
)

func ParseArgs(args []string) (Options, error) {
	cfg, err := config.Load(args)
	if err != nil {
		return Options{}, fmt.Errorf("parse flags: %w", err)
	}
	return Options{
		SourceFile:          cfg.SourceFile,
		RawOutputPath:       cfg.RawOutputPath,
		UniqueOutputPath:    cfg.UniqueOutputPath,
		ValidatedOutputPath: cfg.ValidatedOutputPath,
		MetadataOutputPath:  cfg.MetaOutputPath,
		FetchTimeout:        cfg.FetchTimeout,
		ValidateTimeout:     cfg.ValidateTimeout,
		ValidateConcurrency: cfg.ValidateConcurrency,
		ProbeURL:            cfg.ProbeURL,
	}.withDefaults(), nil
}

func RunCLI(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	opts, err := ParseArgs(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, _ = io.WriteString(stdout, Usage())
			return nil
		}
		return err
	}

	result, err := Run(ctx, opts, DefaultDependencies())
	if err != nil {
		return err
	}

	fmt.Fprintf(
		stdout,
		"sources=%d raw=%d unique=%d valid=%d raw_output=%s unique_output=%s validated_output=%s metadata=%s\n",
		result.Metadata.Totals.SourceURLs,
		len(result.Raw),
		len(result.Unique),
		len(result.Valid),
		opts.RawOutputPath,
		opts.UniqueOutputPath,
		opts.ValidatedOutputPath,
		opts.MetadataOutputPath,
	)
	_ = stderr
	return nil
}

func Usage() string {
	defaults := config.Default()
	return fmt.Sprintf(`Usage: sky-socks5 [flags]

Flags:
  -source-file string
    	Path to the source URL file (default %q)
  -output-dir string
    	Directory used for generated artifacts (default %q)
  -raw-output string
    	Filename or path for aggregated raw proxies (default %q)
  -unique-output string
    	Filename or path for unique proxies (default %q)
  -validated-output string
    	Filename or path for validated proxies (default %q)
  -meta-output string
    	Filename or path for metadata JSON (default %q)
  -probe-url string
    	Probe URL used during validation (default %q)
  -fetch-timeout duration
    	HTTP timeout for source fetching (default %s)
  -validate-timeout duration
    	Timeout for validating a proxy (default %s)
  -validate-concurrency int
    	Maximum number of concurrent proxy validations (default %d)
`, defaults.SourceFile, defaults.OutputDir, config.DefaultRawOutputFile, config.DefaultUniqueOutputFile, config.DefaultValidatedOutputFile, config.DefaultMetaOutputFile, defaults.ProbeURL, defaults.FetchTimeout, defaults.ValidateTimeout, defaults.ValidateConcurrency)
}

func DefaultDependencies() Dependencies {
	return Dependencies{
		LoadSources: source.LoadFile,
		FetchAll:    fetchAllSources,
		Normalize:   proxy.Normalize,
		Validate:    validate.Check,
		WriteRaw:    writeTextLines,
		WriteText:   writeProxyText,
		WriteJSON:   output.WriteJSON,
		Now:         time.Now,
	}.withDefaults()
}

func writeTextLines(path string, addresses []string) error {
	lines := append([]string(nil), addresses...)
	sort.Strings(lines)

	if err := ensureParentDir(path); err != nil {
		return err
	}
	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func writeProxyText(path string, addresses []string) error {
	proxies := make([]proxy.Proxy, 0, len(addresses))
	for _, address := range addresses {
		parsed, err := proxy.Parse(address)
		if err != nil {
			return fmt.Errorf("parse proxy %q: %w", address, err)
		}
		proxies = append(proxies, parsed)
	}
	return output.WriteText(path, proxies)
}

func fetchAllSources(ctx context.Context, urls []string, timeout time.Duration) ([]SourceResult, error) {
	results, err := fetch.FetchAll(ctx, urls, timeout)
	adapted := make([]SourceResult, 0, len(results))
	for _, result := range results {
		var resultErr error
		if result.Err != nil {
			resultErr = result.Err
		} else if result.Error != "" {
			resultErr = errors.New(result.Error)
		}
		adapted = append(adapted, SourceResult{
			URL:     result.URL,
			Entries: result.RawLines,
			Error:   resultErr,
		})
	}
	return adapted, errors.Join(err, fetch.CombinedError(results))
}

func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}
