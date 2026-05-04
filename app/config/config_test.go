package config

import "testing"

func TestParseDefaultsAndAliases(t *testing.T) {
	cfg, err := Parse([]string{
		"-output-dir", "generated",
		"-proxy", "socks5://127.0.0.1:1080",
		"-valid-statuses", "200,204,403",
		"-concurrency", "12",
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if cfg.FetchProxy != "socks5://127.0.0.1:1080" {
		t.Fatalf("FetchProxy = %q", cfg.FetchProxy)
	}
	if cfg.RawOutputPath != "generated/raw_proxies.txt" {
		t.Fatalf("RawOutputPath = %q", cfg.RawOutputPath)
	}
	if cfg.ValidateConcurrency != 12 {
		t.Fatalf("ValidateConcurrency = %d", cfg.ValidateConcurrency)
	}
	if got := StatusesString(cfg.ValidStatuses); got != "200,204,403" {
		t.Fatalf("ValidStatuses = %q", got)
	}
}

func TestParseRejectsDuplicateOutputs(t *testing.T) {
	_, err := Parse([]string{"-raw-output", "same.txt", "-unique-output", "same.txt"})
	if err == nil {
		t.Fatal("Parse() expected duplicate output error")
	}
}

func TestParseStatusesRejectsInvalidValues(t *testing.T) {
	if _, err := ParseStatuses("200,nope"); err == nil {
		t.Fatal("ParseStatuses() expected invalid status error")
	}
	if _, err := ParseStatuses("99"); err == nil {
		t.Fatal("ParseStatuses() expected out-of-range error")
	}
}
