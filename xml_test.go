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

//go:embed testdata/trademark_search_error.xml
var trademarkSearchErrorXML []byte

//go:embed testdata/design_search_error.xml
var designSearchErrorXML []byte

func TestParsePatentSearch(t *testing.T) {
	result, err := ParsePatentSearch(patentSearchXML)
	if err != nil {
		t.Fatalf("ParsePatentSearch() error = %v", err)
	}

	if result.TotalHits != 4 {
		t.Errorf("TotalHits = %d, want 4", result.TotalHits)
	}
	if len(result.Hits) != 4 {
		t.Fatalf("len(Hits) = %d, want 4", len(result.Hits))
	}
	if result.DocumentHits != 3923 {
		t.Errorf("DocumentHits = %d, want 3923", result.DocumentHits)
	}
	if result.DatabaseHits != 3923 {
		t.Errorf("DatabaseHits = %d, want 3923", result.DatabaseHits)
	}

	if result.RawXML == nil {
		t.Error("RawXML is nil")
	}
	if len(result.RawXML) != len(patentSearchXML) {
		t.Errorf("RawXML length = %d, want %d", len(result.RawXML), len(patentSearchXML))
	}

	h := result.Hits[0]
	if h.Agent != "Trappenberg, Hans, Pat.-Ing., 76185 Karlsruhe, DE" {
		t.Errorf("Hits[0].Agent = %q", h.Agent)
	}
	if h.MainClassification.Edition != "1985.01" {
		t.Errorf("Hits[0].MainClassification.Edition = %q", h.MainClassification.Edition)
	}
	// Fourth (rich) hit carries inventors.
	h3 := result.Hits[3]
	if len(h3.Inventors) != 1 || h3.Inventors[0] != "Ihle, Bernhard, W-7520 Bruchsal, DE" {
		t.Errorf("Hits[3].Inventors = %v", h3.Inventors)
	}
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

	// RawXML retained on info results.
	if len(result.RawXML) != len(patentInfoXML) {
		t.Errorf("RawXML length = %d, want %d", len(result.RawXML), len(patentInfoXML))
	}

	// Agents (representatives) and correspondence party.
	if len(result.Agents) != 1 || result.Agents[0].Name != "KUHNEN & WACKER Patent- und Rechtsanwaltsbüro" {
		t.Errorf("Agents = %+v", result.Agents)
	}
	if result.CorrespondenceParty.Country != "DE" {
		t.Errorf("CorrespondenceParty.Country = %q", result.CorrespondenceParty.Country)
	}
	if result.AgentType != "vertreter" {
		t.Errorf("AgentType = %q", result.AgentType)
	}
	if result.DateLastRegisterUpdate != "2017-06-21" {
		t.Errorf("DateLastRegisterUpdate = %q", result.DateLastRegisterUpdate)
	}

	// Priority claims.
	if len(result.PriorityClaims) != 1 {
		t.Fatalf("len(PriorityClaims) = %d, want 1", len(result.PriorityClaims))
	}
	pc := result.PriorityClaims[0]
	if pc.Country != "JP" || pc.Date != "1999-06-04" || pc.DocNumber != "11158010" {
		t.Errorf("PriorityClaims[0] = %+v", pc)
	}

	// Legal-event / prosecution history tree.
	if len(result.Events) != 5 {
		t.Fatalf("len(Events) = %d, want 5", len(result.Events))
	}
	if result.Events[0].TypeOfProcedure != "vorverfahren" || result.Events[0].ProceduralStatus != "R000" {
		t.Errorf("Events[0] = %+v", result.Events[0])
	}
	// Event with publication-info.
	ev := result.Events[2]
	if ev.PublicationInfo == nil || ev.PublicationInfo.IssueNumber != "16" || ev.PublicationInfo.Year != "2007" {
		t.Errorf("Events[2].PublicationInfo = %+v", ev.PublicationInfo)
	}
	// Event with ipc-main-class / previous-ipc-main-class.
	ev = result.Events[3]
	if ev.IPCMainClass == nil || ev.IPCMainClass.Version != "2006.01" {
		t.Errorf("Events[3].IPCMainClass = %+v", ev.IPCMainClass)
	}
	if ev.PreviousIPCMainClass == nil || ev.PreviousIPCMainClass.Version != "2000.01" {
		t.Errorf("Events[3].PreviousIPCMainClass = %+v", ev.PreviousIPCMainClass)
	}
	// Event with published-documents.
	ev = result.Events[4]
	if len(ev.PublishedDocuments) != 1 || ev.PublishedDocuments[0] != "DE000010027360A1" {
		t.Errorf("Events[4].PublishedDocuments = %v", ev.PublishedDocuments)
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

	if result.TotalHits != 4 {
		t.Errorf("TotalHits = %d, want 4", result.TotalHits)
	}
	if len(result.Hits) != 4 {
		t.Fatalf("len(Hits) = %d, want 4", len(result.Hits))
	}

	if result.RawXML == nil {
		t.Error("RawXML is nil")
	}

	h := result.Hits[0]
	if h.Number != "0" {
		t.Errorf("Hits[0].Number = %q", h.Number)
	}
	if h.RegistrationOfficeCode != "DE" {
		t.Errorf("Hits[0].RegistrationOfficeCode = %q", h.RegistrationOfficeCode)
	}
	if h.MarkDescriptionText != "von Eicken`s Apple and Grape" {
		t.Errorf("Hits[0].MarkDescriptionText = %q", h.MarkDescriptionText)
	}
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
	if result.Classifications[0].DescriptionLang != "DE" {
		t.Errorf("Classifications[0].DescriptionLang = %q", result.Classifications[0].DescriptionLang)
	}

	// RawXML retained.
	if len(result.RawXML) != len(trademarkInfoXML) {
		t.Errorf("RawXML length = %d, want %d", len(result.RawXML), len(trademarkInfoXML))
	}

	// Newly modeled scalar fields.
	if result.RegistrationOfficeCode != "DE" {
		t.Errorf("RegistrationOfficeCode = %q", result.RegistrationOfficeCode)
	}
	if result.TerminationDate != "2014-02-01" {
		t.Errorf("TerminationDate = %q", result.TerminationDate)
	}
	if result.KindMark != "Individual" {
		t.Errorf("KindMark = %q", result.KindMark)
	}
	if result.OppositionPeriodStartDate != "1974-07-31" {
		t.Errorf("OppositionPeriodStartDate = %q", result.OppositionPeriodStartDate)
	}
	if result.ProposedLeadingClassNumber != "34" {
		t.Errorf("ProposedLeadingClassNumber = %q", result.ProposedLeadingClassNumber)
	}
	if result.StandardCharacterIndicator != "true" {
		t.Errorf("StandardCharacterIndicator = %q", result.StandardCharacterIndicator)
	}
	if result.MarkTextLang != "DE" {
		t.Errorf("MarkTextLang = %q", result.MarkTextLang)
	}
	if result.FileNumber != "919224" {
		t.Errorf("FileNumber = %q", result.FileNumber)
	}
	if result.MarkFeatureDPMA != "Word" {
		t.Errorf("MarkFeatureDPMA = %q", result.MarkFeatureDPMA)
	}
	if result.CorrespondenceParty.Street != "Drechslerstr. 1-3" {
		t.Errorf("CorrespondenceParty.Street = %q", result.CorrespondenceParty.Street)
	}

	// Register-event (MarkRecord) history tree.
	if len(result.Records) != 6 {
		t.Fatalf("len(Records) = %d, want 6", len(result.Records))
	}
	// First record is the cancellation (Record5f).
	if result.Records[0].BasicRecordKind != "Loeschung Inhaber" {
		t.Errorf("Records[0].BasicRecordKind = %q", result.Records[0].BasicRecordKind)
	}
	if result.Records[0].LegalGround == "" {
		t.Error("Records[0].LegalGround is empty (de:Record5f dropped)")
	}
	if result.Records[0].CancellationDate != "2014-02-01" {
		t.Errorf("Records[0].CancellationDate = %q", result.Records[0].CancellationDate)
	}
	if result.Records[0].PublicationIdentifier != "2014-51" {
		t.Errorf("Records[0].PublicationIdentifier = %q", result.Records[0].PublicationIdentifier)
	}
	// Find the ownership-transfer (Umschreibung) record with previous/new holders.
	var transfer *TrademarkRecord
	for i := range result.Records {
		if result.Records[i].BasicRecordKind == "Umschreibung" {
			transfer = &result.Records[i]
		}
	}
	if transfer == nil {
		t.Fatal("no Umschreibung record found")
	}
	if transfer.ReceiptDeclarationDate != "2002-12-13" {
		t.Errorf("transfer.ReceiptDeclarationDate = %q", transfer.ReceiptDeclarationDate)
	}
	if transfer.PreviousHolder == nil || transfer.PreviousHolder.Name != "Joh. Wilh. von Eicken GmbH & Co." {
		t.Errorf("transfer.PreviousHolder = %+v", transfer.PreviousHolder)
	}
	if transfer.NewHolder == nil || transfer.NewHolder.Name != "Joh. Wilh. von Eicken GmbH" {
		t.Errorf("transfer.NewHolder = %+v", transfer.NewHolder)
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

	if result.TotalHits != 4 {
		t.Errorf("TotalHits = %d, want 4", result.TotalHits)
	}
	if len(result.Hits) != 4 {
		t.Fatalf("len(Hits) = %d, want 4", len(result.Hits))
	}

	if result.RawXML == nil {
		t.Error("RawXML is nil")
	}

	h := result.Hits[0]
	if h.TotalRepresentationSheet != "7" {
		t.Errorf("Hits[0].TotalRepresentationSheet = %q", h.TotalRepresentationSheet)
	}
	if h.Representative == "" {
		t.Error("Hits[0].Representative is empty")
	}
	if h.FirstRepresentationSheetFilename == "" {
		t.Error("Hits[0].FirstRepresentationSheetFilename is empty")
	}
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
	// Fourth hit carries the (often-dropped) StaffName field.
	if result.Hits[3].StaffName == "" {
		t.Error("Hits[3].StaffName is empty (field dropped)")
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

	// RawXML retained.
	if len(result.RawXML) != len(designInfoXML) {
		t.Errorf("RawXML length = %d, want %d", len(result.RawXML), len(designInfoXML))
	}

	// Application-level metadata.
	if result.RegistrationOfficeCode != "DE" {
		t.Errorf("RegistrationOfficeCode = %q", result.RegistrationOfficeCode)
	}
	if result.ApplicationReference != "M8800174.1" {
		t.Errorf("ApplicationReference = %q", result.ApplicationReference)
	}
	if result.TotalDesign != "1" {
		t.Errorf("TotalDesign = %q", result.TotalDesign)
	}

	// Full Designs slice (multi-design safe).
	if len(result.Designs) != 1 {
		t.Fatalf("len(Designs) = %d, want 1", len(result.Designs))
	}
	d := result.Designs[0]
	if d.TotalSpecimen != "2" {
		t.Errorf("Designs[0].TotalSpecimen = %q", d.TotalSpecimen)
	}

	// DE_ST86Extension scalar fields.
	if d.Extension.RegisterNumber != "M8800174" {
		t.Errorf("Extension.RegisterNumber = %q", d.Extension.RegisterNumber)
	}
	if d.Extension.DesignFileNumber != "M8800174.1" {
		t.Errorf("Extension.DesignFileNumber = %q", d.Extension.DesignFileNumber)
	}
	if d.Extension.DateOfCaptureInSystem != "1994-02-01" {
		t.Errorf("Extension.DateOfCaptureInSystem = %q", d.Extension.DateOfCaptureInSystem)
	}
	if d.Extension.TypeOfDeposit != "Representation" {
		t.Errorf("Extension.TypeOfDeposit = %q", d.Extension.TypeOfDeposit)
	}
	if d.Extension.CancellationDate != "1994-02-01" {
		t.Errorf("Extension.CancellationDate = %q", d.Extension.CancellationDate)
	}

	// Image / representation views.
	if d.PreferedView == nil || d.PreferedView.Filename != "/old_bild/174/880017400101.jpg" {
		t.Errorf("Designs[0].PreferedView = %+v", d.PreferedView)
	}
	if d.PreferedView != nil && (d.PreferedView.Height != "985" || d.PreferedView.Width != "753" || d.PreferedView.Unit != "Pixel") {
		t.Errorf("PreferedView size = %+v", d.PreferedView)
	}
	if len(d.Representations) != 1 || d.Representations[0].Number != "1" {
		t.Errorf("Designs[0].Representations = %+v", d.Representations)
	}

	// Classes carry classification metadata.
	if len(d.Classes) != 1 {
		t.Fatalf("len(Designs[0].Classes) = %d, want 1", len(d.Classes))
	}
	if d.Classes[0].ClassificationKindCode != "Locarno" || d.Classes[0].ClassificationVersion != "5" {
		t.Errorf("Classes[0] classification = %+v", d.Classes[0])
	}
	if d.Classes[0].Comment == "" {
		t.Error("Classes[0].Comment is empty")
	}

	// Priorities.
	if len(d.Priorities) != 1 {
		t.Fatalf("len(Priorities) = %d, want 1", len(d.Priorities))
	}
	if d.Priorities[0].CountryCode != "US" || d.Priorities[0].Date != "1988-01-07" {
		t.Errorf("Priorities[0] = %+v", d.Priorities[0])
	}

	// Representatives.
	if len(d.Representatives) != 1 || d.Representatives[0].Name == "" {
		t.Errorf("Designs[0].Representatives = %+v", d.Representatives)
	}

	// Design register-event (DesignRecord) tree.
	if len(d.Records) != 2 {
		t.Fatalf("len(Designs[0].Records) = %d, want 2", len(d.Records))
	}
	if d.Records[0].LegalStatus != "Design Registered" {
		t.Errorf("Records[0].LegalStatus = %q", d.Records[0].LegalStatus)
	}
	if d.Records[0].PublicationIdentifier != "1988-21" || d.Records[0].PublicationSubsection != "Part 1a" {
		t.Errorf("Records[0] = %+v", d.Records[0])
	}

	// Correspondence party (application level).
	if result.CorrespondenceParty.Name == "" {
		t.Error("CorrespondenceParty.Name is empty")
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

func TestDetectErrorXML(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		errType string
	}{
		{
			name:    "valid patent search XML",
			data:    []byte(`<?xml version="1.0" encoding="UTF-8"?><PatentHitList HitCount="0"/>`),
			wantErr: false,
		},
		{
			name:    "non-XML data",
			data:    []byte("not xml at all"),
			wantErr: false,
		},
		{
			name: "patent error - data not available",
			data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Transaction>
  <PatentTransactionBody>
    <TransactionErrorDetails>
      <TransactionError>
        <TransactionErrorCode>E001</TransactionErrorCode>
        <TransactionErrorText>Data not available</TransactionErrorText>
      </TransactionError>
    </TransactionErrorDetails>
  </PatentTransactionBody>
</Transaction>`),
			wantErr: true,
			errType: "DataNotAvailableError",
		},
		{
			name: "trademark error - permission denied",
			data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Transaction>
  <TradeMarkTransactionBody>
    <TransactionErrorDetails>
      <TransactionError>
        <TransactionErrorCode>E002</TransactionErrorCode>
        <TransactionErrorText>Permission denied</TransactionErrorText>
      </TransactionError>
    </TransactionErrorDetails>
  </TradeMarkTransactionBody>
</Transaction>`),
			wantErr: true,
			errType: "APIError",
		},
		{
			name: "design error",
			data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Transaction>
  <DesignTransactionBody>
    <TransactionErrorDetails>
      <TransactionError>
        <TransactionErrorCode>E003</TransactionErrorCode>
        <TransactionErrorText>Resource not found</TransactionErrorText>
      </TransactionError>
    </TransactionErrorDetails>
  </DesignTransactionBody>
</Transaction>`),
			wantErr: true,
			errType: "APIError",
		},
		{
			name: "transaction without error details",
			data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Transaction>
  <TradeMarkTransactionBody>
  </TradeMarkTransactionBody>
</Transaction>`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseDPMAError(tt.data, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDPMAError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				return
			}
			switch tt.errType {
			case "DataNotAvailableError":
				var dna *DataNotAvailableError
				if !errors.As(err, &dna) {
					t.Errorf("expected *DataNotAvailableError, got %T: %v", err, err)
				}
			case "APIError":
				var apiErr *APIError
				if !errors.As(err, &apiErr) {
					t.Errorf("expected *APIError, got %T: %v", err, err)
				}
			}
		})
	}
}

func TestParsePatentSearch_ErrorXML(t *testing.T) {
	errorXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Transaction>
  <PatentTransactionBody>
    <TransactionErrorDetails>
      <TransactionError>
        <TransactionErrorCode>E001</TransactionErrorCode>
        <TransactionErrorText>Data not available</TransactionErrorText>
      </TransactionError>
    </TransactionErrorDetails>
  </PatentTransactionBody>
</Transaction>`)

	_, err := ParsePatentSearch(errorXML)
	if err == nil {
		t.Fatal("expected error for error XML")
	}
	var dna *DataNotAvailableError
	if !errors.As(err, &dna) {
		t.Errorf("expected *DataNotAvailableError, got %T: %v", err, err)
	}
}

func TestParsePatentInfo_ErrorXML(t *testing.T) {
	errorXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Transaction>
  <PatentTransactionBody>
    <TransactionErrorDetails>
      <TransactionError>
        <TransactionErrorCode>E002</TransactionErrorCode>
        <TransactionErrorText>Permission denied</TransactionErrorText>
      </TransactionError>
    </TransactionErrorDetails>
  </PatentTransactionBody>
</Transaction>`)

	_, err := ParsePatentInfo(errorXML)
	if err == nil {
		t.Fatal("expected error for error XML")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Code != "E002" {
		t.Errorf("error code = %q, want %q", apiErr.Code, "E002")
	}
}

func TestParseTrademarkInfo_ErrorXML(t *testing.T) {
	errorXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Transaction>
  <TradeMarkTransactionBody>
    <TransactionErrorDetails>
      <TransactionError>
        <TransactionErrorCode>E001</TransactionErrorCode>
        <TransactionErrorText>Data not available</TransactionErrorText>
      </TransactionError>
    </TransactionErrorDetails>
  </TradeMarkTransactionBody>
</Transaction>`)

	_, err := ParseTrademarkInfo(errorXML)
	if err == nil {
		t.Fatal("expected error for error XML")
	}
	var dna *DataNotAvailableError
	if !errors.As(err, &dna) {
		t.Errorf("expected *DataNotAvailableError, got %T: %v", err, err)
	}
}

func TestParseTrademarkSearch_ErrorXML(t *testing.T) {
	errorXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Transaction>
  <TradeMarkTransactionBody>
    <TransactionErrorDetails>
      <TransactionError>
        <TransactionErrorCode>E001</TransactionErrorCode>
        <TransactionErrorText>Data not available</TransactionErrorText>
      </TransactionError>
    </TransactionErrorDetails>
  </TradeMarkTransactionBody>
</Transaction>`)

	_, err := ParseTrademarkSearch(errorXML)
	if err == nil {
		t.Fatal("expected error for error XML")
	}
	var dna *DataNotAvailableError
	if !errors.As(err, &dna) {
		t.Errorf("expected *DataNotAvailableError, got %T: %v", err, err)
	}
}

func TestParseDesignSearch_ErrorXML(t *testing.T) {
	errorXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Transaction>
  <DesignTransactionBody>
    <TransactionErrorDetails>
      <TransactionError>
        <TransactionErrorCode>E003</TransactionErrorCode>
        <TransactionErrorText>Resource not found</TransactionErrorText>
      </TransactionError>
    </TransactionErrorDetails>
  </DesignTransactionBody>
</Transaction>`)

	_, err := ParseDesignSearch(errorXML)
	if err == nil {
		t.Fatal("expected error for error XML")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestParseDesignInfo_ErrorXML(t *testing.T) {
	errorXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Transaction>
  <DesignTransactionBody>
    <TransactionErrorDetails>
      <TransactionError>
        <TransactionErrorCode>E003</TransactionErrorCode>
        <TransactionErrorText>Resource not found</TransactionErrorText>
      </TransactionError>
    </TransactionErrorDetails>
  </DesignTransactionBody>
</Transaction>`)

	_, err := ParseDesignInfo(errorXML)
	if err == nil {
		t.Fatal("expected error for error XML")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

// TestParseSearch_HitListError verifies that hit-list error envelopes are
// recognized as errors instead of silently parsing to zero hits. The trademark
// search service reports a bad query as <HitList><ErrorMessage>...</ErrorMessage>
// and the design search service as <DesignHitList HitCount="0" Message_DE=.../>.
func TestParseSearch_HitListError(t *testing.T) {
	t.Run("trademark HitList ErrorMessage", func(t *testing.T) {
		_, err := ParseTrademarkSearch(trademarkSearchErrorXML)
		if err == nil {
			t.Fatal("expected error for HitList/ErrorMessage envelope")
		}
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected *APIError, got %T: %v", err, err)
		}
		if apiErr.Message == "" {
			t.Error("APIError.Message is empty")
		}
	})

	t.Run("design DesignHitList Message_DE", func(t *testing.T) {
		_, err := ParseDesignSearch(designSearchErrorXML)
		if err == nil {
			t.Fatal("expected error for DesignHitList Message_DE envelope")
		}
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected *APIError, got %T: %v", err, err)
		}
		if apiErr.Message == "" {
			t.Error("APIError.Message is empty")
		}
	})

	// A genuinely empty hit list must NOT be flagged as an error.
	t.Run("empty design hit list is not an error", func(t *testing.T) {
		ok := []byte(`<?xml version="1.0" encoding="UTF-8"?><DesignHitList HitCount="0"><Counter><DocumentHits>0</DocumentHits><DatabaseHits>0</DatabaseHits></Counter></DesignHitList>`)
		result, err := ParseDesignSearch(ok)
		if err != nil {
			t.Fatalf("empty hit list should not error, got %v", err)
		}
		if result.TotalHits != 0 {
			t.Errorf("TotalHits = %d, want 0", result.TotalHits)
		}
	})
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
