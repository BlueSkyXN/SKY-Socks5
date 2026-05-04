package source

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

func LoadFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open source file %q: %w", path, err)
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\uFEFF"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		urls = append(urls, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read source file %q: %w", path, err)
	}
	if len(urls) == 0 {
		return nil, errors.New("source file does not contain any source URLs")
	}
	return urls, nil
}
