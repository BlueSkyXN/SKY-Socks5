package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func WriteLines(path string, lines []string, dedupe bool) error {
	normalized := make([]string, 0, len(lines))
	seen := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if dedupe {
			if _, ok := seen[line]; ok {
				continue
			}
			seen[line] = struct{}{}
		}
		normalized = append(normalized, line)
	}
	sort.Strings(normalized)
	body := ""
	if len(normalized) > 0 {
		body = strings.Join(normalized, "\n") + "\n"
	}
	return writeFile(path, []byte(body))
}

func WriteJSON(path string, payload any) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	data = append(data, '\n')
	return writeFile(path, data)
}

func writeFile(path string, data []byte) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("output path is required")
	}
	path = filepath.Clean(path)
	if path == "." || path == string(filepath.Separator) {
		return fmt.Errorf("output path %q must include a file name", path)
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create output directory %q: %w", dir, err)
		}
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write file %q: %w", path, err)
	}
	return nil
}
