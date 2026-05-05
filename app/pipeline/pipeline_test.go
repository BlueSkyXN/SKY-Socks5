package pipeline

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/BlueSkyXN/SKY-Socks5/app/config"
	"github.com/BlueSkyXN/SKY-Socks5/app/fetch"
	"github.com/BlueSkyXN/SKY-Socks5/app/report"
	"github.com/BlueSkyXN/SKY-Socks5/app/validate"
)

func TestRunWritesExpectedArtifacts(t *testing.T) {
	tmp := t.TempDir()
	statuses, _ := config.ParseStatuses("200")
	cfg := config.Default()
	cfg.OutputDir = tmp
	cfg.SourceFile = "sources.txt"
	cfg.RawOutputPath = filepath.Join(tmp, "raw.txt")
	cfg.UniqueOutputPath = filepath.Join(tmp, "unique.txt")
	cfg.ValidatedOutputPath = filepath.Join(tmp, "validated.txt")
	cfg.ValidatedCSVPath = filepath.Join(tmp, "validated.csv")
	cfg.ReportOutputPath = filepath.Join(tmp, "report.json")
	cfg.ValidStatuses = statuses
	cfg.ValidateConcurrency = 2

	writes := make(map[string][]string)
	var csvRows [][]string
	var writtenReport report.Metadata
	result, err := Run(context.Background(), cfg, Dependencies{
		LoadSources: func(path string) ([]string, error) {
			return []string{"https://source-one.example", "https://source-two.example"}, nil
		},
		FetchAll: func(ctx context.Context, urls []string, opts fetch.Options) []fetch.Result {
			return []fetch.Result{
				{URL: urls[0], StatusCode: 200, Lines: []string{"2.2.2.2:1080", "1.1.1.1:1080", "bad"}},
				{URL: urls[1], StatusCode: 200, Lines: []string{"1.1.1.1:1080", "3.3.3.3:1080"}},
			}
		},
		Validate: func(ctx context.Context, addresses []string, opts validate.Options) ([]validate.Result, error) {
			return []validate.Result{
				{Address: "1.1.1.1:1080", Reachable: true, StatusCode: 200, ExitIP: "198.51.100.10", ExitCountry: "US", CloudflareColo: "SJC"},
				{Address: "2.2.2.2:1080", Reachable: false, StatusCode: 500},
				{Address: "3.3.3.3:1080", Reachable: true, StatusCode: 200, ExitIP: "203.0.113.10", ExitCountry: "FR", CloudflareColo: "CDG"},
			}, nil
		},
		WriteLines: func(path string, lines []string, dedupe bool) error {
			writes[filepath.Base(path)] = append([]string(nil), lines...)
			return nil
		},
		WriteCSV: func(path string, rows [][]string) error {
			csvRows = rows
			return nil
		},
		WriteJSON: func(path string, payload any) error {
			data, err := json.Marshal(payload)
			if err != nil {
				return err
			}
			return json.Unmarshal(data, &writtenReport)
		},
		Now: fixedClock(),
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !reflect.DeepEqual(result.Unique, []string{"1.1.1.1:1080", "2.2.2.2:1080", "3.3.3.3:1080"}) {
		t.Fatalf("Unique = %#v", result.Unique)
	}
	if !reflect.DeepEqual(writes["raw.txt"], []string{"1.1.1.1:1080", "1.1.1.1:1080", "2.2.2.2:1080", "3.3.3.3:1080"}) {
		t.Fatalf("raw writes = %#v", writes["raw.txt"])
	}
	if !reflect.DeepEqual(writes["validated.txt"], []string{"1.1.1.1:1080", "3.3.3.3:1080"}) {
		t.Fatalf("validated writes = %#v", writes["validated.txt"])
	}
	if writtenReport.Totals.ParseErrors != 1 || writtenReport.Totals.UniqueProxies != 3 || writtenReport.Totals.ValidProxies != 2 {
		t.Fatalf("report totals = %#v", writtenReport.Totals)
	}
	if writtenReport.Outputs.ValidatedCSVPath != cfg.ValidatedCSVPath {
		t.Fatalf("ValidatedCSVPath = %q", writtenReport.Outputs.ValidatedCSVPath)
	}
	if len(csvRows) != 3 {
		t.Fatalf("csv rows = %#v", csvRows)
	}
	if csvRows[1][0] != "1.1.1.1:1080" || csvRows[1][6] != "198.51.100.10" || csvRows[1][7] != "US" {
		t.Fatalf("first csv data row = %#v", csvRows[1])
	}
}

func fixedClock() func() time.Time {
	ticks := []time.Time{
		time.Date(2026, 5, 4, 1, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 4, 1, 0, 2, 0, time.UTC),
	}
	i := 0
	return func() time.Time {
		if i >= len(ticks) {
			return ticks[len(ticks)-1]
		}
		value := ticks[i]
		i++
		return value
	}
}
