package proxy

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "host port", raw: "Example.COM:1080", want: "example.com:1080"},
		{name: "url", raw: "socks5://user:pass@127.0.0.1:1080", want: "127.0.0.1:1080"},
		{name: "first token", raw: "10.0.0.1:1080 extra", want: "10.0.0.1:1080"},
		{name: "ipv6", raw: "[2001:db8::1]:1080", want: "[2001:db8::1]:1080"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Normalize(tt.raw)
			if err != nil {
				t.Fatalf("Normalize() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("Normalize() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeRejectsBadInput(t *testing.T) {
	badInputs := []string{"", "# comment", "missing-port", "127.0.0.1:99999", "2001:db8::1:1080", "http://127.0.0.1:8080"}
	for _, raw := range badInputs {
		if _, err := Normalize(raw); err == nil {
			t.Fatalf("Normalize(%q) expected error", raw)
		}
	}
}

func TestExtractCandidatesHandlesWhitespaceAndCSVMetadata(t *testing.T) {
	got, errs := ExtractCandidates("socks5://1.1.1.1:1080,FR,Paris 2.2.2.2:1080")
	if len(errs) != 0 {
		t.Fatalf("ExtractCandidates() errors = %#v", errs)
	}
	if len(got) != 2 {
		t.Fatalf("ExtractCandidates() len = %d, want 2", len(got))
	}
	if got[0].Address != "1.1.1.1:1080" || got[0].SourceCountry != "FR" || got[0].SourceCity != "Paris" {
		t.Fatalf("first candidate = %#v", got[0])
	}
	if got[1].Address != "2.2.2.2:1080" {
		t.Fatalf("second candidate = %#v", got[1])
	}
}
