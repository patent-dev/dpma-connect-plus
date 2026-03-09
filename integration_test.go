//go:build integration

package dpmaconnect

import (
	"context"
	"os"
	"testing"
	"time"
)

func getTestClient(t *testing.T) *Client {
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

func TestIntegration_GetVersion(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	services := []string{ServicePatent, ServiceDesign, ServiceTrademark}

	for _, service := range services {
		t.Run(service, func(t *testing.T) {
			version, err := client.GetVersion(ctx, service)
			if err != nil {
				t.Errorf("GetVersion(%s) error = %v", service, err)
				return
			}

			if version == "" {
				t.Errorf("GetVersion(%s) returned empty version", service)
			}

			t.Logf("GetVersion(%s) = %s", service, version)
		})
	}
}

func TestIntegration_SearchPatents(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	query := "TI=Elektrofahrzeug"

	data, err := client.SearchPatents(ctx, query)
	if err != nil {
		t.Errorf("SearchPatents(%s) error = %v", query, err)
		return
	}

	if len(data) == 0 {
		t.Error("SearchPatents returned empty data")
	}

	t.Logf("SearchPatents(%s) returned %d bytes", query, len(data))
}

func TestIntegration_SearchDesigns(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	query := "INH=Samsung"

	data, err := client.SearchDesigns(ctx, query)
	if err != nil {
		t.Errorf("SearchDesigns(%s) error = %v", query, err)
		return
	}

	if len(data) == 0 {
		t.Error("SearchDesigns returned empty data")
	}

	t.Logf("SearchDesigns(%s) returned %d bytes", query, len(data))
}

func TestIntegration_SearchTrademarks(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	query := "md=Apple"

	data, err := client.SearchTrademarks(ctx, query)
	if err != nil {
		t.Errorf("SearchTrademarks(%s) error = %v", query, err)
		return
	}

	if len(data) == 0 {
		t.Error("SearchTrademarks returned empty data")
	}

	t.Logf("SearchTrademarks(%s) returned %d bytes", query, len(data))
}

func TestIntegration_GetPatentPublicationPDF(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	docID := "DE102019200907A1"

	data, err := client.GetPatentPublicationPDF(ctx, docID)
	if err != nil {
		if _, ok := err.(*NotFoundError); ok {
			t.Logf("Document %s not found (expected for some IDs)", docID)
			return
		}
		t.Errorf("GetPatentPublicationPDF(%s) error = %v", docID, err)
		return
	}

	if len(data) == 0 {
		t.Error("GetPatentPublicationPDF returned empty data")
	}

	if len(data) >= 4 && string(data[:4]) != "%PDF" {
		t.Error("GetPatentPublicationPDF did not return valid PDF data")
	}

	t.Logf("GetPatentPublicationPDF(%s) returned %d bytes", docID, len(data))
}

func TestIntegration_GetPatentInfo(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	id := "100273602"

	data, err := client.GetPatentInfo(ctx, id)
	if err != nil {
		if _, ok := err.(*NotFoundError); ok {
			t.Logf("Patent %s not found (expected for some IDs)", id)
			return
		}
		t.Errorf("GetPatentInfo(%s) error = %v", id, err)
		return
	}

	if len(data) == 0 {
		t.Error("GetPatentInfo returned empty data")
	}

	t.Logf("GetPatentInfo(%s) returned %d bytes", id, len(data))
}

func TestIntegration_GetDesignInfo(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	id := "M8800174-0001"

	data, err := client.GetDesignInfo(ctx, id)
	if err != nil {
		if _, ok := err.(*NotFoundError); ok {
			t.Logf("Design %s not found (expected for some IDs)", id)
			return
		}
		t.Errorf("GetDesignInfo(%s) error = %v", id, err)
		return
	}

	if len(data) == 0 {
		t.Error("GetDesignInfo returned empty data")
	}

	t.Logf("GetDesignInfo(%s) returned %d bytes", id, len(data))
}

func TestIntegration_GetTrademarkInfo(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	id := "919224"

	data, err := client.GetTrademarkInfo(ctx, id)
	if err != nil {
		if _, ok := err.(*NotFoundError); ok {
			t.Logf("Trademark %s not found (expected for some IDs)", id)
			return
		}
		t.Errorf("GetTrademarkInfo(%s) error = %v", id, err)
		return
	}

	if len(data) == 0 {
		t.Error("GetTrademarkInfo returned empty data")
	}

	t.Logf("GetTrademarkInfo(%s) returned %d bytes", id, len(data))
}

func TestIntegration_GetDisclosureDocumentsXML(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	year := 2026
	week := 9

	data, err := client.GetDisclosureDocumentsXML(ctx, year, week)
	if err != nil {
		if _, ok := err.(*DataNotAvailableError); ok {
			t.Logf("Data not available for %d week %d (expected for some weeks)", year, week)
			return
		}
		t.Errorf("GetDisclosureDocumentsXML(%d, %d) error = %v", year, week, err)
		return
	}

	if len(data) == 0 {
		t.Error("GetDisclosureDocumentsXML returned empty data")
	}

	if len(data) >= 2 && string(data[:2]) != "PK" {
		t.Error("GetDisclosureDocumentsXML did not return valid ZIP data")
	}

	t.Logf("GetDisclosureDocumentsXML(%d, %d) returned %d bytes", year, week, len(data))
}

func TestIntegration_GetDesignBibliographicDataXML(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	year := 2026
	week := 9

	data, err := client.GetDesignBibliographicDataXML(ctx, year, week)
	if err != nil {
		if _, ok := err.(*DataNotAvailableError); ok {
			t.Logf("Data not available for %d week %d", year, week)
			return
		}
		t.Errorf("GetDesignBibliographicDataXML(%d, %d) error = %v", year, week, err)
		return
	}

	if len(data) == 0 {
		t.Error("GetDesignBibliographicDataXML returned empty data")
	}

	t.Logf("GetDesignBibliographicDataXML(%d, %d) returned %d bytes", year, week, len(data))
}

func TestIntegration_GetTrademarkBibDataApplied(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	year := 2026
	week := 9

	data, err := client.GetTrademarkBibDataApplied(ctx, year, week)
	if err != nil {
		if _, ok := err.(*DataNotAvailableError); ok {
			t.Logf("Data not available for %d week %d", year, week)
			return
		}
		t.Errorf("GetTrademarkBibDataApplied(%d, %d) error = %v", year, week, err)
		return
	}

	if len(data) == 0 {
		t.Error("GetTrademarkBibDataApplied returned empty data")
	}

	t.Logf("GetTrademarkBibDataApplied(%d, %d) returned %d bytes", year, week, len(data))
}

func TestIntegration_GetPatentRegisterExtract(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	date := time.Date(2024, 10, 23, 0, 0, 0, 0, time.UTC)

	data, err := client.GetPatentRegisterExtract(ctx, date, PeriodDaily)
	if err != nil {
		if _, ok := err.(*DataNotAvailableError); ok {
			t.Logf("Data not available for %s", FormatDate(date))
			return
		}
		t.Errorf("GetPatentRegisterExtract() error = %v", err)
		return
	}

	if len(data) == 0 {
		t.Error("GetPatentRegisterExtract returned empty data")
	}

	t.Logf("GetPatentRegisterExtract(%s, %s) returned %d bytes", FormatDate(date), PeriodDaily, len(data))
}

func TestIntegration_SearchPatentsParsed(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	result, err := client.SearchPatentsParsed(ctx, "TI=Elektrofahrzeug")
	if err != nil {
		t.Fatalf("SearchPatentsParsed() error = %v", err)
	}

	if result.TotalHits == 0 {
		t.Error("TotalHits = 0, expected > 0")
	}
	if len(result.Hits) == 0 {
		t.Error("no hits returned")
	}

	h := result.Hits[0]
	if h.LeadingRegisteredNumber == "" {
		t.Error("Hits[0].LeadingRegisteredNumber is empty")
	}
	if h.Title == "" {
		t.Error("Hits[0].Title is empty")
	}

	t.Logf("SearchPatentsParsed: %d hits, first: %s - %s", result.TotalHits, h.LeadingRegisteredNumber, h.Title)
}

func TestIntegration_GetPatentInfoParsed(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	result, err := client.GetPatentInfoParsed(ctx, "100273602")
	if err != nil {
		t.Fatalf("GetPatentInfoParsed() error = %v", err)
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

func TestIntegration_SearchDesignsParsed(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	result, err := client.SearchDesignsParsed(ctx, "INH=Samsung")
	if err != nil {
		t.Fatalf("SearchDesignsParsed() error = %v", err)
	}

	if result.TotalHits == 0 {
		t.Error("TotalHits = 0")
	}
	if len(result.Hits) == 0 {
		t.Error("no hits returned")
	}

	t.Logf("SearchDesignsParsed: %d hits, first: %s - %s", result.TotalHits, result.Hits[0].DesignIdentifier, result.Hits[0].Title)
}

func TestIntegration_GetDesignInfoParsed(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	result, err := client.GetDesignInfoParsed(ctx, "M8800174-0001")
	if err != nil {
		t.Fatalf("GetDesignInfoParsed() error = %v", err)
	}

	if result.DesignIdentifier != "M8800174-0001" {
		t.Errorf("DesignIdentifier = %q", result.DesignIdentifier)
	}
	if result.Title == "" {
		t.Error("Title is empty")
	}

	t.Logf("GetDesignInfoParsed: %s - %s (%s)", result.DesignIdentifier, result.Title, result.Status)
}

func TestIntegration_SearchTrademarksParsed(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	result, err := client.SearchTrademarksParsed(ctx, "md=Apple")
	if err != nil {
		t.Fatalf("SearchTrademarksParsed() error = %v", err)
	}

	if result.TotalHits == 0 {
		t.Error("TotalHits = 0")
	}
	if len(result.Hits) == 0 {
		t.Error("no hits returned")
	}

	t.Logf("SearchTrademarksParsed: %d hits, first: %s - %s", result.TotalHits, result.Hits[0].ApplicationNumber, result.Hits[0].MarkText)
}

func TestIntegration_GetTrademarkInfoParsed(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	result, err := client.GetTrademarkInfoParsed(ctx, "919224")
	if err != nil {
		t.Fatalf("GetTrademarkInfoParsed() error = %v", err)
	}

	if result.RegistrationNumber != "919224" {
		t.Errorf("RegistrationNumber = %q", result.RegistrationNumber)
	}
	if result.MarkText == "" {
		t.Error("MarkText is empty")
	}
	if len(result.Applicants) == 0 {
		t.Error("no applicants")
	}

	t.Logf("GetTrademarkInfoParsed: %s - %s (%s)", result.RegistrationNumber, result.MarkText, result.Status)
}

func TestIntegration_DataNotAvailable(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	year := 2099
	week := 1

	_, err := client.GetDisclosureDocumentsXML(ctx, year, week)
	if err == nil {
		t.Error("Expected DataNotAvailableError for future week")
		return
	}

	if _, ok := err.(*DataNotAvailableError); !ok {
		t.Errorf("Expected *DataNotAvailableError, got %T: %v", err, err)
	}

	t.Logf("Correctly received DataNotAvailableError for %d week %d", year, week)
}
