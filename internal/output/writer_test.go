package output

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BlueSkyXN/SKY-Socks5/internal/proxy"
)

func TestWriteTextCreatesDirectoriesAndSorts(t *testing.T) {
	baseDir := outputTestWorkspace(t)
	path := filepath.Join(baseDir, "nested", "unique.txt")

	err := WriteText(path, []proxy.Proxy{
		{Host: "2.2.2.2", Port: 1080},
		{Host: "1.1.1.1", Port: 1080},
		{Host: "2.2.2.2", Port: 1080},
	})
	if err != nil {
		t.Fatalf("WriteText() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	want := "1.1.1.1:1080\n2.2.2.2:1080\n"
	if string(data) != want {
		t.Fatalf("WriteText() wrote %q, want %q", string(data), want)
	}
}

func TestWriteJSONWritesIndentedPayload(t *testing.T) {
	baseDir := outputTestWorkspace(t)
	path := filepath.Join(baseDir, "meta.json")

	err := WriteJSON(path, map[string]any{
		"b": 2,
		"a": 1,
	})
	if err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	want := "{\n  \"a\": 1,\n  \"b\": 2\n}\n"
	if string(data) != want {
		t.Fatalf("WriteJSON() wrote %q, want %q", string(data), want)
	}
}

func TestWriteLinesSortsAndDeduplicates(t *testing.T) {
	baseDir := outputTestWorkspace(t)
	path := filepath.Join(baseDir, "lines.txt")

	if err := WriteLines(path, []string{"b", "a", "b", " "}); err != nil {
		t.Fatalf("WriteLines() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	if string(data) != "a\nb\n" {
		t.Fatalf("WriteLines() wrote %q", string(data))
	}
}

func outputTestWorkspace(t *testing.T) string {
	t.Helper()

	dir := filepath.Join("testdata", "generated")
	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("RemoveAll(%q) error = %v", dir, err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", dir, err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	return dir
}
