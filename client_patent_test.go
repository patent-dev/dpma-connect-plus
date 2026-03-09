package dpmaconnect

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestSearchPatents(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterPatService/search/")
		w.WriteHeader(http.StatusOK)
		w.Write(patentSearchXML)
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
	handler := func(w http.ResponseWriter, r *http.Request) {
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
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(xmlErr))
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
		w.Write([]byte("%PDF-1.4 fake content"))
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
	handler := func(w http.ResponseWriter, r *http.Request) {
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
		w.Write(patentInfoXML)
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
	handler := func(w http.ResponseWriter, r *http.Request) {
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
		w.Write([]byte("%PDF-fulltext"))
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
	handler := func(w http.ResponseWriter, r *http.Request) {
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
		w.Write(patentSearchXML)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	result, err := client.SearchPatentsParsed(context.Background(), "TI=Elektrofahrzeug")
	if err != nil {
		t.Fatalf("SearchPatentsParsed() error = %v", err)
	}
	if result.TotalHits != 3 {
		t.Errorf("TotalHits = %d, want 3", result.TotalHits)
	}
	if len(result.Hits) != 3 {
		t.Errorf("len(Hits) = %d, want 3", len(result.Hits))
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
		w.Write(patentInfoXML)
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

func TestGetPatentInfoParsed_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.GetPatentInfoParsed(context.Background(), "INVALID")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("expected *NotFoundError, got %T", err)
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
			w.Write(patentSearchXML)
		} else {
			requirePath(t, r, "/DPMAregisterPatService/getRegisterInfo/")
			w.Write(patentInfoXML)
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

func TestGetPatentInfoByPublicationNumber_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><PatentHitList HitCount="0"/>`))
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
		w.Write([]byte(xmlErr))
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
