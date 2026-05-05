package validate

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestCheckUsesConfiguredStatusPolicy(t *testing.T) {
	results, err := Check(context.Background(), []string{"one:1", "two:2"}, Options{
		ProbeURL:    "https://example.com",
		Timeout:     time.Second,
		Concurrency: 2,
		ValidStatuses: map[int]struct{}{
			200: {},
			204: {},
		},
		Probe: func(ctx context.Context, address string, opts Options) Result {
			if address == "one:1" {
				return Result{Address: address, StatusCode: 204, Reachable: true}
			}
			return Result{Address: address, StatusCode: 500, Error: "unexpected status: 500"}
		},
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	want := []Result{
		{Address: "one:1", Reachable: true, StatusCode: 204},
		{Address: "two:2", StatusCode: 500, Error: "unexpected status: 500"},
	}
	if !reflect.DeepEqual(results, want) {
		t.Fatalf("Check() = %#v, want %#v", results, want)
	}
}

func TestParseCloudflareTrace(t *testing.T) {
	trace := parseCloudflareTrace("ip=203.0.113.10\nloc=FR\ncolo=CDG\nhttp=http/2\ntls=TLSv1.3\nsni=plaintext\nkex=X25519\nwarp=off\ngateway=on\nrbi=off\n")
	if trace["ip"] != "203.0.113.10" || trace["loc"] != "FR" || trace["colo"] != "CDG" || trace["http"] != "http/2" || trace["tls"] != "TLSv1.3" || trace["sni"] != "plaintext" || trace["kex"] != "X25519" {
		t.Fatalf("trace = %#v", trace)
	}
	if flags := cloudflareTraceFlags(trace); !reflect.DeepEqual(flags, []string{"gateway=on"}) {
		t.Fatalf("cloudflare flags = %#v", flags)
	}
}
