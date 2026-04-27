package report

import (
	"testing"
	"time"
)

func TestMetadataFinalizeSortsAndComputesTotals(t *testing.T) {
	startedAt := time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC)
	finishedAt := startedAt.Add(3 * time.Second)

	meta := Metadata{
		Totals: Totals{
			UniqueProxies: 5,
			ValidProxies:  2,
		},
		Sources: []Source{
			{URL: "https://b.example"},
			{URL: "https://a.example"},
		},
		Errors: []string{"zeta", "alpha"},
	}

	meta.Finalize(startedAt, finishedAt)

	if got, want := meta.Duration, "3s"; got != want {
		t.Fatalf("Duration = %q, want %q", got, want)
	}
	if got, want := meta.Totals.InvalidProxies, 3; got != want {
		t.Fatalf("InvalidProxies = %d, want %d", got, want)
	}
	if got, want := meta.Sources[0].URL, "https://a.example"; got != want {
		t.Fatalf("Sources[0].URL = %q, want %q", got, want)
	}
	if got, want := meta.Errors[0], "alpha"; got != want {
		t.Fatalf("Errors[0] = %q, want %q", got, want)
	}
	if !meta.GeneratedAt.Equal(finishedAt) {
		t.Fatalf("GeneratedAt = %v, want %v", meta.GeneratedAt, finishedAt)
	}
}
