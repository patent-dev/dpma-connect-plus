package dpmaconnect

import (
	"context"
	"net/http"
	"testing"
)

func TestSearchDesigns(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterGsmService/search/")
		w.WriteHeader(http.StatusOK)
		w.Write(designSearchXML)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	data, err := client.SearchDesigns(context.Background(), "INH=Samsung")
	if err != nil {
		t.Fatalf("SearchDesigns() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("returned empty data")
	}
}

func TestSearchDesigns_Unauthorized(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.SearchDesigns(context.Background(), "INH=Samsung")
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestGetDesignInfo(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterGsmService/getRegisterInfo/M8800174-0001")
		w.WriteHeader(http.StatusOK)
		w.Write(designInfoXML)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	data, err := client.GetDesignInfo(context.Background(), "M8800174-0001")
	if err != nil {
		t.Fatalf("GetDesignInfo() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("returned empty data")
	}
}

func TestGetDesignInfo_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.GetDesignInfo(context.Background(), "INVALID")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("expected *NotFoundError, got %T", err)
	}
}

func TestGetDesignImage(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterGsmService/getRegisterFullImage/M8800174-0001/1")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("\xff\xd8\xff\xe0 JPEG data"))
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	data, err := client.GetDesignImage(context.Background(), "M8800174-0001", "1")
	if err != nil {
		t.Fatalf("GetDesignImage() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("returned empty data")
	}
}

func TestGetDesignImage_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.GetDesignImage(context.Background(), "INVALID", "1")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("expected *NotFoundError, got %T", err)
	}
}

func TestGetDesignThumbnail(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterGsmService/getRegisterThumbnailImage/M8800174-0001/1")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("thumbnail data"))
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	data, err := client.GetDesignThumbnail(context.Background(), "M8800174-0001", "1")
	if err != nil {
		t.Fatalf("GetDesignThumbnail() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("returned empty data")
	}
}

func TestSearchDesignsParsed(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterGsmService/search/")
		w.WriteHeader(http.StatusOK)
		w.Write(designSearchXML)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	result, err := client.SearchDesignsParsed(context.Background(), "INH=Samsung")
	if err != nil {
		t.Fatalf("SearchDesignsParsed() error = %v", err)
	}
	if result.TotalHits != 3 {
		t.Errorf("TotalHits = %d, want 3", result.TotalHits)
	}
	if result.Hits[0].DesignIdentifier != "M9000603-0001" {
		t.Errorf("Hits[0].DesignIdentifier = %q", result.Hits[0].DesignIdentifier)
	}
}

func TestGetDesignInfoParsed(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterGsmService/getRegisterInfo/M8800174-0001")
		w.WriteHeader(http.StatusOK)
		w.Write(designInfoXML)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	result, err := client.GetDesignInfoParsed(context.Background(), "M8800174-0001")
	if err != nil {
		t.Fatalf("GetDesignInfoParsed() error = %v", err)
	}
	if result.DesignIdentifier != "M8800174-0001" {
		t.Errorf("DesignIdentifier = %q", result.DesignIdentifier)
	}
	if result.Title != "Stuhl" {
		t.Errorf("Title = %q", result.Title)
	}
}

func TestGetDesignInfoParsed_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.GetDesignInfoParsed(context.Background(), "INVALID")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}
