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
	if e.Code != "" {
		return fmt.Sprintf("API error %s: %s (HTTP %d)", e.Code, e.Message, e.StatusCode)
	}
	return fmt.Sprintf("API error: %s (HTTP %d)", e.Message, e.StatusCode)
}
