package fetch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestFetchAllReturnsSortedResultsAndSourceErrors(t *testing.T) {
	successServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("2.2.2.2:1080\n\n1.1.1.1:1080 \n"))
	}))
	defer successServer.Close()

	failureServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer failureServer.Close()

	results, err := FetchAll(context.Background(), []string{
		failureServer.URL,
		successServer.URL,
	}, 5*time.Second)
	if err != nil {
		t.Fatalf("FetchAll() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}

	if results[0].URL > results[1].URL {
		t.Fatalf("results are not sorted by URL: %#v", results)
	}

	byURL := make(map[string]SourceResult, len(results))
	for _, result := range results {
		byURL[result.URL] = result
	}

	success := byURL[successServer.URL]
	if !success.Successful() {
		t.Fatalf("success result = %#v, want successful", success)
	}
	if !reflect.DeepEqual(success.RawLines, []string{"2.2.2.2:1080", "1.1.1.1:1080"}) {
		t.Fatalf("success.RawLines = %#v", success.RawLines)
	}

	failure := byURL[failureServer.URL]
	if failure.Err == nil {
		t.Fatalf("failure.Err = nil, want non-nil")
	}
	if failure.StatusCode != http.StatusInternalServerError {
		t.Fatalf("failure.StatusCode = %d, want %d", failure.StatusCode, http.StatusInternalServerError)
	}

	if CombinedError(results) == nil {
		t.Fatal("CombinedError(results) = nil, want non-nil")
	}
}
