package config

import (
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.SourceFile != DefaultSourceFile {
		t.Fatalf("SourceFile = %q, want %q", cfg.SourceFile, DefaultSourceFile)
	}
	if cfg.RawOutputPath != DefaultRawOutputFile {
		t.Fatalf("RawOutputPath = %q, want %q", cfg.RawOutputPath, DefaultRawOutputFile)
	}
	if cfg.UniqueOutputPath != DefaultUniqueOutputFile {
		t.Fatalf("UniqueOutputPath = %q, want %q", cfg.UniqueOutputPath, DefaultUniqueOutputFile)
	}
	if cfg.ValidatedOutputPath != DefaultValidatedOutputFile {
		t.Fatalf("ValidatedOutputPath = %q, want %q", cfg.ValidatedOutputPath, DefaultValidatedOutputFile)
	}
	if cfg.MetaOutputPath != DefaultMetaOutputFile {
		t.Fatalf("MetaOutputPath = %q, want %q", cfg.MetaOutputPath, DefaultMetaOutputFile)
	}
	if cfg.FetchTimeout != DefaultFetchTimeout {
		t.Fatalf("FetchTimeout = %v, want %v", cfg.FetchTimeout, DefaultFetchTimeout)
	}
	if cfg.ValidateTimeout != DefaultValidateTimeout {
		t.Fatalf("ValidateTimeout = %v, want %v", cfg.ValidateTimeout, DefaultValidateTimeout)
	}
	if cfg.ValidateConcurrency != DefaultValidateConcurrency {
		t.Fatalf("ValidateConcurrency = %d, want %d", cfg.ValidateConcurrency, DefaultValidateConcurrency)
	}
}

func TestLoadCustomArgs(t *testing.T) {
	cfg, err := Load([]string{
		"-source-file", "sources/custom.txt",
		"-output-dir", "artifacts",
		"-raw-output", "raw.txt",
		"-unique-output", "unique.txt",
		"-validated-output", "validated.txt",
		"-metadata-output", "meta/report.json",
		"-probe-url", "https://example.com/health",
		"-fetch-timeout", "15s",
		"-validate-timeout", "8s",
		"-concurrency", "64",
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.SourceFile != filepath.Clean("sources/custom.txt") {
		t.Fatalf("SourceFile = %q", cfg.SourceFile)
	}
	if cfg.RawOutputPath != filepath.Join("artifacts", "raw.txt") {
		t.Fatalf("RawOutputPath = %q", cfg.RawOutputPath)
	}
	if cfg.UniqueOutputPath != filepath.Join("artifacts", "unique.txt") {
		t.Fatalf("UniqueOutputPath = %q", cfg.UniqueOutputPath)
	}
	if cfg.ValidatedOutputPath != filepath.Join("artifacts", "validated.txt") {
		t.Fatalf("ValidatedOutputPath = %q", cfg.ValidatedOutputPath)
	}
	if cfg.MetaOutputPath != filepath.Join("artifacts", "meta", "report.json") {
		t.Fatalf("MetaOutputPath = %q", cfg.MetaOutputPath)
	}
	if cfg.FetchTimeout != 15*time.Second {
		t.Fatalf("FetchTimeout = %v", cfg.FetchTimeout)
	}
	if cfg.ValidateTimeout != 8*time.Second {
		t.Fatalf("ValidateTimeout = %v", cfg.ValidateTimeout)
	}
	if cfg.ValidateConcurrency != 64 {
		t.Fatalf("ValidateConcurrency = %d", cfg.ValidateConcurrency)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "invalid probe URL",
			args: []string{"-probe-url", "ftp://example.com"},
		},
		{
			name: "invalid concurrency",
			args: []string{"-validate-concurrency", "0"},
		},
		{
			name: "duplicate output path",
			args: []string{"-raw-output", "same.txt", "-unique-output", "same.txt"},
		},
		{
			name: "blank source file",
			args: []string{"-source-file", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Load(tt.args); err == nil {
				t.Fatalf("Load(%v) error = nil, want non-nil", tt.args)
			}
		})
	}
}
