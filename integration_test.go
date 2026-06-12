//go:build integration

package dpmaconnect

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

// errBoundReached signals that boundedWriter has captured its limit and the
// caller should stop the stream; reachBulk treats it as a clean success.
var errBoundReached = errors.New("bounded writer limit reached")

// boundedWriter captures up to limit bytes then returns errBoundReached, so a
// bulk stream is proven reachable without downloading a potentially huge archive.
type boundedWriter struct {
	buf   bytes.Buffer
	limit int
}

func (w *boundedWriter) Write(p []byte) (int, error) {
	remaining := w.limit - w.buf.Len()
	if remaining <= 0 {
		return 0, errBoundReached
	}
	if len(p) > remaining {
		w.buf.Write(p[:remaining])
		return len(p), errBoundReached
	}
	return w.buf.Write(p)
}

// integration_test.go holds one TestIntegration<Method> per exported Client
// method (see scripts/check-integration-coverage.sh, which enforces the 1:1
// mapping). Every test RUNS live against the DPMA API when credentials are in the
// environment and otherwise SKIPS cleanly; none ever FAILs on a documented
// condition. Inputs reuse the demo-proven parameters in demo/examples.
//
// Documented SKIP conditions (account/data limits, not client bugs):
//   - no credentials                       -> getTestClient skips.
//   - "No permission of user=sip ..."      -> skipIfNoPermission (test account
//     lacks the bulk-data entitlement).
//   - HTTP 401/403/404/429                  -> skipIfTransient (auth/quota/missing
//     resource for the chosen demo input).
//   - *DataNotAvailableError                -> skipIfDataNotAvailable (bulk slot
//     not published for the chosen year/week/date).
//   - *NotFoundError                        -> skipped per-test where a demo id may
//     have aged out.
//
// Bulk PDF/XML downloads can be very large; the bulk endpoints are bounded by a
// short request timeout and treated as a successful reach if the server starts
// streaming (any DataNotAvailable / permission response is a clean skip).

// ---- shared inputs (mirror demo/examples/*/request.txt) --------------------

const (
	itPatentRegisteredNumber = "100273602"        // demo get_patent_info
	itPatentPublicationID    = "DE102019200907A1" // demo get_patent_publication_pdf / searchable_full_text
	itPatentSearchQuery      = "TI=Elektrofahrzeug"

	itDesignNumber      = "M8800174-0001" // demo get_design_info / image / thumbnail
	itDesignSearchQuery = "INH=Samsung"

	itTrademarkNumber      = "919224" // demo get_trademark_info
	itTrademarkSearchQuery = "md=Apple"

	itBulkYear = 2026
	itBulkWeek = 9
)

func itBulkDate() time.Time { return time.Date(2024, 10, 23, 0, 0, 0, 0, time.UTC) }

// ---- skip helpers ----------------------------------------------------------

func getTestClient(t *testing.T) *Client {
	t.Helper()
	username := os.Getenv("DPMA_CONNECT_PLUS_USERNAME")
	password := os.Getenv("DPMA_CONNECT_PLUS_PASSWORD")
	if username == "" || password == "" {
		t.Skip("Skipping integration test: DPMA_CONNECT_PLUS_USERNAME and DPMA_CONNECT_PLUS_PASSWORD must be set")
	}
	config := DefaultConfig()
	config.Username = username
	config.Password = password
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	return client
}

// skipIfNoPermission turns the known "No permission of user=sip for right=..."
// entitlement limit into a test skip. The DPMA test account does not have access
// to every bulk endpoint; that is an account limitation, not a client bug.
func skipIfNoPermission(t *testing.T, err error) {
	t.Helper()
	var apiErr *APIError
	if errors.As(err, &apiErr) && strings.Contains(apiErr.Message, "No permission of user") {
		t.Skipf("account lacks entitlement: %v", err)
	}
}

// skipIfDataNotAvailable turns a bulk slot that is not published into a skip.
func skipIfDataNotAvailable(t *testing.T, err error) {
	t.Helper()
	var dna *DataNotAvailableError
	if errors.As(err, &dna) {
		t.Skipf("data not available for the chosen slot: %v", err)
	}
}

// skipIfTransient turns documented HTTP conditions (401/403 auth, 404 missing
// resource for the demo input, 429 quota) into a skip so the suite reports
// PASS-or-SKIP, never FAIL, on account/data limits outside the client's control.
func skipIfTransient(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		return
	}
	skipIfNoPermission(t, err)
	skipIfDataNotAvailable(t, err)
	var nf *NotFoundError
	if errors.As(err, &nf) {
		t.Skipf("resource not found for demo input: %v", err)
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case 401, 403, 404, 429:
			t.Skipf("transient/account HTTP %d: %v", apiErr.StatusCode, err)
		}
	}
}

// assertXMLWellFormed asserts the bytes are non-empty, well-formed XML.
func assertXMLWellFormed(t *testing.T, data []byte) {
	t.Helper()
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		_, err := dec.Token()
		if err != nil {
			if err.Error() == "EOF" {
				return
			}
			t.Fatalf("not well-formed XML: %v", err)
		}
	}
}

// reachBulk reads up to a small bound from a bulk stream and returns the bytes
// captured plus the streaming error (nil on EOF). It proves the endpoint was
// reached without downloading a potentially huge archive.
func reachBulk(t *testing.T, fn func(dst *boundedWriter) error) ([]byte, error) {
	t.Helper()
	w := &boundedWriter{limit: 64 * 1024}
	err := fn(w)
	if errors.Is(err, errBoundReached) {
		err = nil // we deliberately stopped the stream after the bound
	}
	return w.buf.Bytes(), err
}

// ---- version ---------------------------------------------------------------

func TestIntegrationGetVersion(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	for _, service := range []string{ServicePatent, ServiceDesign, ServiceTrademark} {
		t.Run(service, func(t *testing.T) {
			version, err := client.GetVersion(ctx, service)
			if err != nil {
				skipIfTransient(t, err)
				t.Fatalf("GetVersion(%s) error = %v", service, err)
			}
			if version == "" {
				t.Errorf("GetVersion(%s) returned empty version", service)
			}
			t.Logf("GetVersion(%s) = %s", service, version)
		})
	}
}

// ---- patent: search / info / parsed ----------------------------------------

func TestIntegrationSearchPatents(t *testing.T) {
	client := getTestClient(t)
	data, err := client.SearchPatents(context.Background(), itPatentSearchQuery)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("SearchPatents error = %v", err)
	}
	assertXMLWellFormed(t, data)
	t.Logf("SearchPatents returned %d bytes", len(data))
}

func TestIntegrationSearchPatentsParsed(t *testing.T) {
	client := getTestClient(t)
	result, err := client.SearchPatentsParsed(context.Background(), itPatentSearchQuery)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("SearchPatentsParsed error = %v", err)
	}
	if result.TotalHits == 0 || len(result.Hits) == 0 {
		t.Fatalf("no hits: TotalHits=%d len(Hits)=%d", result.TotalHits, len(result.Hits))
	}
	h := result.Hits[0]
	if h.LeadingRegisteredNumber == "" {
		t.Error("Hits[0].LeadingRegisteredNumber is empty")
	}
	if h.Title == "" {
		t.Error("Hits[0].Title is empty")
	}
	t.Logf("SearchPatentsParsed: %d hits, first %s - %s", result.TotalHits, h.LeadingRegisteredNumber, h.Title)
}

func TestIntegrationGetPatentInfo(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetPatentInfo(context.Background(), itPatentRegisteredNumber)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetPatentInfo error = %v", err)
	}
	assertXMLWellFormed(t, data)
	t.Logf("GetPatentInfo returned %d bytes", len(data))
}

func TestIntegrationGetPatentInfoParsed(t *testing.T) {
	client := getTestClient(t)
	result, err := client.GetPatentInfoParsed(context.Background(), itPatentRegisteredNumber)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetPatentInfoParsed error = %v", err)
	}
	if result.IPRightType == "" {
		t.Error("IPRightType is empty")
	}
	if result.Title == "" {
		t.Error("Title is empty")
	}
	if len(result.Applicants) == 0 {
		t.Error("no applicants")
	}
	t.Logf("GetPatentInfoParsed: %s (%s) - %s", result.ApplicationRef.Number, result.IPRightType, result.Title)
}

func TestIntegrationGetPatentInfoByPublicationNumber(t *testing.T) {
	client := getTestClient(t)
	result, err := client.GetPatentInfoByPublicationNumber(context.Background(), itPatentPublicationID)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetPatentInfoByPublicationNumber error = %v", err)
	}
	if result.Title == "" {
		t.Error("Title is empty")
	}
	t.Logf("GetPatentInfoByPublicationNumber: %s - %s", result.ApplicationRef.Number, result.Title)
}

func TestIntegrationGetPatentPublicationPDF(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetPatentPublicationPDF(context.Background(), itPatentPublicationID)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetPatentPublicationPDF error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty PDF")
	}
	if len(data) >= 4 && string(data[:4]) != "%PDF" {
		t.Errorf("not a PDF (magic %q)", data[:4])
	}
	t.Logf("GetPatentPublicationPDF returned %d bytes", len(data))
}

func TestIntegrationGetSearchableFullText(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetSearchableFullText(context.Background(), itPatentPublicationID)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetSearchableFullText error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetSearchableFullText returned %d bytes", len(data))
}

// ---- patent: bulk XML ------------------------------------------------------

func TestIntegrationGetDisclosureDocumentsXML(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetDisclosureDocumentsXML(context.Background(), itBulkYear, itBulkWeek)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetDisclosureDocumentsXML error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetDisclosureDocumentsXML returned %d bytes", len(data))
}

func TestIntegrationGetPatentSpecificationsXML(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetPatentSpecificationsXML(context.Background(), itBulkYear, itBulkWeek)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetPatentSpecificationsXML error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetPatentSpecificationsXML returned %d bytes", len(data))
}

func TestIntegrationGetUtilityModelsXML(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetUtilityModelsXML(context.Background(), itBulkYear, itBulkWeek)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetUtilityModelsXML error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetUtilityModelsXML returned %d bytes", len(data))
}

func TestIntegrationGetPublicationDataXML(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetPublicationDataXML(context.Background(), itBulkYear, itBulkWeek)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetPublicationDataXML error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetPublicationDataXML returned %d bytes", len(data))
}

func TestIntegrationGetApplicantCitationsXML(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetApplicantCitationsXML(context.Background(), itBulkYear, itBulkWeek)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetApplicantCitationsXML error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetApplicantCitationsXML returned %d bytes", len(data))
}

func TestIntegrationGetEuropeanPatentSpecificationsXML(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetEuropeanPatentSpecificationsXML(context.Background(), itBulkYear, itBulkWeek)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetEuropeanPatentSpecificationsXML error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetEuropeanPatentSpecificationsXML returned %d bytes", len(data))
}

// ---- patent: bulk PDF ------------------------------------------------------

func TestIntegrationGetDisclosureDocumentsPDF(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetDisclosureDocumentsPDF(context.Background(), itBulkYear, itBulkWeek)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetDisclosureDocumentsPDF error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetDisclosureDocumentsPDF returned %d bytes", len(data))
}

func TestIntegrationGetPatentSpecificationsPDF(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetPatentSpecificationsPDF(context.Background(), itBulkYear, itBulkWeek)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetPatentSpecificationsPDF error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetPatentSpecificationsPDF returned %d bytes", len(data))
}

func TestIntegrationGetEuropeanPatentSpecificationsPDF(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetEuropeanPatentSpecificationsPDF(context.Background(), itBulkYear, itBulkWeek)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetEuropeanPatentSpecificationsPDF error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetEuropeanPatentSpecificationsPDF returned %d bytes", len(data))
}

func TestIntegrationGetUtilityModelsPDF(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetUtilityModelsPDF(context.Background(), itBulkYear, itBulkWeek)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetUtilityModelsPDF error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetUtilityModelsPDF returned %d bytes", len(data))
}

func TestIntegrationGetPatentRegisterExtract(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetPatentRegisterExtract(context.Background(), itBulkDate(), PeriodDaily)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetPatentRegisterExtract error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetPatentRegisterExtract returned %d bytes", len(data))
}

// ---- patent: streams -------------------------------------------------------

func TestIntegrationGetDisclosureDocumentsXMLStream(t *testing.T) {
	client := getTestClient(t)
	data, err := reachBulk(t, func(w *boundedWriter) error {
		return client.GetDisclosureDocumentsXMLStream(context.Background(), itBulkYear, itBulkWeek, w)
	})
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetDisclosureDocumentsXMLStream error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stream produced no bytes")
	}
	t.Logf("GetDisclosureDocumentsXMLStream read %d bytes", len(data))
}

func TestIntegrationGetPatentSpecificationsXMLStream(t *testing.T) {
	client := getTestClient(t)
	data, err := reachBulk(t, func(w *boundedWriter) error {
		return client.GetPatentSpecificationsXMLStream(context.Background(), itBulkYear, itBulkWeek, w)
	})
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetPatentSpecificationsXMLStream error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stream produced no bytes")
	}
	t.Logf("GetPatentSpecificationsXMLStream read %d bytes", len(data))
}

func TestIntegrationGetUtilityModelsXMLStream(t *testing.T) {
	client := getTestClient(t)
	data, err := reachBulk(t, func(w *boundedWriter) error {
		return client.GetUtilityModelsXMLStream(context.Background(), itBulkYear, itBulkWeek, w)
	})
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetUtilityModelsXMLStream error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stream produced no bytes")
	}
	t.Logf("GetUtilityModelsXMLStream read %d bytes", len(data))
}

func TestIntegrationGetPublicationDataXMLStream(t *testing.T) {
	client := getTestClient(t)
	data, err := reachBulk(t, func(w *boundedWriter) error {
		return client.GetPublicationDataXMLStream(context.Background(), itBulkYear, itBulkWeek, w)
	})
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetPublicationDataXMLStream error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stream produced no bytes")
	}
	t.Logf("GetPublicationDataXMLStream read %d bytes", len(data))
}

func TestIntegrationGetApplicantCitationsXMLStream(t *testing.T) {
	client := getTestClient(t)
	data, err := reachBulk(t, func(w *boundedWriter) error {
		return client.GetApplicantCitationsXMLStream(context.Background(), itBulkYear, itBulkWeek, w)
	})
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetApplicantCitationsXMLStream error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stream produced no bytes")
	}
	t.Logf("GetApplicantCitationsXMLStream read %d bytes", len(data))
}

func TestIntegrationGetEuropeanPatentSpecificationsXMLStream(t *testing.T) {
	client := getTestClient(t)
	data, err := reachBulk(t, func(w *boundedWriter) error {
		return client.GetEuropeanPatentSpecificationsXMLStream(context.Background(), itBulkYear, itBulkWeek, w)
	})
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetEuropeanPatentSpecificationsXMLStream error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stream produced no bytes")
	}
	t.Logf("GetEuropeanPatentSpecificationsXMLStream read %d bytes", len(data))
}

func TestIntegrationGetDisclosureDocumentsPDFStream(t *testing.T) {
	client := getTestClient(t)
	data, err := reachBulk(t, func(w *boundedWriter) error {
		return client.GetDisclosureDocumentsPDFStream(context.Background(), itBulkYear, itBulkWeek, w)
	})
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetDisclosureDocumentsPDFStream error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stream produced no bytes")
	}
	t.Logf("GetDisclosureDocumentsPDFStream read %d bytes", len(data))
}

func TestIntegrationGetPatentSpecificationsPDFStream(t *testing.T) {
	client := getTestClient(t)
	data, err := reachBulk(t, func(w *boundedWriter) error {
		return client.GetPatentSpecificationsPDFStream(context.Background(), itBulkYear, itBulkWeek, w)
	})
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetPatentSpecificationsPDFStream error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stream produced no bytes")
	}
	t.Logf("GetPatentSpecificationsPDFStream read %d bytes", len(data))
}

func TestIntegrationGetEuropeanPatentSpecificationsPDFStream(t *testing.T) {
	client := getTestClient(t)
	data, err := reachBulk(t, func(w *boundedWriter) error {
		return client.GetEuropeanPatentSpecificationsPDFStream(context.Background(), itBulkYear, itBulkWeek, w)
	})
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetEuropeanPatentSpecificationsPDFStream error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stream produced no bytes")
	}
	t.Logf("GetEuropeanPatentSpecificationsPDFStream read %d bytes", len(data))
}

func TestIntegrationGetUtilityModelsPDFStream(t *testing.T) {
	client := getTestClient(t)
	data, err := reachBulk(t, func(w *boundedWriter) error {
		return client.GetUtilityModelsPDFStream(context.Background(), itBulkYear, itBulkWeek, w)
	})
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetUtilityModelsPDFStream error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stream produced no bytes")
	}
	t.Logf("GetUtilityModelsPDFStream read %d bytes", len(data))
}

func TestIntegrationGetPatentRegisterExtractStream(t *testing.T) {
	client := getTestClient(t)
	data, err := reachBulk(t, func(w *boundedWriter) error {
		return client.GetPatentRegisterExtractStream(context.Background(), itBulkDate(), PeriodDaily, w)
	})
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetPatentRegisterExtractStream error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stream produced no bytes")
	}
	t.Logf("GetPatentRegisterExtractStream read %d bytes", len(data))
}

// ---- design ----------------------------------------------------------------

func TestIntegrationSearchDesigns(t *testing.T) {
	client := getTestClient(t)
	data, err := client.SearchDesigns(context.Background(), itDesignSearchQuery)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("SearchDesigns error = %v", err)
	}
	assertXMLWellFormed(t, data)
	t.Logf("SearchDesigns returned %d bytes", len(data))
}

func TestIntegrationSearchDesignsParsed(t *testing.T) {
	client := getTestClient(t)
	result, err := client.SearchDesignsParsed(context.Background(), itDesignSearchQuery)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("SearchDesignsParsed error = %v", err)
	}
	if result.TotalHits == 0 || len(result.Hits) == 0 {
		t.Fatalf("no hits: TotalHits=%d len(Hits)=%d", result.TotalHits, len(result.Hits))
	}
	t.Logf("SearchDesignsParsed: %d hits, first %s - %s", result.TotalHits, result.Hits[0].DesignIdentifier, result.Hits[0].Title)
}

func TestIntegrationGetDesignInfo(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetDesignInfo(context.Background(), itDesignNumber)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetDesignInfo error = %v", err)
	}
	assertXMLWellFormed(t, data)
	t.Logf("GetDesignInfo returned %d bytes", len(data))
}

func TestIntegrationGetDesignInfoParsed(t *testing.T) {
	client := getTestClient(t)
	result, err := client.GetDesignInfoParsed(context.Background(), itDesignNumber)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetDesignInfoParsed error = %v", err)
	}
	if result.DesignIdentifier != itDesignNumber {
		t.Errorf("DesignIdentifier = %q, want %q", result.DesignIdentifier, itDesignNumber)
	}
	if result.Title == "" {
		t.Error("Title is empty")
	}
	t.Logf("GetDesignInfoParsed: %s - %s (%s)", result.DesignIdentifier, result.Title, result.Status)
}

func TestIntegrationGetDesignImage(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetDesignImage(context.Background(), itDesignNumber, "1")
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetDesignImage error = %v", err)
	}
	assertImageBytes(t, "GetDesignImage", data)
}

func TestIntegrationGetDesignThumbnail(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetDesignThumbnail(context.Background(), itDesignNumber, "1")
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetDesignThumbnail error = %v", err)
	}
	assertImageBytes(t, "GetDesignThumbnail", data)
}

func TestIntegrationGetDesignBibliographicDataXML(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetDesignBibliographicDataXML(context.Background(), itBulkYear, itBulkWeek)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetDesignBibliographicDataXML error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetDesignBibliographicDataXML returned %d bytes", len(data))
}

func TestIntegrationGetDesignImages(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetDesignImages(context.Background(), itBulkYear, itBulkWeek)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetDesignImages error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetDesignImages returned %d bytes", len(data))
}

func TestIntegrationGetDesignRegisterExtract(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetDesignRegisterExtract(context.Background(), itBulkDate(), PeriodDaily)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetDesignRegisterExtract error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetDesignRegisterExtract returned %d bytes", len(data))
}

func TestIntegrationGetDesignBibliographicDataXMLStream(t *testing.T) {
	client := getTestClient(t)
	data, err := reachBulk(t, func(w *boundedWriter) error {
		return client.GetDesignBibliographicDataXMLStream(context.Background(), itBulkYear, itBulkWeek, w)
	})
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetDesignBibliographicDataXMLStream error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stream produced no bytes")
	}
	t.Logf("GetDesignBibliographicDataXMLStream read %d bytes", len(data))
}

func TestIntegrationGetDesignImagesStream(t *testing.T) {
	client := getTestClient(t)
	data, err := reachBulk(t, func(w *boundedWriter) error {
		return client.GetDesignImagesStream(context.Background(), itBulkYear, itBulkWeek, w)
	})
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetDesignImagesStream error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stream produced no bytes")
	}
	t.Logf("GetDesignImagesStream read %d bytes", len(data))
}

func TestIntegrationGetDesignRegisterExtractStream(t *testing.T) {
	client := getTestClient(t)
	data, err := reachBulk(t, func(w *boundedWriter) error {
		return client.GetDesignRegisterExtractStream(context.Background(), itBulkDate(), PeriodDaily, w)
	})
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetDesignRegisterExtractStream error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stream produced no bytes")
	}
	t.Logf("GetDesignRegisterExtractStream read %d bytes", len(data))
}

// ---- trademark -------------------------------------------------------------

func TestIntegrationSearchTrademarks(t *testing.T) {
	client := getTestClient(t)
	data, err := client.SearchTrademarks(context.Background(), itTrademarkSearchQuery)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("SearchTrademarks error = %v", err)
	}
	assertXMLWellFormed(t, data)
	t.Logf("SearchTrademarks returned %d bytes", len(data))
}

func TestIntegrationSearchTrademarksParsed(t *testing.T) {
	client := getTestClient(t)
	result, err := client.SearchTrademarksParsed(context.Background(), itTrademarkSearchQuery)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("SearchTrademarksParsed error = %v", err)
	}
	if result.TotalHits == 0 || len(result.Hits) == 0 {
		t.Fatalf("no hits: TotalHits=%d len(Hits)=%d", result.TotalHits, len(result.Hits))
	}
	t.Logf("SearchTrademarksParsed: %d hits, first %s - %s", result.TotalHits, result.Hits[0].ApplicationNumber, result.Hits[0].MarkText)
}

func TestIntegrationGetTrademarkInfo(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetTrademarkInfo(context.Background(), itTrademarkNumber)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetTrademarkInfo error = %v", err)
	}
	assertXMLWellFormed(t, data)
	t.Logf("GetTrademarkInfo returned %d bytes", len(data))
}

func TestIntegrationGetTrademarkInfoParsed(t *testing.T) {
	client := getTestClient(t)
	result, err := client.GetTrademarkInfoParsed(context.Background(), itTrademarkNumber)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetTrademarkInfoParsed error = %v", err)
	}
	if result.RegistrationNumber != itTrademarkNumber {
		t.Errorf("RegistrationNumber = %q, want %q", result.RegistrationNumber, itTrademarkNumber)
	}
	if result.MarkText == "" {
		t.Error("MarkText is empty")
	}
	if len(result.Applicants) == 0 {
		t.Error("no applicants")
	}
	t.Logf("GetTrademarkInfoParsed: %s - %s (%s)", result.RegistrationNumber, result.MarkText, result.Status)
}

func TestIntegrationGetTrademarkImage(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetTrademarkImage(context.Background(), itTrademarkNumber)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetTrademarkImage error = %v", err)
	}
	assertImageBytes(t, "GetTrademarkImage", data)
}

func TestIntegrationGetTrademarkThumbnail(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetTrademarkThumbnail(context.Background(), itTrademarkNumber)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetTrademarkThumbnail error = %v", err)
	}
	assertImageBytes(t, "GetTrademarkThumbnail", data)
}

func TestIntegrationGetTrademarkBibDataApplied(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetTrademarkBibDataApplied(context.Background(), itBulkYear, itBulkWeek)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetTrademarkBibDataApplied error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetTrademarkBibDataApplied returned %d bytes", len(data))
}

func TestIntegrationGetTrademarkBibDataRegistered(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetTrademarkBibDataRegistered(context.Background(), itBulkYear, itBulkWeek)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetTrademarkBibDataRegistered error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetTrademarkBibDataRegistered returned %d bytes", len(data))
}

func TestIntegrationGetTrademarkBibDataRejected(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetTrademarkBibDataRejected(context.Background(), itBulkYear, itBulkWeek)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetTrademarkBibDataRejected error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetTrademarkBibDataRejected returned %d bytes", len(data))
}

func TestIntegrationGetTrademarkRegisterExtract(t *testing.T) {
	client := getTestClient(t)
	data, err := client.GetTrademarkRegisterExtract(context.Background(), itBulkDate(), PeriodDaily)
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetTrademarkRegisterExtract error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("GetTrademarkRegisterExtract returned %d bytes", len(data))
}

func TestIntegrationGetTrademarkBibDataAppliedStream(t *testing.T) {
	client := getTestClient(t)
	data, err := reachBulk(t, func(w *boundedWriter) error {
		return client.GetTrademarkBibDataAppliedStream(context.Background(), itBulkYear, itBulkWeek, w)
	})
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetTrademarkBibDataAppliedStream error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stream produced no bytes")
	}
	t.Logf("GetTrademarkBibDataAppliedStream read %d bytes", len(data))
}

func TestIntegrationGetTrademarkBibDataRegisteredStream(t *testing.T) {
	client := getTestClient(t)
	data, err := reachBulk(t, func(w *boundedWriter) error {
		return client.GetTrademarkBibDataRegisteredStream(context.Background(), itBulkYear, itBulkWeek, w)
	})
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetTrademarkBibDataRegisteredStream error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stream produced no bytes")
	}
	t.Logf("GetTrademarkBibDataRegisteredStream read %d bytes", len(data))
}

func TestIntegrationGetTrademarkBibDataRejectedStream(t *testing.T) {
	client := getTestClient(t)
	data, err := reachBulk(t, func(w *boundedWriter) error {
		return client.GetTrademarkBibDataRejectedStream(context.Background(), itBulkYear, itBulkWeek, w)
	})
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetTrademarkBibDataRejectedStream error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stream produced no bytes")
	}
	t.Logf("GetTrademarkBibDataRejectedStream read %d bytes", len(data))
}

func TestIntegrationGetTrademarkRegisterExtractStream(t *testing.T) {
	client := getTestClient(t)
	data, err := reachBulk(t, func(w *boundedWriter) error {
		return client.GetTrademarkRegisterExtractStream(context.Background(), itBulkDate(), PeriodDaily, w)
	})
	if err != nil {
		skipIfTransient(t, err)
		t.Fatalf("GetTrademarkRegisterExtractStream error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stream produced no bytes")
	}
	t.Logf("GetTrademarkRegisterExtractStream read %d bytes", len(data))
}

// ---- supplementary edge-case tests (fold-in of the original suite) ---------
// These exercise error/normalization paths beyond the per-endpoint smoke tests.
// Their names carry an underscore so they do not collide with the
// TestIntegration<Method> coverage names above.

func TestIntegration_GetPatentInfoParsed_InvalidNumber(t *testing.T) {
	client := getTestClient(t)
	// 999999999 is a bare registered number that does not exist; the API returns
	// an <Error> root instead of <dpma-patent-document>. The client must surface a
	// typed error, not an XMLParseError.
	_, err := client.GetPatentInfoParsed(context.Background(), "999999999")
	if err == nil {
		t.Fatal("expected error for invalid patent number")
	}
	var xmlErr *XMLParseError
	if errors.As(err, &xmlErr) {
		t.Errorf("got XMLParseError (error response not detected): %v", err)
	}
	t.Logf("correctly received %T: %v", err, err)
}

func TestIntegration_DataNotAvailable(t *testing.T) {
	client := getTestClient(t)
	_, err := client.GetDisclosureDocumentsXML(context.Background(), 2099, 1)
	if err == nil {
		t.Fatal("expected DataNotAvailableError for a future week")
	}
	if _, ok := err.(*DataNotAvailableError); !ok {
		t.Errorf("expected *DataNotAvailableError, got %T: %v", err, err)
	}
	t.Logf("correctly received DataNotAvailableError")
}

func TestIntegration_GetPatentInfoParsed_NormalizeFormats(t *testing.T) {
	client := getTestClient(t)
	tests := []struct{ name, input string }{
		{"bare registered number", "100273602"},
		{"with DE prefix", "DE10027362"},
		{"with DE prefix and kind code C2", "DE10027362C2"},
		{"publication number A1", "DE102019200907A1"},
		{"lowercase", "de10027362c2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := client.GetPatentInfoParsed(context.Background(), tt.input)
			if err != nil {
				skipIfTransient(t, err)
				t.Fatalf("GetPatentInfoParsed(%q) error = %v", tt.input, err)
			}
			if info.Title == "" {
				t.Errorf("GetPatentInfoParsed(%q) returned empty title", tt.input)
			}
			t.Logf("GetPatentInfoParsed(%q) -> title=%q type=%q", tt.input, info.Title, info.IPRightType)
		})
	}
}

func TestIntegration_GetPatentInfoByPublicationNumber_OldPatents(t *testing.T) {
	client := getTestClient(t)
	tests := []struct{ name, input string }{
		{"DE123C (publication kind)", "DE123C"},
		{"DE123A (application kind)", "DE123A"},
		{"search PN=DE000000000123A", "DE000000000123A"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := client.GetPatentInfoByPublicationNumber(context.Background(), tt.input)
			if err != nil {
				var notFound *NotFoundError
				if errors.As(err, &notFound) {
					t.Logf("%s not found via publication number search", tt.input)
					return
				}
				skipIfTransient(t, err)
				t.Fatalf("GetPatentInfoByPublicationNumber(%s) error = %v", tt.input, err)
			}
			t.Logf("GetPatentInfoByPublicationNumber(%s) -> title=%q", tt.input, info.Title)
		})
	}
}

// ---- test-only helpers -----------------------------------------------------

// assertImageBytes pins the image/media contract: non-empty bytes with a
// recognisable raster/document format magic (PNG, JPEG, GIF, TIFF, BMP or PDF).
func assertImageBytes(t *testing.T, label string, data []byte) {
	t.Helper()
	if len(data) == 0 {
		t.Fatalf("%s returned empty data", label)
	}
	if !looksLikeImage(data) {
		n := 8
		if len(data) < n {
			n = len(data)
		}
		t.Errorf("%s did not return a recognised image/PDF (magic %x)", label, data[:n])
	}
	t.Logf("%s returned %d bytes", label, len(data))
}

func looksLikeImage(data []byte) bool {
	switch {
	case bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")):
		return true
	case bytes.HasPrefix(data, []byte("\xff\xd8\xff")): // JPEG
		return true
	case bytes.HasPrefix(data, []byte("GIF87a")), bytes.HasPrefix(data, []byte("GIF89a")):
		return true
	case bytes.HasPrefix(data, []byte("II*\x00")), bytes.HasPrefix(data, []byte("MM\x00*")): // TIFF
		return true
	case bytes.HasPrefix(data, []byte("BM")): // BMP
		return true
	case bytes.HasPrefix(data, []byte("%PDF")):
		return true
	default:
		return false
	}
}
