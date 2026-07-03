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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/patent-dev/dpma-connect-plus/generated"
)

// Version is the library version. It surfaces through the default User-Agent.
const Version = "0.3.4"

// DefaultUserAgent identifies this library in outbound requests.
const DefaultUserAgent = "dpma-connect-plus-go/" + Version + " (patent.dev; +https://github.com/patent-dev/dpma-connect-plus)"

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
	UserAgent  string        // UserAgent overrides the outbound User-Agent header.
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL: "https://dpmaconnect.dpma.de/dpmaws/rest-services",
		Timeout: 20 * time.Minute,
	}
}

// uaTransport adds the User-Agent header to every outgoing request.
type uaTransport struct {
	base      http.RoundTripper
	userAgent string
}

func (t *uaTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := req.Clone(req.Context())
	r.Header.Set("User-Agent", t.userAgent)
	return t.base.RoundTrip(r)
}

func transportOrDefault(c *http.Client) http.RoundTripper {
	if c.Transport != nil {
		return c.Transport
	}
	return http.DefaultTransport
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

	userAgent := config.UserAgent
	if userAgent == "" {
		userAgent = DefaultUserAgent
	}

	var httpClient *http.Client
	if config.HTTPClient != nil {
		// Copy the caller's client so the User-Agent transport can be added
		// without mutating the caller-owned client in place (which would also
		// stack a new uaTransport on every NewClient call). The shallow copy
		// preserves Timeout, Jar and CheckRedirect.
		c := *config.HTTPClient
		c.Transport = &uaTransport{base: transportOrDefault(config.HTTPClient), userAgent: userAgent}
		httpClient = &c
	} else {
		timeout := config.Timeout
		if timeout == 0 {
			timeout = 20 * time.Minute
		}
		httpClient = &http.Client{
			Timeout:   timeout,
			Transport: &uaTransport{base: http.DefaultTransport, userAgent: userAgent},
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

// GetVersion retrieves version information for a service
func (c *Client) GetVersion(ctx context.Context, service string) (string, error) {
	resp, err := c.generated.GetVersionWithResponse(ctx, generated.GetVersionParamsService(service))
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	if apiErr := parseDPMAError(resp.Body, resp.StatusCode()); apiErr != nil {
		return "", apiErr
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
		if apiErr := parseDPMAError(body, resp.StatusCode); apiErr != nil {
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
		if apiErr := parseDPMAError(full, resp.StatusCode); apiErr != nil {
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
	if apiErr := parseDPMAError(body, statusCode); apiErr != nil {
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

// weeklyResponse describes the shape shared by all generated *WithResponse types
// for weekly bulk endpoints: a buffered body plus an HTTP status code.
type weeklyResponse interface {
	StatusCode() int
}

// fetchWeeklyBulk formats the publication week, invokes a generated buffered call,
// and runs the response through bulkResult. It removes the per-endpoint copy-paste
// while keeping each public method's signature unchanged.
//
// body extracts the buffered bytes from the generated response (which expose Body
// as a field, not reachable through a generic constraint).
func fetchWeeklyBulk[R weeklyResponse](
	year, week int,
	getErrMsg, downloadErrMsg string,
	call func(pubWeek string) (R, error),
	body func(R) []byte,
) ([]byte, error) {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return nil, err
	}
	resp, err := call(pubWeek)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", getErrMsg, err)
	}
	return bulkResult(body(resp), resp.StatusCode(), downloadErrMsg)
}

// streamWeekly formats the publication week, invokes a generated streaming call,
// and streams the result to dst via streamResponse.
func streamWeekly(
	year, week int,
	getErrMsg string,
	dst io.Writer,
	call func(pubWeek string) (*http.Response, error),
) error {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return err
	}
	resp, err := call(pubWeek)
	return streamResponse(resp, err, getErrMsg, dst)
}
