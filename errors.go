package dpmaconnect

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
)

// ErrorResponse represents the XML error response structure from DPMA API.
// The DPMA API uses different body element names depending on the service:
//   - TradeMarkTransactionBody (trademark service)
//   - DesignTransactionBody (design service)
//   - PatentTransactionBody (patent service)
//
// All share the same nested TransactionErrorDetails structure.
// We parse all three variants and use whichever has error content.
type ErrorResponse struct {
	XMLName       xml.Name        `xml:"Transaction"`
	TradeMarkBody transactionBody `xml:"TradeMarkTransactionBody"`
	DesignBody    transactionBody `xml:"DesignTransactionBody"`
	PatentBody    transactionBody `xml:"PatentTransactionBody"`
}

type transactionBody struct {
	TransactionErrorDetails struct {
		TransactionError struct {
			TransactionErrorCode string `xml:"TransactionErrorCode"`
			TransactionErrorText string `xml:"TransactionErrorText"`
		} `xml:"TransactionError"`
	} `xml:"TransactionErrorDetails"`
}

// errorCodeAndText returns the error code and text from whichever body has content.
func (r *ErrorResponse) errorCodeAndText() (code, text string) {
	for _, body := range []transactionBody{r.TradeMarkBody, r.DesignBody, r.PatentBody} {
		c := body.TransactionErrorDetails.TransactionError.TransactionErrorCode
		t := body.TransactionErrorDetails.TransactionError.TransactionErrorText
		if c != "" || t != "" {
			return c, t
		}
	}
	return "", ""
}

// simpleErrorResponse represents the <Error> root element returned by some DPMA endpoints.
// Format: <Error Message_DE="..." Message_EN="..."/>
type simpleErrorResponse struct {
	XMLName   xml.Name `xml:"Error"`
	MessageDE string   `xml:"Message_DE,attr"`
	MessageEN string   `xml:"Message_EN,attr"`
}

// hitListErrorResponse represents an error reported inside a search hit-list
// envelope. The patent/trademark search service returns a bad query as
// <HitList><ErrorMessage>...</ErrorMessage></HitList> (the message text itself
// carries embedded Message_DE/Message_EN strings), while the design search
// service returns <DesignHitList HitCount="0" Message_DE="..." Message_EN="..."/>.
//
// All three hit-list roots (PatentHitList, HitList, DesignHitList) are matched
// by allowing the XMLName to be empty; only the presence of an error marker
// (ErrorMessage element or Message_* attribute) flags an error.
type hitListErrorResponse struct {
	ErrorMessage string `xml:"ErrorMessage"`
	MessageDE    string `xml:"Message_DE,attr"`
	MessageEN    string `xml:"Message_EN,attr"`
}

// NotFoundError represents resource not found errors
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

// DataNotAvailableError represents data unavailable for requested period
type DataNotAvailableError struct{}

func (e *DataNotAvailableError) Error() string {
	return "data not available"
}

// APIError represents a generic API error from DPMA
type APIError struct {
	Code       string
	Message    string
	StatusCode int
}

func (e *APIError) Error() string {
	suffix := ""
	if e.StatusCode != 0 {
		suffix = fmt.Sprintf(" (HTTP %d)", e.StatusCode)
	}
	if e.Code != "" {
		return fmt.Sprintf("API error %s: %s%s", e.Code, e.Message, suffix)
	}
	return fmt.Sprintf("API error: %s%s", e.Message, suffix)
}

// AuthError is returned for HTTP 401/403 responses that carry no DPMA error XML
// (real DPMA authentication failures arrive without a Transaction body).
// Responses that do carry error XML keep their content-typed *APIError.
type AuthError struct {
	StatusCode int
	Message    string
}

func (e *AuthError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("authentication failed (HTTP %d): %s", e.StatusCode, e.Message)
	}
	return "authentication failed: " + e.Message
}

// RateLimitError is returned for HTTP 429 responses. RetryAfter is the value of
// the Retry-After header in seconds when known (0 otherwise).
type RateLimitError struct {
	RetryAfter int
	Message    string
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("rate limited, retry after %d seconds: %s", e.RetryAfter, e.Message)
	}
	return "rate limited: " + e.Message
}

// ResolutionLoopError is returned when resolving a patent number keeps producing
// another non-registered number instead of terminating - i.e. the register data
// points back at itself.
type ResolutionLoopError struct {
	Number string
}

func (e *ResolutionLoopError) Error() string {
	return fmt.Sprintf("patent number resolution for %s did not terminate: register data points back at a non-registered number", e.Number)
}

// xmlRootHead bounds how many leading bytes xmlRootName inspects. The root start
// element of every DPMA response appears well within this window; anything beyond
// it (multi-gigabyte bulk payloads) must never be scanned.
const xmlRootHead = 4096

// xmlRootName returns the local name of the first XML start element within the
// leading bytes of body, or "" when the body does not start with XML (binary
// ZIP/PDF payloads, plain text). A UTF-8 BOM is tolerated.
func xmlRootName(body []byte) string {
	head := body
	if len(head) > xmlRootHead {
		head = head[:xmlRootHead]
	}
	head = bytes.TrimPrefix(head, []byte{0xEF, 0xBB, 0xBF})
	head = bytes.TrimSpace(head)
	if !bytes.HasPrefix(head, []byte("<")) {
		return ""
	}
	dec := xml.NewDecoder(bytes.NewReader(head))
	for {
		tok, err := dec.Token()
		if err != nil {
			return ""
		}
		if start, ok := tok.(xml.StartElement); ok {
			return start.Name.Local
		}
	}
}

// parseDPMAError inspects a response body for DPMA error content and returns a
// typed error, or nil if the body is not an error.
//
// It detects the three error formats the API uses:
//   - <Transaction> with TransactionErrorDetails (search/bulk endpoints)
//   - <Error Message_DE="..." Message_EN="..."/> (patent info endpoint)
//   - hit-list envelopes: <HitList><ErrorMessage>, <PatentHitList>, and
//     <DesignHitList HitCount="0" Message_DE=.../> (search endpoints, HTTP 200)
//
// Each unmarshal is gated on the actual root element (see xmlRootName), so a
// body is decoded at most once and binary ZIP/PDF payloads are never fed to
// xml.Unmarshal at all.
//
// statusCode is threaded into the returned APIError. When statusCode >= 400 but
// no recognizable error XML is present, a typed fallback is returned: AuthError
// for 401/403, RateLimitError for 429, and a generic APIError with a body
// preview otherwise. Pass statusCode 0 when no HTTP status is available (e.g.
// when inspecting already-buffered data in the parse helpers).
func parseDPMAError(body []byte, statusCode int) error {
	switch xmlRootName(body) {
	case "Transaction":
		// <Transaction> error format.
		var errResp ErrorResponse
		if err := xml.Unmarshal(body, &errResp); err == nil {
			code, text := errResp.errorCodeAndText()

			// "Data not available" can arrive as code "E001" or "Error".
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

	case "Error":
		// <Error> element format.
		var simpleErr simpleErrorResponse
		if err := xml.Unmarshal(body, &simpleErr); err == nil {
			msg := simpleErr.MessageEN
			if msg == "" {
				msg = simpleErr.MessageDE
			}
			if msg != "" {
				return &APIError{
					Code:       "Error",
					Message:    msg,
					StatusCode: statusCode,
				}
			}
		}

	case "HitList", "PatentHitList", "DesignHitList":
		// Hit-list error envelopes: <HitList><ErrorMessage> (patent/trademark)
		// and <DesignHitList HitCount="0" Message_DE=.../> (design). These arrive
		// with HTTP 200 and an otherwise empty hit list, so without this check a bad
		// query would silently parse as zero hits instead of surfacing an error.
		var hitErr hitListErrorResponse
		if err := xml.Unmarshal(body, &hitErr); err == nil {
			msg := hitErr.ErrorMessage
			if msg == "" {
				msg = hitErr.MessageEN
			}
			if msg == "" {
				msg = hitErr.MessageDE
			}
			if msg != "" {
				return &APIError{
					Code:       "Error",
					Message:    msg,
					StatusCode: statusCode,
				}
			}
		}
	}

	// No recognizable error XML; fall back to a typed error on HTTP failure.
	// Real DPMA 401/403 responses carry no Transaction XML, so they land here.
	if statusCode >= 400 {
		runes := []rune(string(body))
		var preview string
		if len(runes) > 200 {
			preview = string(runes[:200]) + "..."
		} else {
			preview = string(body)
		}
		switch statusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			return &AuthError{StatusCode: statusCode, Message: preview}
		case http.StatusTooManyRequests:
			return &RateLimitError{Message: preview}
		}
		return &APIError{
			Message:    fmt.Sprintf("unexpected error response: %s", preview),
			StatusCode: statusCode,
		}
	}
	return nil
}
