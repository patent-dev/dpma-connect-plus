package dpmaconnect

import (
	"context"
	"fmt"
	"io"
	"net/http"
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
	return fetchWeeklyBulk(year, week,
		"failed to get disclosure documents", "failed to download disclosure documents",
		func(pw string) (*generated.GetDisclosureDocumentsXMLResponse, error) {
			return c.generated.GetDisclosureDocumentsXMLWithResponse(ctx, pw)
		},
		func(r *generated.GetDisclosureDocumentsXMLResponse) []byte { return r.Body })
}

// GetPatentSpecificationsXML downloads patent specifications (B, C) as XML for a publication week
func (c *Client) GetPatentSpecificationsXML(ctx context.Context, year, week int) ([]byte, error) {
	return fetchWeeklyBulk(year, week,
		"failed to get patent specifications", "failed to download patent specifications",
		func(pw string) (*generated.GetPatentSpecificationsXMLResponse, error) {
			return c.generated.GetPatentSpecificationsXMLWithResponse(ctx, pw)
		},
		func(r *generated.GetPatentSpecificationsXMLResponse) []byte { return r.Body })
}

// GetUtilityModelsXML downloads utility models (U) as XML for a publication week
func (c *Client) GetUtilityModelsXML(ctx context.Context, year, week int) ([]byte, error) {
	return fetchWeeklyBulk(year, week,
		"failed to get utility models", "failed to download utility models",
		func(pw string) (*generated.GetUtilityModelsXMLResponse, error) {
			return c.generated.GetUtilityModelsXMLWithResponse(ctx, pw)
		},
		func(r *generated.GetUtilityModelsXMLResponse) []byte { return r.Body })
}

// GetPublicationDataXML downloads publication data as XML for a publication week
func (c *Client) GetPublicationDataXML(ctx context.Context, year, week int) ([]byte, error) {
	return fetchWeeklyBulk(year, week,
		"failed to get publication data", "failed to download publication data",
		func(pw string) (*generated.GetPublicationDataXMLResponse, error) {
			return c.generated.GetPublicationDataXMLWithResponse(ctx, pw)
		},
		func(r *generated.GetPublicationDataXMLResponse) []byte { return r.Body })
}

// GetApplicantCitationsXML downloads applicant citations as XML for a publication week
func (c *Client) GetApplicantCitationsXML(ctx context.Context, year, week int) ([]byte, error) {
	return fetchWeeklyBulk(year, week,
		"failed to get applicant citations", "failed to download applicant citations",
		func(pw string) (*generated.GetApplicantCitationsXMLResponse, error) {
			return c.generated.GetApplicantCitationsXMLWithResponse(ctx, pw)
		},
		func(r *generated.GetApplicantCitationsXMLResponse) []byte { return r.Body })
}

// GetEuropeanPatentSpecificationsXML downloads European patent specifications as XML for a publication week
func (c *Client) GetEuropeanPatentSpecificationsXML(ctx context.Context, year, week int) ([]byte, error) {
	return fetchWeeklyBulk(year, week,
		"failed to get European patent specifications", "failed to download European patent specifications",
		func(pw string) (*generated.GetEuropeanPatentSpecificationsXMLResponse, error) {
			return c.generated.GetEuropeanPatentSpecificationsXMLWithResponse(ctx, pw)
		},
		func(r *generated.GetEuropeanPatentSpecificationsXMLResponse) []byte { return r.Body })
}

// GetDisclosureDocumentsPDF downloads disclosure documents as PDF for a publication week
func (c *Client) GetDisclosureDocumentsPDF(ctx context.Context, year, week int) ([]byte, error) {
	return fetchWeeklyBulk(year, week,
		"failed to get disclosure documents PDF", "failed to download disclosure documents PDF",
		func(pw string) (*generated.GetDisclosureDocumentsPDFResponse, error) {
			return c.generated.GetDisclosureDocumentsPDFWithResponse(ctx, pw)
		},
		func(r *generated.GetDisclosureDocumentsPDFResponse) []byte { return r.Body })
}

// GetPatentSpecificationsPDF downloads patent specifications as PDF for a publication week
func (c *Client) GetPatentSpecificationsPDF(ctx context.Context, year, week int) ([]byte, error) {
	return fetchWeeklyBulk(year, week,
		"failed to get patent specifications PDF", "failed to download patent specifications PDF",
		func(pw string) (*generated.GetPatentSpecificationsPDFResponse, error) {
			return c.generated.GetPatentSpecificationsPDFWithResponse(ctx, pw)
		},
		func(r *generated.GetPatentSpecificationsPDFResponse) []byte { return r.Body })
}

// GetEuropeanPatentSpecificationsPDF downloads European patent specifications as PDF for a publication week
func (c *Client) GetEuropeanPatentSpecificationsPDF(ctx context.Context, year, week int) ([]byte, error) {
	return fetchWeeklyBulk(year, week,
		"failed to get European patent specifications PDF", "failed to download European patent specifications PDF",
		func(pw string) (*generated.GetEuropeanPatentSpecificationsPDFResponse, error) {
			return c.generated.GetEuropeanPatentSpecificationsPDFWithResponse(ctx, pw)
		},
		func(r *generated.GetEuropeanPatentSpecificationsPDFResponse) []byte { return r.Body })
}

// GetUtilityModelsPDF downloads utility models as PDF for a publication week
func (c *Client) GetUtilityModelsPDF(ctx context.Context, year, week int) ([]byte, error) {
	return fetchWeeklyBulk(year, week,
		"failed to get utility models PDF", "failed to download utility models PDF",
		func(pw string) (*generated.GetUtilityModelsPDFResponse, error) {
			return c.generated.GetUtilityModelsPDFWithResponse(ctx, pw)
		},
		func(r *generated.GetUtilityModelsPDFResponse) []byte { return r.Body })
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
	return streamWeekly(year, week, "failed to get disclosure documents", dst,
		func(pw string) (*http.Response, error) { return c.generated.GetDisclosureDocumentsXML(ctx, pw) })
}

// GetPatentSpecificationsXMLStream downloads patent specifications as XML and writes to dst
func (c *Client) GetPatentSpecificationsXMLStream(ctx context.Context, year, week int, dst io.Writer) error {
	return streamWeekly(year, week, "failed to get patent specifications", dst,
		func(pw string) (*http.Response, error) { return c.generated.GetPatentSpecificationsXML(ctx, pw) })
}

// GetUtilityModelsXMLStream downloads utility models as XML and writes to dst
func (c *Client) GetUtilityModelsXMLStream(ctx context.Context, year, week int, dst io.Writer) error {
	return streamWeekly(year, week, "failed to get utility models", dst,
		func(pw string) (*http.Response, error) { return c.generated.GetUtilityModelsXML(ctx, pw) })
}

// GetPublicationDataXMLStream downloads publication data as XML and writes to dst
func (c *Client) GetPublicationDataXMLStream(ctx context.Context, year, week int, dst io.Writer) error {
	return streamWeekly(year, week, "failed to get publication data", dst,
		func(pw string) (*http.Response, error) { return c.generated.GetPublicationDataXML(ctx, pw) })
}

// GetApplicantCitationsXMLStream downloads applicant citations as XML and writes to dst
func (c *Client) GetApplicantCitationsXMLStream(ctx context.Context, year, week int, dst io.Writer) error {
	return streamWeekly(year, week, "failed to get applicant citations", dst,
		func(pw string) (*http.Response, error) { return c.generated.GetApplicantCitationsXML(ctx, pw) })
}

// GetEuropeanPatentSpecificationsXMLStream downloads European patent specifications as XML and writes to dst
func (c *Client) GetEuropeanPatentSpecificationsXMLStream(ctx context.Context, year, week int, dst io.Writer) error {
	return streamWeekly(year, week, "failed to get European patent specifications", dst,
		func(pw string) (*http.Response, error) {
			return c.generated.GetEuropeanPatentSpecificationsXML(ctx, pw)
		})
}

// GetDisclosureDocumentsPDFStream downloads disclosure documents as PDF and writes to dst
func (c *Client) GetDisclosureDocumentsPDFStream(ctx context.Context, year, week int, dst io.Writer) error {
	return streamWeekly(year, week, "failed to get disclosure documents PDF", dst,
		func(pw string) (*http.Response, error) { return c.generated.GetDisclosureDocumentsPDF(ctx, pw) })
}

// GetPatentSpecificationsPDFStream downloads patent specifications as PDF and writes to dst
func (c *Client) GetPatentSpecificationsPDFStream(ctx context.Context, year, week int, dst io.Writer) error {
	return streamWeekly(year, week, "failed to get patent specifications PDF", dst,
		func(pw string) (*http.Response, error) { return c.generated.GetPatentSpecificationsPDF(ctx, pw) })
}

// GetEuropeanPatentSpecificationsPDFStream downloads European patent specifications as PDF and writes to dst
func (c *Client) GetEuropeanPatentSpecificationsPDFStream(ctx context.Context, year, week int, dst io.Writer) error {
	return streamWeekly(year, week, "failed to get European patent specifications PDF", dst,
		func(pw string) (*http.Response, error) {
			return c.generated.GetEuropeanPatentSpecificationsPDF(ctx, pw)
		})
}

// GetUtilityModelsPDFStream downloads utility models as PDF and writes to dst
func (c *Client) GetUtilityModelsPDFStream(ctx context.Context, year, week int, dst io.Writer) error {
	return streamWeekly(year, week, "failed to get utility models PDF", dst,
		func(pw string) (*http.Response, error) { return c.generated.GetUtilityModelsPDF(ctx, pw) })
}

// SearchPatentsParsed executes a patent search and returns parsed results.
func (c *Client) SearchPatentsParsed(ctx context.Context, query string) (*PatentSearchResult, error) {
	data, err := c.SearchPatents(ctx, query)
	if err != nil {
		return nil, err
	}
	return ParsePatentSearch(data)
}

// maxPatentResolutionDepth caps the mutual recursion between getPatentInfoParsed
// and getPatentInfoByPublicationNumber. The only legitimate chain is one search
// (publication number -> registered number) followed by one direct fetch; if a
// second search still yields a non-registered number, the register data points
// back at itself and resolution must stop.
const maxPatentResolutionDepth = 2

// GetPatentInfoParsed retrieves patent info and returns parsed bibliographic data.
// Accepts either a bare registered number (e.g., "100273629") or a DE patent number
// with country prefix and/or kind code (e.g., "DE10027362C2", "DE102019200907A1").
// For non-registered numbers, it resolves via publication number search automatically.
// A *ResolutionLoopError is returned when the register data keeps pointing at a
// non-registered number instead of terminating.
func (c *Client) GetPatentInfoParsed(ctx context.Context, patentNumber string) (*PatentInfo, error) {
	return c.getPatentInfoParsed(ctx, patentNumber, 0)
}

func (c *Client) getPatentInfoParsed(ctx context.Context, patentNumber string, depth int) (*PatentInfo, error) {
	patentNumber = strings.TrimSpace(patentNumber)
	if isRegisteredNumber(patentNumber) {
		data, err := c.GetPatentInfo(ctx, patentNumber)
		if err != nil {
			return nil, err
		}
		return ParsePatentInfo(data)
	}
	if depth >= maxPatentResolutionDepth {
		return nil, &ResolutionLoopError{Number: patentNumber}
	}
	// Not a bare registered number - resolve via publication number search
	return c.getPatentInfoByPublicationNumber(ctx, patentNumber, depth)
}

// GetPatentInfoByPublicationNumber resolves a DE publication number (e.g. "DE102019200907A1")
// to a registered number via search and returns the parsed patent info.
func (c *Client) GetPatentInfoByPublicationNumber(ctx context.Context, publicationNumber string) (*PatentInfo, error) {
	return c.getPatentInfoByPublicationNumber(ctx, publicationNumber, 0)
}

func (c *Client) getPatentInfoByPublicationNumber(ctx context.Context, publicationNumber string, depth int) (*PatentInfo, error) {
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
	return c.getPatentInfoParsed(ctx, regNum, depth+1)
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
