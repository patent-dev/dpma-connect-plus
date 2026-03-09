package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileFormat represents the format of a saved response
type FileFormat string

const (
	FormatXML  FileFormat = "xml"
	FormatText FileFormat = "txt"
	FormatBin  FileFormat = "bin"
)

// ExampleSaver saves request/response pairs to disk
type ExampleSaver struct {
	baseDir string
}

// NewExampleSaver creates a new ExampleSaver with the specified base directory
func NewExampleSaver(baseDir string) *ExampleSaver {
	return &ExampleSaver{baseDir: baseDir}
}

// SaveExample saves a request description and response data to the examples directory
func (s *ExampleSaver) SaveExample(endpointName string, requestDesc string, response []byte, format FileFormat) error {
	dir := filepath.Join(s.baseDir, endpointName)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	requestFile := filepath.Join(dir, "request.txt")
	if err := os.WriteFile(requestFile, []byte(requestDesc), 0600); err != nil {
		return fmt.Errorf("failed to save request: %w", err)
	}

	responseFile := filepath.Join(dir, fmt.Sprintf("response.%s", format))
	if err := os.WriteFile(responseFile, response, 0600); err != nil {
		return fmt.Errorf("failed to save response: %w", err)
	}

	return nil
}

// DetectFormat detects whether data is XML or binary
func DetectFormat(data []byte) FileFormat {
	if len(data) == 0 {
		return FormatText
	}

	trimmed := strings.TrimSpace(string(data))
	if strings.HasPrefix(trimmed, "<?xml") || strings.HasPrefix(trimmed, "<") {
		return FormatXML
	}

	return FormatBin
}

// FormatRequestDescription formats a request description with parameters
func FormatRequestDescription(method string, params map[string]string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Method: %s\n\n", method))
	if len(params) > 0 {
		sb.WriteString("Parameters:\n")
		for k, v := range params {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}
	return sb.String()
}
