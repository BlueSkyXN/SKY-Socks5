package pipeline

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/BlueSkyXN/SKY-Socks5/internal/report"
	"github.com/BlueSkyXN/SKY-Socks5/internal/validate"
)

func TestRunNormalizesSortsAndWritesOutputs(t *testing.T) {
	writtenText := make(map[string][]string)
	var writtenJSON any

	nowCalls := 0
	result, err := Run(context.Background(), Options{
		SourceFile:          "sources.txt",
		RawOutputPath:       "raw.txt",
		UniqueOutputPath:    "unique.txt",
		ValidatedOutputPath: "valid.txt",
		MetadataOutputPath:  "report.json",
		ProbeURL:            "http://probe.example",
		FetchTimeout:        3 * time.Second,
		ValidateTimeout:     4 * time.Second,
		ValidateConcurrency: 3,
	}, Dependencies{
		LoadSources: func(path string) ([]string, error) {
			if got, want := path, "sources.txt"; got != want {
				t.Fatalf("LoadSources path = %q, want %q", got, want)
			}
			return []string{"https://a.example", "https://b.example"}, nil
		},
		FetchAll: func(ctx context.Context, urls []string, timeout time.Duration) ([]SourceResult, error) {
			return []SourceResult{
				{
					URL:     urls[0],
					Entries: []string{" 2.2.2.2:1080 ", "invalid-entry", "1.1.1.1:1080"},
				},
				{
					URL:     urls[1],
					Entries: []string{"1.1.1.1:1080", "", "3.3.3.3:1080"},
					Error:   fmt.Errorf("partial fetch failure"),
				},
			}, fmt.Errorf("one source failed")
		},
		Normalize: func(raw string) (string, error) {
			switch raw {
			case "1.1.1.1:1080", "2.2.2.2:1080", "3.3.3.3:1080":
				return raw, nil
			default:
				return "", fmt.Errorf("invalid proxy")
			}
		},
		Validate: func(ctx context.Context, addresses []string, opts validate.Options) ([]validate.Result, error) {
			if got, want := opts.Concurrency, 3; got != want {
				t.Fatalf("Validate concurrency = %d, want %d", got, want)
			}
			if got, want := opts.ProbeURL, "http://probe.example"; got != want {
				t.Fatalf("Validate probe URL = %q, want %q", got, want)
			}
			return []validate.Result{
				{Address: addresses[0], Reachable: false},
				{Address: addresses[1], Reachable: true},
				{Address: addresses[2], Reachable: true},
			}, nil
		},
		WriteText: func(path string, addresses []string) error {
			writtenText[path] = append([]string(nil), addresses...)
			return nil
		},
		WriteRaw: func(path string, addresses []string) error {
			writtenText[path] = append([]string(nil), addresses...)
			return nil
		},
		WriteJSON: func(path string, payload any) error {
			if got, want := path, "report.json"; got != want {
				t.Fatalf("WriteJSON path = %q, want %q", got, want)
			}
			writtenJSON = payload
			return nil
		},
		Now: func() time.Time {
			defer func() { nowCalls++ }()
			return time.Date(2024, time.January, 1, 10, 0, nowCalls, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if got, want := result.Unique, []string{"1.1.1.1:1080", "2.2.2.2:1080", "3.3.3.3:1080"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Unique = %v, want %v", got, want)
	}
	if got, want := result.Raw, []string{"1.1.1.1:1080", "1.1.1.1:1080", "2.2.2.2:1080", "3.3.3.3:1080"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Raw = %v, want %v", got, want)
	}
	if got, want := result.Valid, []string{"2.2.2.2:1080", "3.3.3.3:1080"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Valid = %v, want %v", got, want)
	}
	if got, want := writtenText["raw.txt"], []string{"1.1.1.1:1080", "1.1.1.1:1080", "2.2.2.2:1080", "3.3.3.3:1080"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("raw output = %v, want %v", got, want)
	}
	if got, want := writtenText["unique.txt"], []string{"1.1.1.1:1080", "2.2.2.2:1080", "3.3.3.3:1080"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unique output = %v, want %v", got, want)
	}
	if got, want := writtenText["valid.txt"], []string{"2.2.2.2:1080", "3.3.3.3:1080"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("validated output = %v, want %v", got, want)
	}

	meta, ok := writtenJSON.(report.Metadata)
	if !ok {
		t.Fatalf("written JSON payload type = %T, want report.Metadata", writtenJSON)
	}
	if got, want := meta.Totals.FetchErrors, 1; got != want {
		t.Fatalf("FetchErrors = %d, want %d", got, want)
	}
	if got, want := meta.Totals.RawCandidates, 5; got != want {
		t.Fatalf("RawCandidates = %d, want %d", got, want)
	}
	if got, want := meta.Totals.ParseErrors, 1; got != want {
		t.Fatalf("ParseErrors = %d, want %d", got, want)
	}
	if got, want := meta.Totals.ParsedProxies, 4; got != want {
		t.Fatalf("ParsedProxies = %d, want %d", got, want)
	}
	if got, want := meta.Totals.ValidProxies, 2; got != want {
		t.Fatalf("ValidProxies = %d, want %d", got, want)
	}
	if got, want := meta.Errors, []string{"one source failed"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Errors = %v, want %v", got, want)
	}
}

func TestRunRequiresDependencies(t *testing.T) {
	_, err := Run(context.Background(), Options{}, Dependencies{})
	if err == nil {
		t.Fatal("Run error = nil, want non-nil")
	}
}

func TestParseArgsUsesExpectedDefaults(t *testing.T) {
	opts, err := ParseArgs(nil)
	if err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}

	if got, want := opts.SourceFile, defaultSourceFile; got != want {
		t.Fatalf("SourceFile = %q, want %q", got, want)
	}
	if got, want := opts.RawOutputPath, defaultRawOutputPath; got != want {
		t.Fatalf("RawOutputPath = %q, want %q", got, want)
	}
	if got, want := opts.UniqueOutputPath, defaultUniqueOutputPath; got != want {
		t.Fatalf("UniqueOutputPath = %q, want %q", got, want)
	}
	if got, want := opts.ValidatedOutputPath, defaultValidatedOutput; got != want {
		t.Fatalf("ValidatedOutputPath = %q, want %q", got, want)
	}
	if got, want := opts.MetadataOutputPath, defaultMetadataOutputPath; got != want {
		t.Fatalf("MetadataOutputPath = %q, want %q", got, want)
	}
	if got, want := opts.ValidateConcurrency, defaultConcurrency; got != want {
		t.Fatalf("ValidateConcurrency = %d, want %d", got, want)
	}
}
