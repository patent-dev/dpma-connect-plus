package dpmaconnect

import (
	"reflect"
	"testing"
)

// The fixtures below were recorded after TestXMLCompleteness was written; they
// get the same element-coverage and round-trip layers here.
func TestXMLCompletenessFurtherPatentInfo(t *testing.T) {
	for _, name := range []string{"patent_info_dd", "patent_info_de_related_documents"} {
		t.Run(name, func(t *testing.T) {
			data := mustReadFixture(t, name+".xml")
			assertComplete(t, name, data, reflect.TypeOf(xmlDPMAPatentDocument{}))
			assertRoundTrip(t, name, data, reflect.TypeOf(xmlDPMAPatentDocument{}))
		})
	}
}

func TestParsePatentInfoRelatedDocumentsAndCitations(t *testing.T) {
	info, err := ParsePatentInfo(mustReadFixture(t, "patent_info_de_related_documents.xml"))
	if err != nil {
		t.Fatal(err)
	}
	wantRelations := []PatentRelation{
		{
			Type:    RelationUtilityModelBasis,
			Parents: []PatentDocumentRef{{Country: "DE", Number: "1020190032535"}},
			Children: []PatentDocumentRef{
				{Country: "DE", Number: "2020191033215"},
				{Country: "DE", Number: "2020190056556"},
			},
		},
		{
			Type:     RelationDivision,
			Parents:  []PatentDocumentRef{{Country: "DE", Number: "1020191067375"}},
			Children: []PatentDocumentRef{{Country: "DE", Number: "1020190032535"}},
		},
	}
	if !reflect.DeepEqual(info.RelatedDocuments, wantRelations) {
		t.Errorf("RelatedDocuments = %+v\nwant %+v", info.RelatedDocuments, wantRelations)
	}
	wantCitations := []PatentCitation{
		{Sequence: "1", DocNumber: "EP000001463074A1", RelevanceIndicators: []string{"X"}},
		{Sequence: "2", DocNumber: "DE102019103531A1", RelevanceIndicators: []string{"X"}},
	}
	if !reflect.DeepEqual(info.Citations, wantCitations) {
		t.Errorf("Citations = %+v\nwant %+v", info.Citations, wantCitations)
	}
}

func TestParsePatentInfoEventParties(t *testing.T) {
	info, err := ParsePatentInfo(mustReadFixture(t, "patent_info_de_related_documents.xml"))
	if err != nil {
		t.Fatal(err)
	}
	var applicantChange, repChange *PatentEvent
	for i := range info.Events {
		e := &info.Events[i]
		if e.TypeOfProcedure == "anmelderaenderung" && len(e.PreviousApplicants) > 0 && applicantChange == nil {
			applicantChange = e
		}
		if e.PreviousRepresentative != "" && repChange == nil {
			repChange = e
		}
	}
	if applicantChange == nil || repChange == nil {
		t.Fatalf("expected an applicant change and a representative change event, got %+v", info.Events)
	}
	if got, want := applicantChange.Applicants, []string{"DEHN SE + Co KG, 92318 Neumarkt, DE"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Applicants = %q, want %q", got, want)
	}
	if got, want := applicantChange.PreviousApplicants, []string{"DEHN + SÖHNE GmbH + Co. KG, 92318 Neumarkt, DE"}; !reflect.DeepEqual(got, want) {
		t.Errorf("PreviousApplicants = %q, want %q", got, want)
	}
	wantParty := []Party{{
		Text:    "DEHN + SÖHNE GmbH + Co. KG, 92318 Neumarkt, DE",
		Name:    "DEHN + SÖHNE GmbH + Co. KG",
		Address: "92318 Neumarkt",
		Country: "DE",
	}}
	if !reflect.DeepEqual(applicantChange.PreviousApplicantParties, wantParty) {
		t.Errorf("PreviousApplicantParties = %+v, want %+v", applicantChange.PreviousApplicantParties, wantParty)
	}
	if len(applicantChange.ApplicantParties) != 1 || applicantChange.ApplicantParties[0].Name != "DEHN SE + Co KG" {
		t.Errorf("ApplicantParties = %+v", applicantChange.ApplicantParties)
	}
	if repChange.PreviousRepresentativeType == "" || len(repChange.PreviousRepresentativeParties) == 0 ||
		repChange.PreviousRepresentativeParties[0].Text != repChange.PreviousRepresentative {
		t.Errorf("representative change not carried: %+v", *repChange)
	}
}

// A GDR patent: DD publications, a DE application reference with the GDR file
// number, and an applicant change event with both applicant lists.
func TestParsePatentInfoDD(t *testing.T) {
	info, err := ParsePatentInfo(mustReadFixture(t, "patent_info_dd.xml"))
	if err != nil {
		t.Fatal(err)
	}
	wantPubs := []PatentPublication{
		{Sequence: "1", Country: "DD", Number: "000000136869", Kind: "A1"},
		{Sequence: "2", Country: "DD", Number: "000000136869", Kind: "B1"},
	}
	if !reflect.DeepEqual(info.Publications, wantPubs) {
		t.Errorf("Publications = %+v, want %+v", info.Publications, wantPubs)
	}
	if want := (PatentDocumentRef{Country: "DE", Number: "DD2053842", Date: "1978-05-16"}); info.ApplicationRef != want {
		t.Errorf("ApplicationRef = %+v, want %+v", info.ApplicationRef, want)
	}
	if len(info.Events) == 0 || len(info.Events[0].Applicants) != 2 || len(info.Events[0].ApplicantParties) != 2 ||
		info.Events[0].ApplicantParties[1].Name != "Siemens AG Berlin u. München" {
		t.Errorf("first event applicants not carried: %+v", info.Events)
	}
}

// The capped result lists carry the register's notice in attributes and in the
// DocumentHits counter text.
func TestParseSearchNoticeVerbatim(t *testing.T) {
	p, err := ParsePatentSearch(mustReadFixture(t, "patent_search_limited.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if p.MessageEN != "Result list is limited  - Not all results are returned" ||
		p.MessageDE != "Trefferliste wurde begrenzt  - Nicht alle Ergebnisse wurden zurückgeliefert." {
		t.Errorf("patent messages = %q / %q", p.MessageEN, p.MessageDE)
	}
	if !p.ListMaxHitsReached || !p.DocumentHitsCounter.MaxHitsReached || !p.DatabaseHitsCounter.MaxHitsReached ||
		p.DatabaseHitsCounter.Text != "129659" ||
		p.DocumentHitsCounter.Text != `Message_EN="Result list is limited  - Not all results are returned" Message_DE="Trefferliste wurde begrenzt  - Nicht alle Ergebnisse wurden zurückgeliefert."` {
		t.Errorf("patent counters = %+v / %+v (list flag %v)", p.DocumentHitsCounter, p.DatabaseHitsCounter, p.ListMaxHitsReached)
	}
	d, err := ParseDesignSearch(mustReadFixture(t, "design_search_limited.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if d.MessageEN != "Result list is limited  - Not all results as returned" || !d.ListMaxHitsReached {
		t.Errorf("design message = %q (list flag %v)", d.MessageEN, d.ListMaxHitsReached)
	}
	tm, err := ParseTrademarkSearch(mustReadFixture(t, "trademark_search_limited.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if !tm.ListMaxHitsReached || !tm.DocumentHitsCounter.MaxHitsReached || tm.DocumentHitsCounter.Text == "" {
		t.Errorf("trademark counters = %+v (list flag %v)", tm.DocumentHitsCounter, tm.ListMaxHitsReached)
	}
}

func TestParseTrademarkAndDesignInfoAddedFields(t *testing.T) {
	tm, err := ParseTrademarkInfo(mustReadFixture(t, "trademark_info.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if tm.TransactionCode != "Trade Mark Information" || tm.TransactionIdentifier != "9818831573854928984" ||
		!reflect.DeepEqual(tm.PublicationDates, []string{"1974-07-31"}) {
		t.Errorf("trademark envelope = %q %q %q", tm.TransactionCode, tm.TransactionIdentifier, tm.PublicationDates)
	}
	var corr *Party
	for _, r := range tm.Records {
		if r.NewCorrespondenceParty != nil {
			corr = r.NewCorrespondenceParty
		}
	}
	if corr == nil || corr.Name != "Joh. Wilh. von Eicken GmbH" || corr.Street != "Drechslerstr. 1-3" ||
		corr.Address != "Lübeck" || corr.Postcode != "23556" || corr.County != "DE" {
		t.Errorf("NewCorrespondenceParty = %+v", corr)
	}

	d, err := ParseDesignInfo(mustReadFixture(t, "design_info.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if d.TransactionCode != "Design E-Filing" || len(d.ApplicationApplicants) != 1 ||
		d.ApplicationApplicants[0].Name != "MT Designs" || d.ApplicationApplicants[0].Address != "Greensboro" ||
		d.ApplicationApplicants[0].County != "N.C." || d.ApplicationApplicants[0].State != "US" {
		t.Errorf("design application level = %q %+v", d.TransactionCode, d.ApplicationApplicants)
	}
	if d.CorrespondenceParty.PostOfficeBox != "Abholfach beim DPMA" {
		t.Errorf("CorrespondenceParty.PostOfficeBox = %q", d.CorrespondenceParty.PostOfficeBox)
	}
	if len(d.Designs) == 0 || !reflect.DeepEqual(d.Designs[0].Publications, []DesignPublication{{Identifier: "1", Date: "1988-11-10"}}) {
		t.Errorf("design publications = %+v", d.Designs)
	}
}
