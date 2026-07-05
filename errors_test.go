package dpmaconnect

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestXMLRootName(t *testing.T) {
	tests := []struct {
		name string
		body []byte
		want string
	}{
		{"transaction with declaration", []byte(`<?xml version="1.0" encoding="UTF-8"?><Transaction/>`), "Transaction"},
		{"declaration-less error", []byte(`<Error Message_DE="x" Message_EN="y"/>`), "Error"},
		{"patent hit list fixture", patentSearchXML, "PatentHitList"},
		{"trademark hit list fixture", trademarkSearchXML, "HitList"},
		{"design hit list fixture", designSearchXML, "DesignHitList"},
		{"trademark error fixture", trademarkSearchErrorXML, "HitList"},
		{"design error fixture", designSearchErrorXML, "DesignHitList"},
		{"patent info fixture", patentInfoXML, "dpma-patent-document"},
		{"BOM prefix", append([]byte{0xEF, 0xBB, 0xBF}, []byte(`<Transaction/>`)...), "Transaction"},
		{"leading whitespace", []byte("\n  <HitList/>"), "HitList"},
		{"zip payload", []byte("PK\x03\x04binary"), ""},
		{"pdf payload", []byte("%PDF-1.7 binary"), ""},
		{"plain text", []byte("not xml at all"), ""},
		{"empty", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := xmlRootName(tt.body); got != tt.want {
				t.Errorf("xmlRootName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Binary bulk payloads must never be fed to xml.Unmarshal: only the bounded
// head is inspected, so a large ZIP body is cheap to classify and yields no
// false error.
func TestParseDPMAError_BinaryBodySkipsXMLParsing(t *testing.T) {
	body := append([]byte("PK\x03\x04"), bytes.Repeat([]byte{0x42}, 1<<20)...)

	if err := parseDPMAError(body, http.StatusOK); err != nil {
		t.Errorf("parseDPMAError(zip, 200) = %v, want nil", err)
	}

	err := parseDPMAError(body, http.StatusInternalServerError)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("parseDPMAError(zip, 500) = %T, want *APIError fallback", err)
	}
}

// The no-error-XML fallback must produce the typed taxonomy: AuthError for
// 401/403, RateLimitError for 429, generic APIError otherwise. Responses that
// do carry DPMA error XML keep their content-typed *APIError (E002 on 403).
func TestParseDPMAError_TypedFallbacks(t *testing.T) {
	tests := []struct {
		name       string
		body       []byte
		statusCode int
		check      func(t *testing.T, err error)
	}{
		{
			"401 empty body", nil, http.StatusUnauthorized,
			func(t *testing.T, err error) {
				var authErr *AuthError
				if !errors.As(err, &authErr) {
					t.Fatalf("got %T (%v), want *AuthError", err, err)
				}
				if authErr.StatusCode != http.StatusUnauthorized {
					t.Errorf("StatusCode = %d, want 401", authErr.StatusCode)
				}
			},
		},
		{
			"403 html body", []byte("<html><body>Forbidden</body></html>"), http.StatusForbidden,
			func(t *testing.T, err error) {
				var authErr *AuthError
				if !errors.As(err, &authErr) {
					t.Fatalf("got %T (%v), want *AuthError", err, err)
				}
				if !strings.Contains(authErr.Message, "Forbidden") {
					t.Errorf("Message = %q, want body preview", authErr.Message)
				}
			},
		},
		{
			"429 plain body", []byte("too many requests"), http.StatusTooManyRequests,
			func(t *testing.T, err error) {
				var rlErr *RateLimitError
				if !errors.As(err, &rlErr) {
					t.Fatalf("got %T (%v), want *RateLimitError", err, err)
				}
			},
		},
		{
			"500 stays APIError", []byte("boom"), http.StatusInternalServerError,
			func(t *testing.T, err error) {
				var apiErr *APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("got %T (%v), want *APIError", err, err)
				}
			},
		},
		{
			"403 with E002 transaction XML stays APIError",
			[]byte(`<?xml version="1.0"?><Transaction><PatentTransactionBody><TransactionErrorDetails><TransactionError><TransactionErrorCode>E002</TransactionErrorCode><TransactionErrorText>Permission denied</TransactionErrorText></TransactionError></TransactionErrorDetails></PatentTransactionBody></Transaction>`),
			http.StatusForbidden,
			func(t *testing.T, err error) {
				var apiErr *APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("got %T (%v), want *APIError", err, err)
				}
				if apiErr.Code != "E002" {
					t.Errorf("Code = %q, want E002", apiErr.Code)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseDPMAError(tt.body, tt.statusCode)
			if err == nil {
				t.Fatal("expected an error")
			}
			tt.check(t, err)
		})
	}
}

// The real search-error fixtures must surface as errors while their success
// twins pass through cleanly, with the unmarshal gated on the root element.
func TestParseDPMAError_FixtureEnvelopes(t *testing.T) {
	tests := []struct {
		name    string
		body    []byte
		wantErr bool
	}{
		{"trademark search error", trademarkSearchErrorXML, true},
		{"design search error", designSearchErrorXML, true},
		{"patent search success", patentSearchXML, false},
		{"trademark search success", trademarkSearchXML, false},
		{"design search success", designSearchXML, false},
		{"patent info success", patentInfoXML, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseDPMAError(tt.body, http.StatusOK)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseDPMAError() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				var apiErr *APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("got %T, want *APIError", err)
				}
				if !strings.Contains(apiErr.Message, "not admissible") {
					t.Errorf("Message = %q, want the DPMA error text", apiErr.Message)
				}
			}
		})
	}
}

// A declaration-less <Error .../> root (patent info endpoint) must be detected.
func TestParseDPMAError_DeclarationlessError(t *testing.T) {
	body := []byte(`<Error Message_DE="Nicht gefunden" Message_EN="Not found"/>`)
	err := parseDPMAError(body, http.StatusOK)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("got %T (%v), want *APIError", err, err)
	}
	if apiErr.Message != "Not found" {
		t.Errorf("Message = %q, want Not found", apiErr.Message)
	}
}

// streamResponse previously only recognized "<?xml"/"<Tra" prefixes, missing the
// declaration-less <Error .../> and hit-list envelopes: those were streamed into
// the destination file as if they were payload. All XML bodies are now inspected.
func TestStreamResponse_DetectsAllErrorEnvelopes(t *testing.T) {
	tests := []struct {
		name string
		body []byte
	}{
		{"declaration-less error", []byte(`<Error Message_DE="x" Message_EN="kaputt"/>`)},
		{"trademark hit list error", trademarkSearchErrorXML},
		{"design hit list error", designSearchErrorXML},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(tt.body)),
			}
			var buf bytes.Buffer
			err := streamResponse(resp, nil, "test", &buf)
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("got %T (%v), want *APIError", err, err)
			}
			if buf.Len() != 0 {
				t.Errorf("error body must not be written to the destination, got %d bytes", buf.Len())
			}
		})
	}
}

// Binary PDF payloads and non-error XML must still stream through unchanged.
func TestStreamResponse_Passthrough(t *testing.T) {
	tests := []struct {
		name string
		body []byte
	}{
		{"pdf binary", append([]byte("%PDF-1.7 "), bytes.Repeat([]byte{0x37}, 256)...)},
		{"non-error hit list XML", patentSearchXML},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(tt.body)),
			}
			var buf bytes.Buffer
			if err := streamResponse(resp, nil, "test", &buf); err != nil {
				t.Fatalf("streamResponse() error = %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tt.body) {
				t.Errorf("streamed %d bytes, want %d unchanged", buf.Len(), len(tt.body))
			}
		})
	}
}
