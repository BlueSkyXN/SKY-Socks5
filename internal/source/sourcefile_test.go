package source

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadFileDeduplicatesAndSorts(t *testing.T) {
	path := writeSourceTestFile(t, "sources.txt", `
# comment
https://example.com/b.txt

https://example.com/a.txt
https://example.com/b.txt   # duplicate
`)

	got, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	want := []string{
		"https://example.com/a.txt",
		"https://example.com/b.txt",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadFile() = %#v, want %#v", got, want)
	}
}

func TestLoadFileRejectsInvalidURL(t *testing.T) {
	path := writeSourceTestFile(t, "invalid.txt", "not-a-url\n")

	if _, err := LoadFile(path); err == nil {
		t.Fatal("LoadFile() error = nil, want non-nil")
	}
}

func TestLoadFileRejectsEmptySourceList(t *testing.T) {
	path := writeSourceTestFile(t, "empty.txt", "\n# comment only\n")

	if _, err := LoadFile(path); err == nil {
		t.Fatal("LoadFile() error = nil, want non-nil")
	}
}

func writeSourceTestFile(t *testing.T, name, content string) string {
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

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
	return path
}
