package dpmaconnect

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/patent-dev/dpma-connect-plus/generated"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// isRegisteredNumber returns true if the input looks like a bare DPMA registered number
// (digits only, no DE prefix, no kind code). The DPMA getRegisterInfo API requires
// the full registered number including check digit (e.g., "100273629").
func isRegisteredNumber(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// SearchPatents executes a patent/utility model expert search query
func (c *Client) SearchPatents(ctx context.Context, query string) ([]byte, error) {
	resp, err := c.generated.SearchPatentsWithResponse(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search patents: %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "search failed")
}

// GetPatentPublicationPDF downloads a single patent publication in PDF format
func (c *Client) GetPatentPublicationPDF(ctx context.Context, documentID string) ([]byte, error) {
	resp, err := c.generated.GetPatentPublicationPDFWithResponse(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get patent PDF: %w", err)
	}
	return resourceResult(resp.Body, resp.StatusCode(), "patent publication", documentID, "failed to download PDF")
}

// GetPatentInfo retrieves patent information by registered number (digits only, including check digit).
func (c *Client) GetPatentInfo(ctx context.Context, registeredNumber string) ([]byte, error) {
	resp, err := c.generated.GetPatentInfoWithResponse(ctx, registeredNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get patent info: %w", err)
	}
	return resourceResult(resp.Body, resp.StatusCode(), "patent info", registeredNumber, "failed to get patent info")
}

// GetSearchableFullText retrieves the searchable full text for a document
func (c *Client) GetSearchableFullText(ctx context.Context, documentID string) ([]byte, error) {
	resp, err := c.generated.GetSearchableFullTextWithResponse(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get searchable full text: %w", err)
	}
	return resourceResult(resp.Body, resp.StatusCode(), "searchable full text", documentID, "failed to get searchable full text")
}

// GetDisclosureDocumentsXML downloads disclosure documents (A) as XML for a publication week
func (c *Client) GetDisclosureDocumentsXML(ctx context.Context, year, week int) ([]byte, error) {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return nil, err
	}
	resp, err := c.generated.GetDisclosureDocumentsXMLWithResponse(ctx, pubWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get disclosure documents: %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "failed to download disclosure documents")
}

// GetPatentSpecificationsXML downloads patent specifications (B, C) as XML for a publication week
func (c *Client) GetPatentSpecificationsXML(ctx context.Context, year, week int) ([]byte, error) {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return nil, err
	}
	resp, err := c.generated.GetPatentSpecificationsXMLWithResponse(ctx, pubWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get patent specifications: %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "failed to download patent specifications")
}

// GetUtilityModelsXML downloads utility models (U) as XML for a publication week
func (c *Client) GetUtilityModelsXML(ctx context.Context, year, week int) ([]byte, error) {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return nil, err
	}
	resp, err := c.generated.GetUtilityModelsXMLWithResponse(ctx, pubWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get utility models: %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "failed to download utility models")
}

// GetPublicationDataXML downloads publication data as XML for a publication week
func (c *Client) GetPublicationDataXML(ctx context.Context, year, week int) ([]byte, error) {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return nil, err
	}
	resp, err := c.generated.GetPublicationDataXMLWithResponse(ctx, pubWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get publication data: %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "failed to download publication data")
}

// GetApplicantCitationsXML downloads applicant citations as XML for a publication week
func (c *Client) GetApplicantCitationsXML(ctx context.Context, year, week int) ([]byte, error) {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return nil, err
	}
	resp, err := c.generated.GetApplicantCitationsXMLWithResponse(ctx, pubWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get applicant citations: %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "failed to download applicant citations")
}

// GetEuropeanPatentSpecificationsXML downloads European patent specifications as XML for a publication week
func (c *Client) GetEuropeanPatentSpecificationsXML(ctx context.Context, year, week int) ([]byte, error) {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return nil, err
	}
	resp, err := c.generated.GetEuropeanPatentSpecificationsXMLWithResponse(ctx, pubWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get European patent specifications: %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "failed to download European patent specifications")
}

// GetDisclosureDocumentsPDF downloads disclosure documents as PDF for a publication week
func (c *Client) GetDisclosureDocumentsPDF(ctx context.Context, year, week int) ([]byte, error) {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return nil, err
	}
	resp, err := c.generated.GetDisclosureDocumentsPDFWithResponse(ctx, pubWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get disclosure documents PDF: %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "failed to download disclosure documents PDF")
}

// GetPatentSpecificationsPDF downloads patent specifications as PDF for a publication week
func (c *Client) GetPatentSpecificationsPDF(ctx context.Context, year, week int) ([]byte, error) {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return nil, err
	}
	resp, err := c.generated.GetPatentSpecificationsPDFWithResponse(ctx, pubWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get patent specifications PDF: %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "failed to download patent specifications PDF")
}

// GetEuropeanPatentSpecificationsPDF downloads European patent specifications as PDF for a publication week
func (c *Client) GetEuropeanPatentSpecificationsPDF(ctx context.Context, year, week int) ([]byte, error) {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return nil, err
	}
	resp, err := c.generated.GetEuropeanPatentSpecificationsPDFWithResponse(ctx, pubWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get European patent specifications PDF: %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "failed to download European patent specifications PDF")
}

// GetUtilityModelsPDF downloads utility models as PDF for a publication week
func (c *Client) GetUtilityModelsPDF(ctx context.Context, year, week int) ([]byte, error) {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return nil, err
	}
	resp, err := c.generated.GetUtilityModelsPDFWithResponse(ctx, pubWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get utility models PDF: %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "failed to download utility models PDF")
}

// GetPatentRegisterExtract downloads patent register extract data for a date and period
func (c *Client) GetPatentRegisterExtract(ctx context.Context, date time.Time, period string) ([]byte, error) {
	if err := ValidatePeriod(period); err != nil {
		return nil, err
	}
	d := openapi_types.Date{Time: date}
	resp, err := c.generated.GetPatentRegisterExtractWithResponse(ctx, d, generated.GetPatentRegisterExtractParamsPeriod(period))
	if err != nil {
		return nil, fmt.Errorf("failed to get patent register extract: %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "failed to download patent register extract")
}

// GetDisclosureDocumentsXMLStream downloads disclosure documents as XML and writes to dst
func (c *Client) GetDisclosureDocumentsXMLStream(ctx context.Context, year, week int, dst io.Writer) error {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return err
	}
	resp, err := c.generated.GetDisclosureDocumentsXML(ctx, pubWeek)
	return streamResponse(resp, err, "failed to get disclosure documents", dst)
}

// GetPatentSpecificationsXMLStream downloads patent specifications as XML and writes to dst
func (c *Client) GetPatentSpecificationsXMLStream(ctx context.Context, year, week int, dst io.Writer) error {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return err
	}
	resp, err := c.generated.GetPatentSpecificationsXML(ctx, pubWeek)
	return streamResponse(resp, err, "failed to get patent specifications", dst)
}

// GetUtilityModelsXMLStream downloads utility models as XML and writes to dst
func (c *Client) GetUtilityModelsXMLStream(ctx context.Context, year, week int, dst io.Writer) error {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return err
	}
	resp, err := c.generated.GetUtilityModelsXML(ctx, pubWeek)
	return streamResponse(resp, err, "failed to get utility models", dst)
}

// GetPublicationDataXMLStream downloads publication data as XML and writes to dst
func (c *Client) GetPublicationDataXMLStream(ctx context.Context, year, week int, dst io.Writer) error {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return err
	}
	resp, err := c.generated.GetPublicationDataXML(ctx, pubWeek)
	return streamResponse(resp, err, "failed to get publication data", dst)
}

// GetApplicantCitationsXMLStream downloads applicant citations as XML and writes to dst
func (c *Client) GetApplicantCitationsXMLStream(ctx context.Context, year, week int, dst io.Writer) error {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return err
	}
	resp, err := c.generated.GetApplicantCitationsXML(ctx, pubWeek)
	return streamResponse(resp, err, "failed to get applicant citations", dst)
}

// GetEuropeanPatentSpecificationsXMLStream downloads European patent specifications as XML and writes to dst
func (c *Client) GetEuropeanPatentSpecificationsXMLStream(ctx context.Context, year, week int, dst io.Writer) error {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return err
	}
	resp, err := c.generated.GetEuropeanPatentSpecificationsXML(ctx, pubWeek)
	return streamResponse(resp, err, "failed to get European patent specifications", dst)
}

// GetDisclosureDocumentsPDFStream downloads disclosure documents as PDF and writes to dst
func (c *Client) GetDisclosureDocumentsPDFStream(ctx context.Context, year, week int, dst io.Writer) error {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return err
	}
	resp, err := c.generated.GetDisclosureDocumentsPDF(ctx, pubWeek)
	return streamResponse(resp, err, "failed to get disclosure documents PDF", dst)
}

// GetPatentSpecificationsPDFStream downloads patent specifications as PDF and writes to dst
func (c *Client) GetPatentSpecificationsPDFStream(ctx context.Context, year, week int, dst io.Writer) error {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return err
	}
	resp, err := c.generated.GetPatentSpecificationsPDF(ctx, pubWeek)
	return streamResponse(resp, err, "failed to get patent specifications PDF", dst)
}

// GetEuropeanPatentSpecificationsPDFStream downloads European patent specifications as PDF and writes to dst
func (c *Client) GetEuropeanPatentSpecificationsPDFStream(ctx context.Context, year, week int, dst io.Writer) error {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return err
	}
	resp, err := c.generated.GetEuropeanPatentSpecificationsPDF(ctx, pubWeek)
	return streamResponse(resp, err, "failed to get European patent specifications PDF", dst)
}

// GetUtilityModelsPDFStream downloads utility models as PDF and writes to dst
func (c *Client) GetUtilityModelsPDFStream(ctx context.Context, year, week int, dst io.Writer) error {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return err
	}
	resp, err := c.generated.GetUtilityModelsPDF(ctx, pubWeek)
	return streamResponse(resp, err, "failed to get utility models PDF", dst)
}

// SearchPatentsParsed executes a patent search and returns parsed results.
func (c *Client) SearchPatentsParsed(ctx context.Context, query string) (*PatentSearchResult, error) {
	data, err := c.SearchPatents(ctx, query)
	if err != nil {
		return nil, err
	}
	return ParsePatentSearch(data)
}

// GetPatentInfoParsed retrieves patent info and returns parsed bibliographic data.
// Accepts either a bare registered number (e.g., "100273629") or a DE patent number
// with country prefix and/or kind code (e.g., "DE10027362C2", "DE102019200907A1").
// For non-registered numbers, it resolves via publication number search automatically.
func (c *Client) GetPatentInfoParsed(ctx context.Context, patentNumber string) (*PatentInfo, error) {
	patentNumber = strings.TrimSpace(patentNumber)
	if isRegisteredNumber(patentNumber) {
		data, err := c.GetPatentInfo(ctx, patentNumber)
		if err != nil {
			return nil, err
		}
		return ParsePatentInfo(data)
	}
	// Not a bare registered number - resolve via publication number search
	return c.GetPatentInfoByPublicationNumber(ctx, patentNumber)
}

// GetPatentInfoByPublicationNumber resolves a DE publication number (e.g. "DE102019200907A1")
// to a registered number via search and returns the parsed patent info.
func (c *Client) GetPatentInfoByPublicationNumber(ctx context.Context, publicationNumber string) (*PatentInfo, error) {
	query := fmt.Sprintf("PN=%s", publicationNumber)
	searchResult, err := c.SearchPatentsParsed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve publication number %s: %w", publicationNumber, err)
	}
	if len(searchResult.Hits) == 0 {
		return nil, &NotFoundError{Resource: "patent", ID: publicationNumber}
	}
	regNum := searchResult.Hits[0].LeadingRegisteredNumber
	if regNum == "" {
		return nil, fmt.Errorf("patent %s has no leading-registered-number", publicationNumber)
	}
	return c.GetPatentInfoParsed(ctx, regNum)
}

// GetPatentRegisterExtractStream downloads patent register extract data and writes to dst
func (c *Client) GetPatentRegisterExtractStream(ctx context.Context, date time.Time, period string, dst io.Writer) error {
	if err := ValidatePeriod(period); err != nil {
		return err
	}
	d := openapi_types.Date{Time: date}
	resp, err := c.generated.GetPatentRegisterExtract(ctx, d, generated.GetPatentRegisterExtractParamsPeriod(period))
	return streamResponse(resp, err, "failed to get patent register extract", dst)
}
