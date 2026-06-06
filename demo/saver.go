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

// maxExampleResponseBytes caps the size of a saved response body. Normal single
// responses (search hits, PDFs, register extracts) are kept even at tens of MB;
// only the very large weekly bulk archives (hundreds of MB) are skipped, since
// they must not be committed as examples or diffed by the weekly response-watch.
// Oversized responses are recorded as a short size note instead of the full body.
const maxExampleResponseBytes = 50 << 20 // 50 MiB

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

	if len(response) > maxExampleResponseBytes {
		note := fmt.Sprintf("response omitted: %d bytes (bulk endpoint, exceeds %d byte example cap)\n", len(response), maxExampleResponseBytes)
		noteFile := filepath.Join(dir, "response.info.txt")
		if err := os.WriteFile(noteFile, []byte(note), 0600); err != nil {
			return fmt.Errorf("failed to save response note: %w", err)
		}
		return nil
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
	fmt.Fprintf(&sb, "Method: %s\n\n", method)
	if len(params) > 0 {
		sb.WriteString("Parameters:\n")
		for k, v := range params {
			fmt.Fprintf(&sb, "  %s: %s\n", k, v)
		}
	}
	return sb.String()
}
