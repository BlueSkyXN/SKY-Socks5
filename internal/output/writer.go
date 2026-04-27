package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BlueSkyXN/SKY-Socks5/internal/proxy"
)

// WriteText writes a deterministic text list of proxies to disk.
func WriteText(path string, proxies []proxy.Proxy) error {
	lines, err := sortedUniqueAddresses(proxies)
	if err != nil {
		return err
	}

	return WriteLines(path, lines)
}

// WriteJSON writes indented JSON with a trailing newline.
func WriteJSON(path string, payload any) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	data = append(data, '\n')

	return writeFile(path, data)
}

// WriteMeta is a small semantic wrapper around WriteJSON for metadata files.
func WriteMeta(path string, payload any) error {
	return WriteJSON(path, payload)
}

// WriteLines writes a deterministic newline-delimited text file.
func WriteLines(path string, lines []string) error {
	normalized := make([]string, 0, len(lines))
	seen := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if _, exists := seen[line]; exists {
			continue
		}
		seen[line] = struct{}{}
		normalized = append(normalized, line)
	}
	sort.Strings(normalized)

	body := ""
	if len(normalized) > 0 {
		body = strings.Join(normalized, "\n") + "\n"
	}

	return writeFile(path, []byte(body))
}

func sortedUniqueAddresses(proxies []proxy.Proxy) ([]string, error) {
	cloned := append([]proxy.Proxy(nil), proxies...)
	proxy.Sort(cloned)

	seen := make(map[string]struct{}, len(cloned))
	lines := make([]string, 0, len(cloned))
	for _, candidate := range cloned {
		if err := candidate.Validate(); err != nil {
			return nil, fmt.Errorf("invalid proxy %q: %w", candidate.String(), err)
		}

		address := candidate.String()
		if _, exists := seen[address]; exists {
			continue
		}
		seen[address] = struct{}{}
		lines = append(lines, address)
	}

	sort.Strings(lines)
	return lines, nil
}

func writeFile(path string, data []byte) error {
	cleaned := strings.TrimSpace(path)
	if cleaned == "" {
		return errors.New("output path is required")
	}
	cleaned = filepath.Clean(cleaned)
	if cleaned == "." || cleaned == string(filepath.Separator) {
		return fmt.Errorf("output path %q must include a file name", path)
	}

	dir := filepath.Dir(cleaned)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create output directory %q: %w", dir, err)
		}
	}

	if err := os.WriteFile(cleaned, data, 0o644); err != nil {
		return fmt.Errorf("write file %q: %w", cleaned, err)
	}
	return nil
}
