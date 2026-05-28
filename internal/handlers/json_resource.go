package handlers

import (
	"fmt"
	"os"
	"strings"
)

func loadJSONResource(path string, embedded []byte, name string) ([]byte, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return embedded, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s from %q: %w", name, path, err)
	}
	return data, nil
}
