package dpmaconnect

import (
	"encoding/xml"
	"fmt"
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

// parseDPMAError inspects a response body for DPMA error content and returns a
// typed error, or nil if the body is not an error.
//
// It detects both error formats the API uses:
//   - <Transaction> with TransactionErrorDetails (search/bulk endpoints)
//   - <Error Message_DE="..." Message_EN="..."/> (patent info endpoint)
//
// statusCode is threaded into the returned APIError. When statusCode >= 400 but
// no recognizable error XML is present, a fallback APIError with a body preview
// is returned. Pass statusCode 0 when no HTTP status is available (e.g. when
// inspecting already-buffered data in the parse helpers).
func parseDPMAError(body []byte, statusCode int) error {
	// Try <Transaction> error format.
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

	// Try <Error> element format.
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

	// No recognizable error XML; fall back to a generic error on HTTP failure.
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
