// Package dpmaconnect provides a Go client for the DPMA Connect Plus API
// (DPMAregister web services) for accessing German patent, design, and
// trademark data from the German Patent and Trade Mark Office (DPMA).
//
// Usage:
//
//	config := dpmaconnect.DefaultConfig()
//	config.Username = "your-username"
//	config.Password = "your-password"
//	client, err := dpmaconnect.NewClient(config)
//
// The client is safe for concurrent use by multiple goroutines.
package dpmaconnect

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/patent-dev/dpma-connect-plus/generated"
)

// Client is the main DPMA Connect Plus API client.
// It is safe for concurrent use by multiple goroutines.
type Client struct {
	config    *Config
	generated *generated.ClientWithResponses
}

// Config holds client configuration.
//
// If HTTPClient is set, Timeout is ignored and the custom client's timeout
// applies instead. Callers should configure timeouts on the custom client
// directly.
type Config struct {
	BaseURL    string
	Username   string
	Password   string
	Timeout    time.Duration // HTTP client timeout (default: 20 minutes for bulk downloads)
	HTTPClient *http.Client  // Optional custom HTTP client; if set, Timeout is ignored
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL: "https://dpmaconnect.dpma.de/dpmaws/rest-services",
		Timeout: 20 * time.Minute,
	}
}

// NewClient creates a new DPMA Connect Plus API client
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if config.Username == "" || config.Password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	if config.BaseURL == "" {
		return nil, fmt.Errorf("baseURL is required")
	}

	// Validate BaseURL format
	u, urlErr := url.Parse(config.BaseURL)
	if urlErr != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("baseURL must be a valid HTTP(S) URL")
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		timeout := config.Timeout
		if timeout == 0 {
			timeout = 20 * time.Minute
		}
		httpClient = &http.Client{
			Timeout: timeout,
		}
	}

	authToken := base64.StdEncoding.EncodeToString(
		fmt.Appendf(nil, "%s:%s", config.Username, config.Password),
	)

	requestEditor := generated.RequestEditorFn(func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Basic "+authToken)
		return nil
	})

	genClient, err := generated.NewClientWithResponses(
		config.BaseURL,
		generated.WithHTTPClient(httpClient),
		generated.WithRequestEditorFn(requestEditor),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Client{
		config:    config,
		generated: genClient,
	}, nil
}

// checkErrorResponse attempts to parse XML error response.
// The DPMA API returns errors in Transaction XML with service-specific body
// elements (TradeMarkTransactionBody, DesignTransactionBody, PatentTransactionBody).
func checkErrorResponse(body []byte, statusCode int) error {
	var errResp ErrorResponse
	if err := xml.Unmarshal(body, &errResp); err == nil {
		code, text := errResp.errorCodeAndText()

		// Check for "Data not available" error (code can be "E001" or "Error")
		if (code == "E001" || code == "Error") && text == "Data not available" {
			return &DataNotAvailableError{}
		}

		if code != "" || text != "" {
			return &APIError{
				Code:       code,
				Message:    text,
				StatusCode: statusCode,
			}
		}
	}

	// If XML parsing failed and status indicates an error, return fallback with body preview
	if statusCode >= 400 {
		runes := []rune(string(body))
		var preview string
		if len(runes) > 200 {
			preview = string(runes[:200]) + "..."
		} else {
			preview = string(body)
		}
		return &APIError{
			Message:    fmt.Sprintf("unexpected error response: %s", preview),
			StatusCode: statusCode,
		}
	}
	return nil
}

// GetVersion retrieves version information for a service
func (c *Client) GetVersion(ctx context.Context, service string) (string, error) {
	resp, err := c.generated.GetVersionWithResponse(ctx, generated.GetVersionParamsService(service))
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return "", &APIError{
			Message:    "failed to get version",
			StatusCode: resp.StatusCode(),
		}
	}

	return string(resp.Body), nil
}

// streamResponse handles streaming an HTTP response to dst with proper error detection.
// It peeks at the first bytes to detect XML error responses (which the DPMA API
// returns even on 200 OK for "Data not available"), then streams the rest directly
// without buffering the entire response in memory.
func streamResponse(resp *http.Response, err error, errMsg string, dst io.Writer) error {
	if err != nil {
		return fmt.Errorf("%s: %w", errMsg, err)
	}
	if resp == nil {
		return fmt.Errorf("%s: no response", errMsg)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read error response: %w", err)
		}
		if apiErr := checkErrorResponse(body, resp.StatusCode); apiErr != nil {
			return apiErr
		}
		return &APIError{Message: errMsg, StatusCode: resp.StatusCode}
	}

	// Peek at the beginning to detect XML error responses returned with 200 OK.
	// Successful bulk downloads are binary (ZIP/PDF), so XML indicates an error.
	peek := make([]byte, 64)
	n, peekErr := io.ReadFull(resp.Body, peek)
	if peekErr != nil && peekErr != io.ErrUnexpectedEOF && peekErr != io.EOF {
		return fmt.Errorf("failed to read response: %w", peekErr)
	}
	peek = peek[:n]

	trimmed := bytes.TrimSpace(peek)
	// Strip UTF-8 BOM if present
	if bytes.HasPrefix(trimmed, []byte{0xEF, 0xBB, 0xBF}) {
		trimmed = trimmed[3:]
	}
	if bytes.HasPrefix(trimmed, []byte("<?xml")) || bytes.HasPrefix(trimmed, []byte("<Tra")) {
		rest, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
		full := make([]byte, 0, len(peek)+len(rest))
		full = append(full, peek...)
		full = append(full, rest...)
		if apiErr := checkErrorResponse(full, resp.StatusCode); apiErr != nil {
			return apiErr
		}
		// Valid XML that isn't an error - write it out
		_, err = dst.Write(full)
		if err != nil {
			return fmt.Errorf("failed to write data: %w", err)
		}
		return nil
	}

	// Binary data - stream directly without buffering
	if _, err := dst.Write(peek); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}
	if _, err := io.Copy(dst, resp.Body); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}
	return nil
}

// bulkResult checks a buffered API response for errors.
// It checks for XML error responses even on 200 OK, since the DPMA API
// returns errors with 200 status for some endpoints.
func bulkResult(body []byte, statusCode int, errMsg string) ([]byte, error) {
	if apiErr := checkErrorResponse(body, statusCode); apiErr != nil {
		return nil, apiErr
	}
	if statusCode != http.StatusOK {
		return nil, &APIError{Message: errMsg, StatusCode: statusCode}
	}
	return body, nil
}

// resourceResult checks a buffered response for errors, with 404 handling.
func resourceResult(body []byte, statusCode int, resource, id, errMsg string) ([]byte, error) {
	if statusCode == http.StatusNotFound {
		return nil, &NotFoundError{Resource: resource, ID: id}
	}
	return bulkResult(body, statusCode, errMsg)
}
