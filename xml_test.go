package dpmaconnect

import (
	_ "embed"
	"errors"
	"testing"
)

//go:embed testdata/patent_search.xml
var patentSearchXML []byte

//go:embed testdata/patent_info.xml
var patentInfoXML []byte

//go:embed testdata/trademark_search.xml
var trademarkSearchXML []byte

//go:embed testdata/trademark_info.xml
var trademarkInfoXML []byte

//go:embed testdata/design_search.xml
var designSearchXML []byte

//go:embed testdata/design_info.xml
var designInfoXML []byte

func TestParsePatentSearch(t *testing.T) {
	result, err := ParsePatentSearch(patentSearchXML)
	if err != nil {
		t.Fatalf("ParsePatentSearch() error = %v", err)
	}

	if result.TotalHits != 3 {
		t.Errorf("TotalHits = %d, want 3", result.TotalHits)
	}
	if len(result.Hits) != 3 {
		t.Fatalf("len(Hits) = %d, want 3", len(result.Hits))
	}

	if result.RawXML == nil {
		t.Error("RawXML is nil")
	}
	if len(result.RawXML) != len(patentSearchXML) {
		t.Errorf("RawXML length = %d, want %d", len(result.RawXML), len(patentSearchXML))
	}

	h := result.Hits[0]
	if h.LeadingRegisteredNumber != "82337144" {
		t.Errorf("Hits[0].LeadingRegisteredNumber = %q, want %q", h.LeadingRegisteredNumber, "82337144")
	}
	if h.RegisteredNumber != "G 82 33 714.4" {
		t.Errorf("Hits[0].RegisteredNumber = %q, want %q", h.RegisteredNumber, "G 82 33 714.4")
	}
	if h.Type != "Utility model" {
		t.Errorf("Hits[0].Type = %q, want %q", h.Type, "Utility model")
	}
	if h.LegalStatus != "Not pending/lapsed" {
		t.Errorf("Hits[0].LegalStatus = %q, want %q", h.LegalStatus, "Not pending/lapsed")
	}
	if h.Title != "Elektrofahrzeug" {
		t.Errorf("Hits[0].Title = %q, want %q", h.Title, "Elektrofahrzeug")
	}
	if len(h.Applicants) != 1 {
		t.Fatalf("len(Hits[0].Applicants) = %d, want 1", len(h.Applicants))
	}
	if h.Applicants[0] != "ELO - MA - HG Peter Kurs Ing. VDI GmbH & Co, 75334 Straubenhardt, DE" {
		t.Errorf("Hits[0].Applicants[0] = %q", h.Applicants[0])
	}
	if len(h.IPCClasses) != 1 || h.IPCClasses[0] != "B60K 1/00" {
		t.Errorf("Hits[0].IPCClasses = %v, want [B60K 1/00]", h.IPCClasses)
	}
	if h.ApplicationDate != "1982-12-01" {
		t.Errorf("Hits[0].ApplicationDate = %q", h.ApplicationDate)
	}

	// Third hit has further-classification
	h2 := result.Hits[2]
	if len(h2.IPCClasses) != 2 {
		t.Errorf("Hits[2].IPCClasses = %v, want 2 classes", h2.IPCClasses)
	}
}

func TestParsePatentSearch_Empty(t *testing.T) {
	xml := []byte(`<?xml version="1.0" encoding="UTF-8"?><PatentHitList HitCount="0"/>`)
	result, err := ParsePatentSearch(xml)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if result.TotalHits != 0 {
		t.Errorf("TotalHits = %d, want 0", result.TotalHits)
	}
	if len(result.Hits) != 0 {
		t.Errorf("len(Hits) = %d, want 0", len(result.Hits))
	}
}

func TestParsePatentSearch_MalformedXML(t *testing.T) {
	_, err := ParsePatentSearch([]byte("not xml"))
	if err == nil {
		t.Fatal("expected error for malformed XML")
	}
	var xmlErr *XMLParseError
	if !errors.As(err, &xmlErr) {
		t.Errorf("expected *XMLParseError, got %T", err)
	}
	if xmlErr.Operation != "ParsePatentSearch" {
		t.Errorf("Operation = %q, want ParsePatentSearch", xmlErr.Operation)
	}
}

func TestParsePatentInfo(t *testing.T) {
	result, err := ParsePatentInfo(patentInfoXML)
	if err != nil {
		t.Fatalf("ParsePatentInfo() error = %v", err)
	}

	if result.IPRightType != "patent" {
		t.Errorf("IPRightType = %q, want %q", result.IPRightType, "patent")
	}
	if result.Status != "nicht-anhaengig-erloschen" {
		t.Errorf("Status = %q, want %q", result.Status, "nicht-anhaengig-erloschen")
	}
	if result.Title != "Dreh-Energieumsetzer für ein Hybrid-Elektrofahrzeug" {
		t.Errorf("Title = %q", result.Title)
	}
	if result.FilingDate != "2000-06-02" {
		t.Errorf("FilingDate = %q", result.FilingDate)
	}
	if result.FirstPubDate != "2000-12-07" {
		t.Errorf("FirstPubDate = %q", result.FirstPubDate)
	}

	// Application reference
	if result.ApplicationRef.Country != "DE" {
		t.Errorf("ApplicationRef.Country = %q", result.ApplicationRef.Country)
	}
	if result.ApplicationRef.Number != "10027360.2" {
		t.Errorf("ApplicationRef.Number = %q", result.ApplicationRef.Number)
	}
	if result.ApplicationRef.Date != "2000-06-02" {
		t.Errorf("ApplicationRef.Date = %q", result.ApplicationRef.Date)
	}

	// Publications
	if len(result.Publications) != 1 {
		t.Fatalf("len(Publications) = %d, want 1", len(result.Publications))
	}
	pub := result.Publications[0]
	if pub.Country != "DE" || pub.Number != "000010027360" || pub.Kind != "A1" {
		t.Errorf("Publications[0] = %+v", pub)
	}

	// IPC
	if len(result.IPCClasses) != 1 || result.IPCClasses[0] != "H02K0016000000" {
		t.Errorf("IPCClasses = %v", result.IPCClasses)
	}

	// Applicants
	if len(result.Applicants) != 1 {
		t.Fatalf("len(Applicants) = %d, want 1", len(result.Applicants))
	}
	if result.Applicants[0].Name != "DENSO CORPORATION, Kariya-city" {
		t.Errorf("Applicants[0].Name = %q", result.Applicants[0].Name)
	}
	if result.Applicants[0].Country != "JP" {
		t.Errorf("Applicants[0].Country = %q", result.Applicants[0].Country)
	}

	// Inventors
	if len(result.Inventors) != 1 {
		t.Fatalf("len(Inventors) = %d, want 1", len(result.Inventors))
	}
	if result.Inventors[0].Name != "Seguchi, Masahiro, Kariya" {
		t.Errorf("Inventors[0].Name = %q", result.Inventors[0].Name)
	}

	// Abstract
	if result.Abstract == "" {
		t.Error("Abstract is empty")
	}
	if len(result.Abstract) < 50 {
		t.Errorf("Abstract too short: %d chars", len(result.Abstract))
	}
}

func TestParsePatentInfo_MalformedXML(t *testing.T) {
	_, err := ParsePatentInfo([]byte("<broken"))
	if err == nil {
		t.Fatal("expected error")
	}
	var xmlErr *XMLParseError
	if !errors.As(err, &xmlErr) {
		t.Errorf("expected *XMLParseError, got %T", err)
	}
}

func TestParseTrademarkSearch(t *testing.T) {
	result, err := ParseTrademarkSearch(trademarkSearchXML)
	if err != nil {
		t.Fatalf("ParseTrademarkSearch() error = %v", err)
	}

	if result.TotalHits != 3 {
		t.Errorf("TotalHits = %d, want 3", result.TotalHits)
	}
	if len(result.Hits) != 3 {
		t.Fatalf("len(Hits) = %d, want 3", len(result.Hits))
	}

	if result.RawXML == nil {
		t.Error("RawXML is nil")
	}

	h := result.Hits[0]
	if h.ApplicationNumber != "919224" {
		t.Errorf("Hits[0].ApplicationNumber = %q", h.ApplicationNumber)
	}
	if h.MarkText != "von Eicken`s Apple and Grape" {
		t.Errorf("Hits[0].MarkText = %q", h.MarkText)
	}
	if h.MarkFeature != "wortmarke" {
		t.Errorf("Hits[0].MarkFeature = %q", h.MarkFeature)
	}
	if h.Classification != "34" {
		t.Errorf("Hits[0].Classification = %q", h.Classification)
	}
	if h.Status != "akte-vernichtet" {
		t.Errorf("Hits[0].Status = %q", h.Status)
	}
	if h.ApplicationDate != "1974-01-25" {
		t.Errorf("Hits[0].ApplicationDate = %q", h.ApplicationDate)
	}
	if h.RegistrationDate != "1974-06-06" {
		t.Errorf("Hits[0].RegistrationDate = %q", h.RegistrationDate)
	}
	if h.Applicant != "Joh. Wilh. von Eicken GmbH, 23556 Lübeck, DE" {
		t.Errorf("Hits[0].Applicant = %q", h.Applicant)
	}

	// Third hit has wort-bildmarke
	if result.Hits[2].MarkFeature != "wort-bildmarke" {
		t.Errorf("Hits[2].MarkFeature = %q", result.Hits[2].MarkFeature)
	}
}

func TestParseTrademarkSearch_MalformedXML(t *testing.T) {
	_, err := ParseTrademarkSearch([]byte("garbage"))
	if err == nil {
		t.Fatal("expected error")
	}
	var xmlErr *XMLParseError
	if !errors.As(err, &xmlErr) {
		t.Errorf("expected *XMLParseError, got %T", err)
	}
}

func TestParseTrademarkInfo(t *testing.T) {
	result, err := ParseTrademarkInfo(trademarkInfoXML)
	if err != nil {
		t.Fatalf("ParseTrademarkInfo() error = %v", err)
	}

	if result.ApplicationNumber != "E17341" {
		t.Errorf("ApplicationNumber = %q, want %q", result.ApplicationNumber, "E17341")
	}
	if result.RegistrationNumber != "919224" {
		t.Errorf("RegistrationNumber = %q, want %q", result.RegistrationNumber, "919224")
	}
	if result.ApplicationDate != "1974-01-25" {
		t.Errorf("ApplicationDate = %q", result.ApplicationDate)
	}
	if result.RegistrationDate != "1974-06-06" {
		t.Errorf("RegistrationDate = %q", result.RegistrationDate)
	}
	if result.ExpiryDate != "2014-01-31" {
		t.Errorf("ExpiryDate = %q", result.ExpiryDate)
	}
	if result.Status != "Akte vernichtet" {
		t.Errorf("Status = %q", result.Status)
	}
	if result.MarkFeature != "Word" {
		t.Errorf("MarkFeature = %q", result.MarkFeature)
	}
	if result.MarkText != "von Eicken`s Apple and Grape" {
		t.Errorf("MarkText = %q", result.MarkText)
	}

	// Applicants
	if len(result.Applicants) != 1 {
		t.Fatalf("len(Applicants) = %d, want 1", len(result.Applicants))
	}
	if result.Applicants[0].Name != "Joh. Wilh. von Eicken GmbH" {
		t.Errorf("Applicants[0].Name = %q", result.Applicants[0].Name)
	}
	if result.Applicants[0].Address != "Lübeck" {
		t.Errorf("Applicants[0].Address = %q", result.Applicants[0].Address)
	}
	if result.Applicants[0].Country != "DE" {
		t.Errorf("Applicants[0].Country = %q", result.Applicants[0].Country)
	}

	// Classifications
	if len(result.Classifications) != 1 {
		t.Fatalf("len(Classifications) = %d, want 1", len(result.Classifications))
	}
	if result.Classifications[0].Number != "34" {
		t.Errorf("Classifications[0].Number = %q", result.Classifications[0].Number)
	}
}

func TestParseTrademarkInfo_MalformedXML(t *testing.T) {
	_, err := ParseTrademarkInfo([]byte("bad"))
	if err == nil {
		t.Fatal("expected error")
	}
	var xmlErr *XMLParseError
	if !errors.As(err, &xmlErr) {
		t.Errorf("expected *XMLParseError, got %T", err)
	}
}

func TestParseDesignSearch(t *testing.T) {
	result, err := ParseDesignSearch(designSearchXML)
	if err != nil {
		t.Fatalf("ParseDesignSearch() error = %v", err)
	}

	if result.TotalHits != 3 {
		t.Errorf("TotalHits = %d, want 3", result.TotalHits)
	}
	if len(result.Hits) != 3 {
		t.Fatalf("len(Hits) = %d, want 3", len(result.Hits))
	}

	if result.RawXML == nil {
		t.Error("RawXML is nil")
	}

	h := result.Hits[0]
	if h.DesignIdentifier != "M9000603-0001" {
		t.Errorf("Hits[0].DesignIdentifier = %q", h.DesignIdentifier)
	}
	if h.ApplicationNumber != "M9000603.8" {
		t.Errorf("Hits[0].ApplicationNumber = %q", h.ApplicationNumber)
	}
	if h.Title != "Mikrowellenherde" {
		t.Errorf("Hits[0].Title = %q", h.Title)
	}
	if h.Applicant != "Samsung Electronics Co., Ltd., Suwon-si Gyeonggi-do, KR" {
		t.Errorf("Hits[0].Applicant = %q", h.Applicant)
	}
	if h.ClassNumber != "07-02" {
		t.Errorf("Hits[0].ClassNumber = %q", h.ClassNumber)
	}
	if h.ApplicationDate != "1990-01-25" {
		t.Errorf("Hits[0].ApplicationDate = %q", h.ApplicationDate)
	}
}

func TestParseDesignSearch_MalformedXML(t *testing.T) {
	_, err := ParseDesignSearch([]byte("nope"))
	if err == nil {
		t.Fatal("expected error")
	}
	var xmlErr *XMLParseError
	if !errors.As(err, &xmlErr) {
		t.Errorf("expected *XMLParseError, got %T", err)
	}
}

func TestParseDesignInfo(t *testing.T) {
	result, err := ParseDesignInfo(designInfoXML)
	if err != nil {
		t.Fatalf("ParseDesignInfo() error = %v", err)
	}

	if result.DesignIdentifier != "M8800174-0001" {
		t.Errorf("DesignIdentifier = %q", result.DesignIdentifier)
	}
	if result.ApplicationNumber != "M8800174.1" {
		t.Errorf("ApplicationNumber = %q", result.ApplicationNumber)
	}
	if result.RegistrationNumber != "M8800174" {
		t.Errorf("RegistrationNumber = %q", result.RegistrationNumber)
	}
	if result.RegistrationDate != "1988-09-15" {
		t.Errorf("RegistrationDate = %q", result.RegistrationDate)
	}
	if result.Title != "Stuhl" {
		t.Errorf("Title = %q", result.Title)
	}
	if result.ApplicationDate != "1988-07-07" {
		t.Errorf("ApplicationDate = %q", result.ApplicationDate)
	}
	if result.Status != "Paper File Destroyed" {
		t.Errorf("Status = %q", result.Status)
	}

	// Class info
	if result.ClassNumber != "06 - 01" {
		t.Errorf("ClassNumber = %q", result.ClassNumber)
	}
	if result.ClassDescription != "Stuhl" {
		t.Errorf("ClassDescription = %q", result.ClassDescription)
	}

	// Applicants (from design level)
	if len(result.Applicants) != 1 {
		t.Fatalf("len(Applicants) = %d, want 1", len(result.Applicants))
	}
	if result.Applicants[0].Name != "MT Designs" {
		t.Errorf("Applicants[0].Name = %q", result.Applicants[0].Name)
	}
}

func TestParseDesignInfo_MalformedXML(t *testing.T) {
	_, err := ParseDesignInfo([]byte("xyz"))
	if err == nil {
		t.Fatal("expected error")
	}
	var xmlErr *XMLParseError
	if !errors.As(err, &xmlErr) {
		t.Errorf("expected *XMLParseError, got %T", err)
	}
}

func TestParseTrademarkSearch_Empty(t *testing.T) {
	xml := []byte(`<?xml version="1.0" encoding="UTF-8"?><HitList HitCount="0"/>`)
	result, err := ParseTrademarkSearch(xml)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if result.TotalHits != 0 {
		t.Errorf("TotalHits = %d, want 0", result.TotalHits)
	}
	if len(result.Hits) != 0 {
		t.Errorf("len(Hits) = %d, want 0", len(result.Hits))
	}
}

func TestParseDesignSearch_Empty(t *testing.T) {
	xml := []byte(`<?xml version="1.0" encoding="UTF-8"?><DesignHitList HitCount="0"/>`)
	result, err := ParseDesignSearch(xml)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if result.TotalHits != 0 {
		t.Errorf("TotalHits = %d, want 0", result.TotalHits)
	}
	if len(result.Hits) != 0 {
		t.Errorf("len(Hits) = %d, want 0", len(result.Hits))
	}
}

func TestParsePatentSearch_BOMPrefixed(t *testing.T) {
	bom := "\xEF\xBB\xBF"
	xml := []byte(bom + `<?xml version="1.0" encoding="UTF-8"?><PatentHitList HitCount="1"><PatentHitListRecord><leading-registered-number>12345</leading-registered-number><invention-title>Test</invention-title></PatentHitListRecord></PatentHitList>`)
	result, err := ParsePatentSearch(xml)
	if err != nil {
		t.Fatalf("BOM-prefixed XML should parse, got error = %v", err)
	}
	if result.TotalHits != 1 {
		t.Errorf("TotalHits = %d, want 1", result.TotalHits)
	}
}

func TestXMLParseError_Unwrap(t *testing.T) {
	inner := errors.New("underlying error")
	xmlErr := &XMLParseError{Operation: "TestOp", Err: inner}
	if !errors.Is(xmlErr, inner) {
		t.Error("Unwrap should return inner error")
	}
	if xmlErr.Error() != "TestOp: failed to parse XML: underlying error" {
		t.Errorf("Error() = %q", xmlErr.Error())
	}
}
