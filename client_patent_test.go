package dpmaconnect

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestSearchPatents(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterPatService/search/")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(patentSearchXML)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	data, err := client.SearchPatents(context.Background(), "TI=Elektrofahrzeug")
	if err != nil {
		t.Fatalf("SearchPatents() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("SearchPatents() returned empty data")
	}
}

func TestSearchPatents_Unauthorized(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.SearchPatents(context.Background(), "pa=Siemens")
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestSearchPatentsXMLErrorOn200(t *testing.T) {
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
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(xmlErr))
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.SearchPatents(context.Background(), "TI=test")
	if err == nil {
		t.Fatal("expected error for XML error on 200")
	}
	var dna *DataNotAvailableError
	if !errors.As(err, &dna) {
		t.Errorf("expected *DataNotAvailableError, got %T: %v", err, err)
	}
}

func TestGetPatentPublicationPDF(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterPatService/getPatentpublikation_PDF/DE102023000001A1")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("%PDF-1.4 fake content"))
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	data, err := client.GetPatentPublicationPDF(context.Background(), "DE102023000001A1")
	if err != nil {
		t.Fatalf("GetPatentPublicationPDF() error = %v", err)
	}
	if string(data[:4]) != "%PDF" {
		t.Errorf("expected PDF magic bytes, got %q", data[:4])
	}
}

func TestGetPatentPublicationPDF_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.GetPatentPublicationPDF(context.Background(), "INVALID123")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("expected *NotFoundError, got %T", err)
	}
}

func TestGetPatentInfo(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterPatService/getRegisterInfo/100273602")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(patentInfoXML)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	data, err := client.GetPatentInfo(context.Background(), "100273602")
	if err != nil {
		t.Fatalf("GetPatentInfo() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("GetPatentInfo() returned empty data")
	}
}

func TestGetPatentInfo_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.GetPatentInfo(context.Background(), "INVALID")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("expected *NotFoundError, got %T", err)
	}
}

func TestGetSearchableFullText(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterPatService/getRecherchierbarerVolltext/DE102019200907A1")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("%PDF-fulltext"))
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	data, err := client.GetSearchableFullText(context.Background(), "DE102019200907A1")
	if err != nil {
		t.Fatalf("GetSearchableFullText() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("returned empty data")
	}
}

func TestGetSearchableFullText_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.GetSearchableFullText(context.Background(), "INVALID")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("expected *NotFoundError, got %T", err)
	}
}

func TestSearchPatentsParsed(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterPatService/search/")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(patentSearchXML)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	result, err := client.SearchPatentsParsed(context.Background(), "TI=Elektrofahrzeug")
	if err != nil {
		t.Fatalf("SearchPatentsParsed() error = %v", err)
	}
	if result.TotalHits != 4 {
		t.Errorf("TotalHits = %d, want 4", result.TotalHits)
	}
	if len(result.Hits) != 4 {
		t.Errorf("len(Hits) = %d, want 4", len(result.Hits))
	}
	if result.Hits[0].Title != "Elektrofahrzeug" {
		t.Errorf("Hits[0].Title = %q", result.Hits[0].Title)
	}
}

func TestGetPatentInfoParsed(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterPatService/getRegisterInfo/100273602")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(patentInfoXML)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	result, err := client.GetPatentInfoParsed(context.Background(), "100273602")
	if err != nil {
		t.Fatalf("GetPatentInfoParsed() error = %v", err)
	}
	if result.IPRightType != "patent" {
		t.Errorf("IPRightType = %q", result.IPRightType)
	}
	if result.Title == "" {
		t.Error("Title is empty")
	}
}

func TestIsRegisteredNumber(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"100273629", true},
		{"123", true},
		{"1120190012347", true}, // DE national phase of a PCT application
		{"E247834666", true},    // EP 24783466.6 in the German register
		{"e247834666", true},
		{" E247834666 ", true},
		{"", false},
		{"E", false},
		{"EP4780303", false}, // an EP publication number, resolved by search
		{"E24783466.6", false},
		{"DE123C", false},
		{"DE10027362", false},
		{"de123c", false},
		{"10027362B4", false},
		// A PCT registered number has the shape of a publication number, so it is
		// only fetched when a search reports it, never classified from input.
		{"CN2019089592", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isRegisteredNumber(tt.input)
			if got != tt.want {
				t.Errorf("isRegisteredNumber(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetPatentInfoParsed_RegisteredNumber(t *testing.T) {
	// Bare registered number should call getRegisterInfo directly
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterPatService/getRegisterInfo/100273629")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(patentInfoXML)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	result, err := client.GetPatentInfoParsed(context.Background(), "100273629")
	if err != nil {
		t.Fatalf("GetPatentInfoParsed error = %v", err)
	}
	if result.Title == "" {
		t.Error("Title is empty")
	}
}

func TestGetPatentInfoParsed_PublicationNumber(t *testing.T) {
	// DE patent number with prefix should resolve via search
	reqCount := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		reqCount++
		w.WriteHeader(http.StatusOK)
		if reqCount == 1 {
			// First request: search by PN=DE10027362C2
			requirePath(t, r, "/DPMAregisterPatService/search/")
			_, _ = w.Write(patentSearchXML)
		} else {
			// Second request: get info by registered number from search result
			requirePath(t, r, "/DPMAregisterPatService/getRegisterInfo/")
			_, _ = w.Write(patentInfoXML)
		}
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	result, err := client.GetPatentInfoParsed(context.Background(), "DE10027362C2")
	if err != nil {
		t.Fatalf("GetPatentInfoParsed error = %v", err)
	}
	if result.Title == "" {
		t.Error("Title is empty")
	}
	if reqCount != 2 {
		t.Errorf("expected 2 requests (search + info), got %d", reqCount)
	}
}

func TestGetPatentInfoParsed_NotFound_RegisteredNumber(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.GetPatentInfoParsed(context.Background(), "999999999")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("expected *NotFoundError, got %T", err)
	}
}

func TestGetPatentInfoParsed_NotFound_PublicationNumber(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><PatentHitList HitCount="0"/>`))
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.GetPatentInfoParsed(context.Background(), "DE999999999X1")
	if err == nil {
		t.Fatal("expected error for no results")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected *NotFoundError, got %T: %v", err, err)
	}
}

func TestGetPatentInfoByPublicationNumber(t *testing.T) {
	// Mock server returns search results for the first request,
	// then patent info for the second.
	reqCount := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		reqCount++
		w.WriteHeader(http.StatusOK)
		if reqCount == 1 {
			requirePath(t, r, "/DPMAregisterPatService/search/")
			_, _ = w.Write(patentSearchXML)
		} else {
			requirePath(t, r, "/DPMAregisterPatService/getRegisterInfo/")
			_, _ = w.Write(patentInfoXML)
		}
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	result, err := client.GetPatentInfoByPublicationNumber(context.Background(), "DE102019200907A1")
	if err != nil {
		t.Fatalf("GetPatentInfoByPublicationNumber() error = %v", err)
	}
	if result.Title == "" {
		t.Error("Title is empty")
	}
	if reqCount != 2 {
		t.Errorf("expected 2 requests (search + info), got %d", reqCount)
	}
}

// infoRegisteredNumber returns the registered number a getRegisterInfo request
// asks for, or "" for any other endpoint.
func infoRegisteredNumber(r *http.Request) string {
	const prefix = "/DPMAregisterPatService/getRegisterInfo/"
	i := strings.Index(r.URL.Path, prefix)
	if i < 0 {
		return ""
	}
	return r.URL.Path[i+len(prefix):]
}

// A search hit's leading-registered-number is fetched exactly as reported, even
// when it is not digits-only. The live register reports E247834666 for
// PN=EP4780303A1 and CN2019089592 for PN=WO2020001234A1 (fixtures captured
// 2026-11-16); before, the library re-classified those as publication numbers
// and searched PN=E247834666, which DPMA rejects with "invalid input value".
func TestGetPatentInfoParsed_NonNumericLeadingRegisteredNumber(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		search    []byte
		info      []byte
		wantFetch string
		wantApp   PatentDocumentRef
	}{
		{
			name: "EP application", input: "EP4780303A1",
			search: patentSearchEPXML, info: patentInfoEPApplicationXML,
			wantFetch: "E247834666",
			wantApp:   PatentDocumentRef{Country: "EP", Number: "24783466.6", Date: "2024-09-17"},
		},
		{
			name: "PCT application before the DE phase", input: "WO2020001234A1",
			search: patentSearchPCTXML, info: patentInfoPCTXML,
			wantFetch: "CN2019089592",
			wantApp:   PatentDocumentRef{Country: "WO", Number: "PCT/CN2019/089592", Date: "2019-05-31"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var searches int
			var fetched []string
			handler := func(w http.ResponseWriter, r *http.Request) {
				requireAuth(t, r)
				w.WriteHeader(http.StatusOK)
				if reg := infoRegisteredNumber(r); reg != "" {
					fetched = append(fetched, reg)
					_, _ = w.Write(tt.info)
					return
				}
				requirePath(t, r, "/DPMAregisterPatService/search/")
				searches++
				_, _ = w.Write(tt.search)
			}
			server, client := setupMockServer(t, handler)
			defer server.Close()

			got, err := client.GetPatentInfoParsed(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("GetPatentInfoParsed(%q) error = %v", tt.input, err)
			}
			if searches != 1 || len(fetched) != 1 || fetched[0] != tt.wantFetch {
				t.Fatalf("requests: %d search(es), fetched %v; want 1 search then a fetch of %s", searches, fetched, tt.wantFetch)
			}
			if got.ApplicationRef != tt.wantApp {
				t.Errorf("ApplicationRef = %+v, want %+v", got.ApplicationRef, tt.wantApp)
			}
		})
	}
}

// An E-prefixed registered number given directly is fetched without a search,
// upper-cased because the register is case sensitive.
func TestGetPatentInfoParsed_ERegisteredNumberDirect(t *testing.T) {
	for _, input := range []string{"E247834666", "e247834666"} {
		t.Run(input, func(t *testing.T) {
			var fetched []string
			handler := func(w http.ResponseWriter, r *http.Request) {
				requireAuth(t, r)
				reg := infoRegisteredNumber(r)
				if reg == "" {
					t.Errorf("unexpected request %s: a registered number must not be searched", r.URL.Path)
				}
				fetched = append(fetched, reg)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(patentInfoEPApplicationXML)
			}
			server, client := setupMockServer(t, handler)
			defer server.Close()

			got, err := client.GetPatentInfoParsed(context.Background(), input)
			if err != nil {
				t.Fatalf("GetPatentInfoParsed(%q) error = %v", input, err)
			}
			if len(fetched) != 1 || fetched[0] != "E247834666" {
				t.Errorf("fetched %v, want [E247834666]", fetched)
			}
			if got.Title != "SYSTEM UND VERFAHREN ZUR ANREGUNG UND DETEKTION VON FLUORESZENZINDIKATOREN" {
				t.Errorf("Title = %q", got.Title)
			}
		})
	}
}

// A search hit whose leading-registered-number the register then rejects ends
// after one search and one fetch with the register's own error: resolution can
// no longer recurse into a second search.
func TestGetPatentInfoParsed_ResolutionIsOneSearchOneFetch(t *testing.T) {
	searchXML := `<?xml version='1.0' encoding='utf-8'?>` +
		`<PatentHitList HitCount="1"><Counter><DocumentHits>1</DocumentHits><DatabaseHits>1</DatabaseHits></Counter>` +
		`<PatentHitListRecord><leading-registered-number>DE10027362C2</leading-registered-number>` +
		`<registered-number>DE10027362C2</registered-number><type>Patent</type></PatentHitListRecord></PatentHitList>`

	var searches, fetches int
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		w.WriteHeader(http.StatusOK)
		if infoRegisteredNumber(r) != "" {
			fetches++
			_, _ = w.Write(patentInfoInvalidInputXML)
			return
		}
		searches++
		_, _ = w.Write([]byte(searchXML))
	}
	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.GetPatentInfoParsed(context.Background(), "DE10027362C2")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || !strings.Contains(err.Error(), "invalid input value") {
		t.Fatalf("err = %T %v, want the register's invalid-input APIError", err, err)
	}
	if searches != 1 || fetches != 1 {
		t.Errorf("made %d search(es) and %d fetch(es), want exactly 1 and 1", searches, fetches)
	}
}

func TestGetPatentInfoByPublicationNumber_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><PatentHitList HitCount="0"/>`))
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.GetPatentInfoByPublicationNumber(context.Background(), "DENOTREAL")
	if err == nil {
		t.Fatal("expected error for no search results")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("expected *NotFoundError, got %T: %v", err, err)
	}
}

func TestGetDisclosureDocumentsXML_DataNotAvailable(t *testing.T) {
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
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterPatService/getOffenlegungsschriften_Volltext_XML/202445")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(xmlErr))
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.GetDisclosureDocumentsXML(context.Background(), 2024, 45)
	if err == nil {
		t.Fatal("expected error for data not available")
	}
	if _, ok := err.(*DataNotAvailableError); !ok {
		t.Errorf("expected *DataNotAvailableError, got %T", err)
	}
}

// TestWeeklyBulkHelper_Success verifies the fetchWeeklyBulk-backed methods format
// the publication week, call the endpoint, and return the body on success.
func TestWeeklyBulkHelper_Success(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/202407")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("PK\x03\x04binary-zip"))
	}
	server, client := setupMockServer(t, handler)
	defer server.Close()

	data, err := client.GetPatentSpecificationsXML(context.Background(), 2024, 7)
	if err != nil {
		t.Fatalf("GetPatentSpecificationsXML() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty body")
	}
}

// TestWeeklyBulkHelper_InvalidWeek verifies the helper rejects an invalid week
// before making any HTTP call.
func TestWeeklyBulkHelper_InvalidWeek(t *testing.T) {
	called := false
	handler := func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}
	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.GetUtilityModelsPDF(context.Background(), 2024, 99)
	if err == nil {
		t.Fatal("expected error for invalid week")
	}
	if called {
		t.Error("HTTP call should not be made for an invalid week")
	}
}

// TestWeeklyStreamHelper_Success verifies the streamWeekly-backed methods stream
// the response body to the destination writer.
func TestWeeklyStreamHelper_Success(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/202407")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("PK\x03\x04binary-zip-payload"))
	}
	server, client := setupMockServer(t, handler)
	defer server.Close()

	var buf bytes.Buffer
	if err := client.GetPatentSpecificationsXMLStream(context.Background(), 2024, 7, &buf); err != nil {
		t.Fatalf("GetPatentSpecificationsXMLStream() error = %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected streamed bytes")
	}
}
