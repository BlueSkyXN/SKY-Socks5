package source

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadFileSkipsBlankAndCommentLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sources.txt")
	if err := os.WriteFile(path, []byte("\n# comment\nhttps://one.example/list.txt\n https://two.example/list.txt \n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	want := []string{"https://one.example/list.txt", "https://two.example/list.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadFile() = %#v, want %#v", got, want)
	}
}
