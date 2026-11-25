package dpmaconnect

import (
	"os"
	"testing"
)

func mustReadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// TestPatentHitNumberSystem checks the classification against the
// leading-registered-number / registered-number pairs of the recorded searches:
// the registered-number is the register's own human-readable form of the same
// file, so it tells which system the leading number belongs to.
func TestPatentHitNumberSystem(t *testing.T) {
	tests := []struct {
		fixture    string
		leading    string
		registered string
		system     PatentNumberSystem
		country    string
		office     string
	}{
		{"patent_search.xml", "82337144", "G 82 33 714.4", NumberSystemDE, "DE", ""},
		{"patent_search.xml", "27258373", "P 27 25 837.3", NumberSystemDE, "DE", ""},
		{"patent_search_limited.xml", "DD136869", "136 869", NumberSystemDD, "DD", ""},
		{"patent_search_limited.xml", "DD137459", "137 459", NumberSystemDD, "DD", ""},
		{"patent_search_ep.xml", "E247834666", "24 78 3466.6", NumberSystemEP, "EP", ""},
		{"patent_search_pct.xml", "CN2019089592", "PCT/CN2019/089592", NumberSystemPCT, "WO", "CN"},
	}
	for _, tt := range tests {
		t.Run(tt.leading, func(t *testing.T) {
			res, err := ParsePatentSearch(mustReadFixture(t, tt.fixture))
			if err != nil {
				t.Fatal(err)
			}
			for _, h := range res.Hits {
				if h.LeadingRegisteredNumber != tt.leading {
					continue
				}
				if h.RegisteredNumber != tt.registered {
					t.Fatalf("RegisteredNumber = %q, want %q", h.RegisteredNumber, tt.registered)
				}
				if h.NumberSystem != tt.system || h.Country != tt.country || h.ReceivingOffice != tt.office {
					t.Errorf("got system %q country %q office %q, want %q %q %q",
						h.NumberSystem, h.Country, h.ReceivingOffice, tt.system, tt.country, tt.office)
				}
				return
			}
			t.Fatalf("hit %s not in %s", tt.leading, tt.fixture)
		})
	}
}

// TestRegisteredNumberCountryMatchesRegister ties the Country of each number
// system to what the register record of such a file states itself.
func TestRegisteredNumberCountryMatchesRegister(t *testing.T) {
	// EP file: the register's own application reference is an EP number.
	ep, err := ParsePatentInfo(mustReadFixture(t, "patent_info_ep_application.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if got := ClassifyRegisteredNumber("E247834666").Country; got != ep.ApplicationRef.Country {
		t.Errorf("EP: classified %q, register says %q", got, ep.ApplicationRef.Country)
	}
	// PCT file: application reference country WO.
	pct, err := ParsePatentInfo(mustReadFixture(t, "patent_info_pct.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if got := ClassifyRegisteredNumber("CN2019089592").Country; got != pct.ApplicationRef.Country {
		t.Errorf("PCT: classified %q, register says %q", got, pct.ApplicationRef.Country)
	}
	// GDR patent: its publications carry country DD.
	dd, err := ParsePatentInfo(mustReadFixture(t, "patent_info_dd.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if got := ClassifyRegisteredNumber("DD136869").Country; len(dd.Publications) == 0 || got != dd.Publications[0].Country {
		t.Errorf("DD: classified %q, register publications %+v", got, dd.Publications)
	}
	// German file.
	de, err := ParsePatentInfo(mustReadFixture(t, "patent_info.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if got := ClassifyRegisteredNumber("100273602").Country; got != de.ApplicationRef.Country {
		t.Errorf("DE: classified %q, register says %q", got, de.ApplicationRef.Country)
	}
}

func TestClassifyRegisteredNumberUnrecognised(t *testing.T) {
	// Forms the recordings do not contain must not be guessed.
	for _, in := range []string{"", "PCT/CN2019/089592", "24 78 3466.6", "E", "DE10027362C2", "X123"} {
		if got := ClassifyRegisteredNumber(in); got != (RegisteredNumberClass{}) {
			t.Errorf("ClassifyRegisteredNumber(%q) = %+v, want zero value", in, got)
		}
	}
}

// TestPublicationMatchesNumber uses the publication references of the recorded
// register records: the register zero-pads the doc-number, callers do not.
func TestPublicationMatchesNumber(t *testing.T) {
	tests := []struct {
		fixture   string
		docNumber string // as recorded
		kind      string
		requested string
		want      bool
	}{
		{"patent_info.xml", "000010027360", "A1", "DE10027360A1", true},
		{"patent_info.xml", "000010027360", "A1", "de 100 27 360 a1", true},
		{"patent_info.xml", "000010027360", "A1", "DE10027360", true},
		{"patent_info.xml", "000010027360", "A1", "10027360", true},
		{"patent_info.xml", "000010027360", "A1", "DE000010027360A1", true},
		{"patent_info.xml", "000010027360", "A1", "DE10027360B4", false},
		{"patent_info.xml", "000010027360", "A1", "DE10027362A1", false},
		{"patent_info.xml", "000010027360", "A1", "EP10027360A1", false},
		{"patent_info.xml", "000010027360", "A1", "DE1002736A1", false},
		{"patent_info_de_pct_national.xml", "112019001234", "T5", "DE112019001234T5", true},
		{"patent_info_de_related_documents.xml", "102019003253", "A1", "DE102019003253A1", true},
		{"patent_info_dd.xml", "000000136869", "A1", "DD136869A1", true},
		{"patent_info_dd.xml", "000000136869", "B1", "DD136869B1", true},
		{"patent_info_dd.xml", "000000136869", "B1", "DD136869A1", false},
		{"patent_info_dd.xml", "000000136869", "B1", "DE136869", false},
	}
	for _, tt := range tests {
		t.Run(tt.docNumber+tt.kind+"_"+tt.requested, func(t *testing.T) {
			info, err := ParsePatentInfo(mustReadFixture(t, tt.fixture))
			if err != nil {
				t.Fatal(err)
			}
			for _, p := range info.Publications {
				if p.Number == tt.docNumber && p.Kind == tt.kind {
					if got := p.MatchesNumber(tt.requested); got != tt.want {
						t.Errorf("%+v MatchesNumber(%q) = %v, want %v", p, tt.requested, got, tt.want)
					}
					return
				}
			}
			t.Fatalf("publication %s %s not in %s", tt.docNumber, tt.kind, tt.fixture)
		})
	}
}

// TestRegionalPublicationMatchesNumber covers the EP and WO publications the
// register lists without a kind code (patent_info_ep_granted.xml).
func TestRegionalPublicationMatchesNumber(t *testing.T) {
	info, err := ParsePatentInfo(mustReadFixture(t, "patent_info_ep_granted.xml"))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{ // requested -> matching "country number"
		"EP3500000B1":    "EP 3500000",
		"EP 3 500 000":   "EP 3500000",
		"WO2018053814A1": "WO 2018053814",
		"EP350000":       "",
		"WO3500000":      "",
		"EP2018053814":   "",
	}
	for requested, match := range want {
		got := ""
		for _, p := range info.RegionalPublications {
			if p.MatchesPublicationNumber(requested) {
				got = p.Country + " " + p.Number
			}
		}
		if got != match {
			t.Errorf("%q matched %q, want %q", requested, got, match)
		}
	}
}

// TestParsePublicationNumber covers the combined form the register writes into
// published-documents and references-cited (values from patent_info_dd.xml and
// patent_info_de_related_documents.xml).
func TestParsePublicationNumber(t *testing.T) {
	dd, err := ParsePatentInfo(mustReadFixture(t, "patent_info_dd.xml"))
	if err != nil {
		t.Fatal(err)
	}
	rel, err := ParsePatentInfo(mustReadFixture(t, "patent_info_de_related_documents.xml"))
	if err != nil {
		t.Fatal(err)
	}
	var recorded []string
	for _, e := range dd.Events {
		recorded = append(recorded, e.PublishedDocuments...)
	}
	for _, c := range rel.Citations {
		recorded = append(recorded, c.DocNumber)
	}
	want := map[string]PublicationNumber{
		"DD000000136869A1": {Country: "DD", Number: "136869", Kind: "A1"},
		"DD000000136869B1": {Country: "DD", Number: "136869", Kind: "B1"},
		"EP000001463074A1": {Country: "EP", Number: "1463074", Kind: "A1"},
		"DE102019103531A1": {Country: "DE", Number: "102019103531", Kind: "A1"},
	}
	seen := 0
	for _, r := range recorded {
		w, ok := want[r]
		if !ok {
			continue
		}
		seen++
		got, ok := ParsePublicationNumber(r)
		if !ok || got != w {
			t.Errorf("ParsePublicationNumber(%q) = %+v, %v; want %+v", r, got, ok, w)
		}
	}
	if seen != len(want) {
		t.Fatalf("found %d of %d expected recorded values in %v", seen, len(want), recorded)
	}
	for _, bad := range []string{"", "DE", "PCT/CN2019/089592", "10 2019 003 253.5", "DE10027360A12"} {
		if got, ok := ParsePublicationNumber(bad); ok {
			t.Errorf("ParsePublicationNumber(%q) = %+v, want not ok", bad, got)
		}
	}
}
