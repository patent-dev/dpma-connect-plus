package dpmaconnect

import (
	"context"
	"net/http"
	"testing"
)

func TestSearchTrademarks(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterMarkeService/search/")
		w.WriteHeader(http.StatusOK)
		w.Write(trademarkSearchXML)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	data, err := client.SearchTrademarks(context.Background(), "md=Apple")
	if err != nil {
		t.Fatalf("SearchTrademarks() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("returned empty data")
	}
}

func TestSearchTrademarks_Unauthorized(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.SearchTrademarks(context.Background(), "md=Apple")
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestGetTrademarkInfo(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterMarkeService/getRegisterInfo/919224")
		w.WriteHeader(http.StatusOK)
		w.Write(trademarkInfoXML)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	data, err := client.GetTrademarkInfo(context.Background(), "919224")
	if err != nil {
		t.Fatalf("GetTrademarkInfo() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("returned empty data")
	}
}

func TestGetTrademarkInfo_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.GetTrademarkInfo(context.Background(), "INVALID")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("expected *NotFoundError, got %T", err)
	}
}

func TestGetTrademarkImage(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterMarkeService/getRegisterFullImage/919224")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("\x89PNG image data"))
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	data, err := client.GetTrademarkImage(context.Background(), "919224")
	if err != nil {
		t.Fatalf("GetTrademarkImage() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("returned empty data")
	}
}

func TestGetTrademarkImage_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.GetTrademarkImage(context.Background(), "INVALID")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("expected *NotFoundError, got %T", err)
	}
}

func TestGetTrademarkThumbnail(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterMarkeService/getRegisterThumbnailImage/919224")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("thumbnail data"))
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	data, err := client.GetTrademarkThumbnail(context.Background(), "919224")
	if err != nil {
		t.Fatalf("GetTrademarkThumbnail() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("returned empty data")
	}
}

func TestSearchTrademarksParsed(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterMarkeService/search/")
		w.WriteHeader(http.StatusOK)
		w.Write(trademarkSearchXML)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	result, err := client.SearchTrademarksParsed(context.Background(), "md=Apple")
	if err != nil {
		t.Fatalf("SearchTrademarksParsed() error = %v", err)
	}
	if result.TotalHits != 3 {
		t.Errorf("TotalHits = %d, want 3", result.TotalHits)
	}
}

func TestGetTrademarkInfoParsed(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireAuth(t, r)
		requirePath(t, r, "/DPMAregisterMarkeService/getRegisterInfo/919224")
		w.WriteHeader(http.StatusOK)
		w.Write(trademarkInfoXML)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	result, err := client.GetTrademarkInfoParsed(context.Background(), "919224")
	if err != nil {
		t.Fatalf("GetTrademarkInfoParsed() error = %v", err)
	}
	if result.RegistrationNumber != "919224" {
		t.Errorf("RegistrationNumber = %q", result.RegistrationNumber)
	}
	if result.MarkFeature != "Word" {
		t.Errorf("MarkFeature = %q", result.MarkFeature)
	}
}

func TestGetTrademarkInfoParsed_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	_, err := client.GetTrademarkInfoParsed(context.Background(), "INVALID")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}
