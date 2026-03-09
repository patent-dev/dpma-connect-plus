package dpmaconnect

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testDate() time.Time {
	return time.Date(2024, 10, 23, 0, 0, 0, 0, time.UTC)
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Username: "testuser",
				Password: "testpass",
				BaseURL:  "https://dpmaconnect.dpma.de/dpmaws/rest-services",
			},
			wantErr: false,
		},
		{
			name: "nil config uses default",
			config: func() *Config {
				c := DefaultConfig()
				c.Username = "testuser"
				c.Password = "testpass"
				return c
			}(),
			wantErr: false,
		},
		{
			name: "missing username",
			config: &Config{
				Password: "testpass",
			},
			wantErr: true,
		},
		{
			name: "missing password",
			config: &Config{
				Username: "testuser",
			},
			wantErr: true,
		},
		{
			name:    "empty config",
			config:  &Config{},
			wantErr: true,
		},
		{
			name: "short baseURL does not panic",
			config: &Config{
				Username: "testuser",
				Password: "testpass",
				BaseURL:  "x",
			},
			wantErr: true,
		},
		{
			name: "rejects httpnotaurl",
			config: &Config{
				Username: "testuser",
				Password: "testpass",
				BaseURL:  "httpnotaurl",
			},
			wantErr: true,
		},
		{
			name: "rejects ftp scheme",
			config: &Config{
				Username: "testuser",
				Password: "testpass",
				BaseURL:  "ftp://example.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func setupMockServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)

	config := &Config{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
		Timeout:  10 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	return server, client
}

// requireAuth checks the Authorization header and fails the test if missing.
func requireAuth(t *testing.T, r *http.Request) {
	t.Helper()
	if !strings.HasPrefix(r.Header.Get("Authorization"), "Basic ") {
		t.Error("missing or invalid Authorization header")
	}
}

// requirePath checks that the request path contains the expected substring.
func requirePath(t *testing.T, r *http.Request, substr string) {
	t.Helper()
	if !strings.Contains(r.URL.Path, substr) {
		t.Errorf("expected path to contain %q, got %q", substr, r.URL.Path)
	}
}

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name       string
		service    string
		statusCode int
		response   string
		wantErr    bool
	}{
		{
			name:       "successful version request",
			service:    "DPMAregisterPatService",
			statusCode: http.StatusOK,
			response:   "DPMAregisterPatService version 1.0",
			wantErr:    false,
		},
		{
			name:       "unauthorized",
			service:    "DPMAregisterPatService",
			statusCode: http.StatusUnauthorized,
			response:   "",
			wantErr:    true,
		},
		{
			name:       "internal server error",
			service:    "DPMAregisterPatService",
			statusCode: http.StatusInternalServerError,
			response:   "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				auth := r.Header.Get("Authorization")
				if !strings.HasPrefix(auth, "Basic ") {
					t.Error("Missing or invalid Authorization header")
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}

			server, client := setupMockServer(t, handler)
			defer server.Close()

			ctx := context.Background()
			version, err := client.GetVersion(ctx, tt.service)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && version != tt.response {
				t.Errorf("GetVersion() = %s, want %s", version, tt.response)
			}
		})
	}
}

func TestCheckErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		body       []byte
		statusCode int
		wantErr    bool
		errType    string
	}{
		{
			name: "data not available E001",
			body: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Transaction>
  <TradeMarkTransactionBody>
    <TransactionErrorDetails>
      <TransactionError>
        <TransactionErrorCode>E001</TransactionErrorCode>
        <TransactionErrorText>Data not available</TransactionErrorText>
      </TransactionError>
    </TransactionErrorDetails>
  </TradeMarkTransactionBody>
</Transaction>`),
			statusCode: http.StatusOK,
			wantErr:    true,
			errType:    "DataNotAvailableError",
		},
		{
			name: "data not available with Error code",
			body: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Transaction>
  <TradeMarkTransactionBody>
    <TransactionErrorDetails>
      <TransactionError>
        <TransactionErrorCode>Error</TransactionErrorCode>
        <TransactionErrorText>Data not available</TransactionErrorText>
      </TransactionError>
    </TransactionErrorDetails>
  </TradeMarkTransactionBody>
</Transaction>`),
			statusCode: http.StatusOK,
			wantErr:    true,
			errType:    "DataNotAvailableError",
		},
		{
			name: "generic API error",
			body: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Transaction>
  <TradeMarkTransactionBody>
    <TransactionErrorDetails>
      <TransactionError>
        <TransactionErrorCode>E002</TransactionErrorCode>
        <TransactionErrorText>Some other error</TransactionErrorText>
      </TransactionError>
    </TransactionErrorDetails>
  </TradeMarkTransactionBody>
</Transaction>`),
			statusCode: http.StatusBadRequest,
			wantErr:    true,
			errType:    "APIError",
		},
		{
			name:       "not an error response",
			body:       []byte("<valid>data</valid>"),
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "malformed XML with 500 returns fallback error",
			body:       []byte("not xml at all"),
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
			errType:    "APIError",
		},
		{
			name:       "malformed XML with 200 returns nil",
			body:       []byte("not xml at all"),
			statusCode: http.StatusOK,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkErrorResponse(tt.body, tt.statusCode)

			if (err != nil) != tt.wantErr {
				t.Errorf("checkErrorResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != "" {
				switch tt.errType {
				case "DataNotAvailableError":
					if _, ok := err.(*DataNotAvailableError); !ok {
						t.Errorf("checkErrorResponse() error type = %T, want *DataNotAvailableError", err)
					}
				case "APIError":
					if _, ok := err.(*APIError); !ok {
						t.Errorf("checkErrorResponse() error type = %T, want *APIError", err)
					}
				}
			}
		})
	}
}

func TestCheckErrorResponse_DesignTransactionBody(t *testing.T) {
	body := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Transaction>
  <DesignTransactionBody>
    <TransactionErrorDetails>
      <TransactionError>
        <TransactionErrorCode>E001</TransactionErrorCode>
        <TransactionErrorText>Data not available</TransactionErrorText>
      </TransactionError>
    </TransactionErrorDetails>
  </DesignTransactionBody>
</Transaction>`)
	err := checkErrorResponse(body, http.StatusOK)
	if err == nil {
		t.Fatal("expected error for DesignTransactionBody error")
	}
	if _, ok := err.(*DataNotAvailableError); !ok {
		t.Errorf("expected *DataNotAvailableError, got %T: %v", err, err)
	}
}

func TestCheckErrorResponse_PatentTransactionBody(t *testing.T) {
	body := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Transaction>
  <PatentTransactionBody>
    <TransactionErrorDetails>
      <TransactionError>
        <TransactionErrorCode>E002</TransactionErrorCode>
        <TransactionErrorText>Permission denied</TransactionErrorText>
      </TransactionError>
    </TransactionErrorDetails>
  </PatentTransactionBody>
</Transaction>`)
	err := checkErrorResponse(body, http.StatusForbidden)
	if err == nil {
		t.Fatal("expected error for PatentTransactionBody error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Code != "E002" {
		t.Errorf("expected code E002, got %s", apiErr.Code)
	}
}

func TestCheckErrorResponse_ValidXMLEmptyCodes(t *testing.T) {
	// Valid XML that parses as ErrorResponse but has empty code/text fields
	body := []byte(`<?xml version="1.0"?><Transaction><TransactionHeader/><TradeMarkTransactionBody/></Transaction>`)

	// With 200, should return nil (not an error)
	err := checkErrorResponse(body, http.StatusOK)
	if err != nil {
		t.Errorf("expected nil for valid XML with empty codes on 200, got %v", err)
	}

	// With 400, empty codes in XML so falls through to fallback
	err = checkErrorResponse(body, http.StatusBadRequest)
	if err == nil {
		t.Error("expected error for valid XML with empty codes on 400")
	}
}

func TestCheckErrorResponse_LongBodyTruncated(t *testing.T) {
	body := bytes.Repeat([]byte("x"), 500)
	err := checkErrorResponse(body, http.StatusInternalServerError)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if !strings.HasSuffix(apiErr.Message, "...") {
		t.Error("expected truncated body preview to end with '...'")
	}
}

func TestStreamResponse_BinaryData(t *testing.T) {
	body := []byte("PK\x03\x04 this is a ZIP file with enough bytes to fill the peek buffer fully ok")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}

	var buf bytes.Buffer
	err := streamResponse(resp, nil, "test", &buf)
	if err != nil {
		t.Fatalf("streamResponse() error = %v", err)
	}
	if !bytes.Equal(buf.Bytes(), body) {
		t.Errorf("streamResponse() wrote %d bytes, want %d", buf.Len(), len(body))
	}
}

func TestStreamResponse_XMLErrorOn200(t *testing.T) {
	// DPMA returns XML errors with 200 OK for "Data not available"
	xmlErr := `<?xml version="1.0" encoding="UTF-8"?>
<Transaction>
  <TradeMarkTransactionBody>
    <TransactionErrorDetails>
      <TransactionError>
        <TransactionErrorCode>E001</TransactionErrorCode>
        <TransactionErrorText>Data not available</TransactionErrorText>
      </TransactionError>
    </TransactionErrorDetails>
  </TradeMarkTransactionBody>
</Transaction>`
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(xmlErr)),
	}

	var buf bytes.Buffer
	err := streamResponse(resp, nil, "test", &buf)
	if err == nil {
		t.Fatal("expected error for XML error on 200")
	}
	var dna *DataNotAvailableError
	if !errors.As(err, &dna) {
		t.Errorf("expected *DataNotAvailableError, got %T: %v", err, err)
	}
}

func TestStreamResponse_Non200WithXMLError(t *testing.T) {
	xmlErr := `<?xml version="1.0"?>
<Transaction>
  <TradeMarkTransactionBody>
    <TransactionErrorDetails>
      <TransactionError>
        <TransactionErrorCode>E002</TransactionErrorCode>
        <TransactionErrorText>Permission denied</TransactionErrorText>
      </TransactionError>
    </TransactionErrorDetails>
  </TradeMarkTransactionBody>
</Transaction>`
	resp := &http.Response{
		StatusCode: http.StatusForbidden,
		Body:       io.NopCloser(strings.NewReader(xmlErr)),
	}

	var buf bytes.Buffer
	err := streamResponse(resp, nil, "test", &buf)
	if err == nil {
		t.Fatal("expected error for 403")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Code != "E002" {
		t.Errorf("expected code E002, got %s", apiErr.Code)
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", apiErr.StatusCode)
	}
}

func TestStreamResponse_Non200Plain(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader("Internal Server Error")),
	}

	err := streamResponse(resp, nil, "download failed", io.Discard)
	if err == nil {
		t.Fatal("expected error for 500")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
}

func TestStreamResponse_HTTPError(t *testing.T) {
	err := streamResponse(nil, errors.New("connection refused"), "test", io.Discard)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("expected wrapped error, got %v", err)
	}
}

func TestStreamResponse_NilResponse(t *testing.T) {
	err := streamResponse(nil, nil, "test", io.Discard)
	if err == nil {
		t.Fatal("expected error for nil response")
	}
	if !strings.Contains(err.Error(), "no response") {
		t.Errorf("expected 'no response' in error, got %v", err)
	}
}

func TestStreamResponse_BOMPrefixedXMLError(t *testing.T) {
	bom := "\xEF\xBB\xBF"
	xmlErr := bom + `<?xml version="1.0"?>
<Transaction>
  <TradeMarkTransactionBody>
    <TransactionErrorDetails>
      <TransactionError>
        <TransactionErrorCode>E001</TransactionErrorCode>
        <TransactionErrorText>Data not available</TransactionErrorText>
      </TransactionError>
    </TransactionErrorDetails>
  </TradeMarkTransactionBody>
</Transaction>`
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(xmlErr)),
	}

	var buf bytes.Buffer
	err := streamResponse(resp, nil, "test", &buf)
	if err == nil {
		t.Fatal("expected error for BOM-prefixed XML error")
	}
	var dna *DataNotAvailableError
	if !errors.As(err, &dna) {
		t.Errorf("expected *DataNotAvailableError, got %T: %v", err, err)
	}
}

func TestStreamResponse_DesignXMLErrorOn200(t *testing.T) {
	xmlErr := `<?xml version="1.0" encoding="UTF-8"?>
<Transaction>
  <DesignTransactionBody>
    <TransactionErrorDetails>
      <TransactionError>
        <TransactionErrorCode>E001</TransactionErrorCode>
        <TransactionErrorText>Data not available</TransactionErrorText>
      </TransactionError>
    </TransactionErrorDetails>
  </DesignTransactionBody>
</Transaction>`
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(xmlErr)),
	}

	var buf bytes.Buffer
	err := streamResponse(resp, nil, "test", &buf)
	if err == nil {
		t.Fatal("expected error for design XML error on 200")
	}
	var dna *DataNotAvailableError
	if !errors.As(err, &dna) {
		t.Errorf("expected *DataNotAvailableError, got %T: %v", err, err)
	}
}

func TestStreamResponse_ValidXMLNonError(t *testing.T) {
	// XML that isn't an error should be written through
	xmlData := `<?xml version="1.0"?><results><item>hello</item></results>`
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(xmlData)),
	}

	var buf bytes.Buffer
	err := streamResponse(resp, nil, "test", &buf)
	if err != nil {
		t.Fatalf("streamResponse() error = %v", err)
	}
	if buf.String() != xmlData {
		t.Errorf("expected XML passthrough, got %q", buf.String())
	}
}

func TestBulkMethodsRejectInvalidWeek(t *testing.T) {
	server, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called for invalid week")
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	ctx := context.Background()

	// []byte methods
	if _, err := client.GetDisclosureDocumentsXML(ctx, 2024, 0); err == nil {
		t.Error("expected error for week 0")
	}
	if _, err := client.GetDisclosureDocumentsXML(ctx, 2024, 54); err == nil {
		t.Error("expected error for week 54")
	}
	if _, err := client.GetDisclosureDocumentsXML(ctx, -1, 1); err == nil {
		t.Error("expected error for negative year")
	}

	// Stream methods
	if err := client.GetDisclosureDocumentsXMLStream(ctx, 2024, 0, io.Discard); err == nil {
		t.Error("expected error for week 0 (stream)")
	}
	if err := client.GetPatentSpecificationsPDFStream(ctx, 0, 1, io.Discard); err == nil {
		t.Error("expected error for year 0 (stream)")
	}
}

func TestRegisterExtractRejectsInvalidPeriod(t *testing.T) {
	server, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called for invalid period")
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	ctx := context.Background()
	date := testDate()

	if _, err := client.GetPatentRegisterExtract(ctx, date, "biweekly"); err == nil {
		t.Error("expected error for invalid period (patent)")
	}
	if err := client.GetPatentRegisterExtractStream(ctx, date, "biweekly", io.Discard); err == nil {
		t.Error("expected error for invalid period (patent stream)")
	}
	if _, err := client.GetDesignRegisterExtract(ctx, date, "bogus"); err == nil {
		t.Error("expected error for invalid period (design)")
	}
	if err := client.GetDesignRegisterExtractStream(ctx, date, "bogus", io.Discard); err == nil {
		t.Error("expected error for invalid period (design stream)")
	}
	if _, err := client.GetTrademarkRegisterExtract(ctx, date, "bogus"); err == nil {
		t.Error("expected error for invalid period (trademark)")
	}
	if err := client.GetTrademarkRegisterExtractStream(ctx, date, "bogus", io.Discard); err == nil {
		t.Error("expected error for invalid period (trademark stream)")
	}
}

func TestCustomHTTPClient(t *testing.T) {
	headerSeen := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Test") == "present" {
			headerSeen = true
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	customTransport := &headerTransport{header: "X-Custom-Test", value: "present"}
	customClient := &http.Client{Transport: customTransport}

	client, err := NewClient(&Config{
		BaseURL:    server.URL,
		Username:   "testuser",
		Password:   "testpass",
		HTTPClient: customClient,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.GetVersion(context.Background(), ServicePatent)
	if err != nil {
		t.Fatalf("GetVersion() error = %v", err)
	}

	if !headerSeen {
		t.Error("custom HTTP client was not used (custom header not seen by server)")
	}
}

type headerTransport struct {
	header string
	value  string
}

func (ht *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set(ht.header, ht.value)
	return http.DefaultTransport.RoundTrip(req)
}
