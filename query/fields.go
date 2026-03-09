// Package query provides a query parser and validator for DPMAregister expert search syntax.
//
// DPMAregister uses an expert search syntax with field codes like TI=Elektrofahrzeug,
// Boolean operators (AND/UND, OR/ODER, NOT/NICHT), comparison operators (=, >=, <=, >, <),
// wildcards (? any chars, ! one char, # one or no char), quoted values, and parentheses.
//
// Procedure data searches use curly braces: {VST=pub-offenlegungschrift UND VSTT=05.01.2011}
//
// Each DPMA service (Patent, Design, Trademark) supports different field codes.
// Use [ParseQuery] to parse and validate queries, optionally scoped to a specific service.
package query

import "slices"

// Service represents a DPMAregister service type for field validation.
type Service string

const (
	// ServicePatent validates against patent and utility model fields (Schutzrechtsart P, G).
	ServicePatent Service = "patent"

	// ServiceDesign validates against design fields (Schutzrechtsart G for Geschmacksmuster).
	ServiceDesign Service = "design"

	// ServiceTrademark validates against trademark fields (Marke).
	ServiceTrademark Service = "trademark"

	// ServiceAny accepts fields from any service (no service-specific validation).
	ServiceAny Service = ""
)

// InputType describes the expected input format for a field.
type InputType string

// InputType constants for DPMAregister field input formats.
const (
	InputText InputType = "text"
	InputDate InputType = "date" // dd.mm.yyyy format
)

// Field describes a DPMAregister search field.
type Field struct {
	// Code is the field abbreviation used in queries (e.g. "TI", "INH").
	Code string

	// Name is the German field name (e.g. "Bezeichnung/Titel").
	Name string

	// Description is a brief English description.
	Description string

	// Input is the expected input type (text or date).
	Input InputType

	// Services lists which services support this field.
	Services []Service
}

// patentFields contains fields available for patent/utility model searches.
// Source: DPMAregister Schnittstellenbeschreibung, sections 2.6.1-2.6.3
var patentFields = map[string]Field{
	// Shared fields (Stammdaten gemeinsam)
	"AKZ":   {Code: "AKZ", Name: "Aktenzeichen/Veröffentlichungsnummer", Description: "file number / publication number", Input: InputText, Services: []Service{ServicePatent}},
	"PN":    {Code: "PN", Name: "Veröffentlichungsnummer", Description: "publication number (DEPN + EPWOPN combined)", Input: InputText, Services: []Service{ServicePatent}},
	"PNN":   {Code: "PNN", Name: "Veröffentlichte Dokumente Dokumentennummer", Description: "publication number without country/kind code", Input: InputText, Services: []Service{ServicePatent}},
	"INH":   {Code: "INH", Name: "Anmelder/Inhaber", Description: "applicant / proprietor", Input: InputText, Services: []Service{ServicePatent}},
	"IN":    {Code: "IN", Name: "Erfinder", Description: "inventor", Input: InputText, Services: []Service{ServicePatent}},
	"AT":    {Code: "AT", Name: "Anmeldetag", Description: "filing date", Input: InputDate, Services: []Service{ServicePatent}},
	"DAT":   {Code: "DAT", Name: "Anmeldetag DE", Description: "effective filing date for Germany", Input: InputDate, Services: []Service{ServicePatent}},
	"TI":    {Code: "TI", Name: "Bezeichnung/Titel", Description: "title / designation", Input: InputText, Services: []Service{ServicePatent}},
	"FT":    {Code: "FT", Name: "Fälligkeitstag", Description: "fee due date", Input: InputDate, Services: []Service{ServicePatent}},
	"FG":    {Code: "FG", Name: "Fälligkeit Grund", Description: "fee due reason", Input: InputText, Services: []Service{ServicePatent}},
	"FOEZ":  {Code: "FOEZ", Name: "Förderkennzeichen", Description: "public funding identifier", Input: InputText, Services: []Service{ServicePatent}},
	"IC":    {Code: "IC", Name: "IPC-Klasse", Description: "IPC classification (main + secondary)", Input: InputText, Services: []Service{ServicePatent}},
	"ICM":   {Code: "ICM", Name: "IPC-Hauptklasse", Description: "IPC main classification", Input: InputText, Services: []Service{ServicePatent}},
	"ICMV":  {Code: "ICMV", Name: "IPC-Version Hauptklasse", Description: "IPC main classification version", Input: InputText, Services: []Service{ServicePatent}},
	"ICS":   {Code: "ICS", Name: "IPC-Nebenklasse", Description: "IPC secondary classification", Input: InputText, Services: []Service{ServicePatent}},
	"ICSV":  {Code: "ICSV", Name: "IPC-Version Nebenklasse", Description: "IPC secondary classification version", Input: InputText, Services: []Service{ServicePatent}},
	"LIZ":   {Code: "LIZ", Name: "Lizenz", Description: "license declaration", Input: InputText, Services: []Service{ServicePatent}},
	"PUB":   {Code: "PUB", Name: "Publikationstag", Description: "publication date", Input: InputDate, Services: []Service{ServicePatent}},
	"SART":  {Code: "SART", Name: "Schutzrechtsart", Description: "protection right type (Patent, Gebrauchsmuster, etc.)", Input: InputText, Services: []Service{ServicePatent}},
	"ST":    {Code: "ST", Name: "Status", Description: "current status of the protection right", Input: InputText, Services: []Service{ServicePatent}},
	"EREGT": {Code: "EREGT", Name: "Tag der ersten Übernahme in DPMAregister", Description: "date first entered in DPMAregister", Input: InputDate, Services: []Service{ServicePatent}},
	"REGT":  {Code: "REGT", Name: "Tag der letzten Aktualisierung", Description: "date of last update in DPMAregister", Input: InputDate, Services: []Service{ServicePatent}},
	"VSTT":  {Code: "VSTT", Name: "Verfahrensstandtag", Description: "procedure status date", Input: InputDate, Services: []Service{ServicePatent}},
	"VST":   {Code: "VST", Name: "Verfahrensstand", Description: "procedure status (use in curly braces)", Input: InputText, Services: []Service{ServicePatent}},
	"PET":   {Code: "PET", Name: "Veröffentlichungstag der Erteilung/Eintragung", Description: "publication date of grant/registration", Input: InputDate, Services: []Service{ServicePatent}},
	"VTR":   {Code: "VTR", Name: "Vertreter", Description: "representative (patent attorney)", Input: InputText, Services: []Service{ServicePatent}},
	"AB":    {Code: "AB", Name: "Abstract", Description: "abstract text", Input: InputText, Services: []Service{ServicePatent}},
	"EVT":   {Code: "EVT", Name: "Erstveröffentlichungstag", Description: "date of first publication", Input: InputDate, Services: []Service{ServicePatent}},

	// Patent-specific fields (Stammdaten Patente)
	"EAT":    {Code: "EAT", Name: "Anmeldetag EP", Description: "effective filing date for EP patent", Input: InputDate, Services: []Service{ServicePatent}},
	"WAT":    {Code: "WAT", Name: "Anmeldetag WO", Description: "effective filing date for PCT application", Input: InputDate, Services: []Service{ServicePatent}},
	"PRC":    {Code: "PRC", Name: "Ausländische Priorität Land", Description: "foreign priority country", Input: InputText, Services: []Service{ServicePatent}},
	"PRNA":   {Code: "PRNA", Name: "Ausländische Priorität Aktenzeichen", Description: "foreign priority application number", Input: InputText, Services: []Service{ServicePatent}},
	"PRDA":   {Code: "PRDA", Name: "Ausländische Priorität Datum", Description: "foreign priority date", Input: InputDate, Services: []Service{ServicePatent}},
	"APR":    {Code: "APR", Name: "Ausstellungspriorität", Description: "exhibition priority name", Input: InputText, Services: []Service{ServicePatent}},
	"APRD":   {Code: "APRD", Name: "Ausstellungspriorität Datum", Description: "exhibition priority date", Input: InputDate, Services: []Service{ServicePatent}},
	"EDS":    {Code: "EDS", Name: "Benannte Vertragsstaaten EP", Description: "designated contracting states EP", Input: InputText, Services: []Service{ServicePatent}},
	"DS":     {Code: "DS", Name: "Benennungs-/Bestimmungsstaaten", Description: "designated states (EDS + WDS combined)", Input: InputText, Services: []Service{ServicePatent}},
	"WDS":    {Code: "WDS", Name: "Bestimmungsstaaten WO", Description: "designated states WO/PCT", Input: InputText, Services: []Service{ServicePatent}},
	"CT":     {Code: "CT", Name: "Entgegenhaltungen", Description: "citations / prior art references", Input: InputText, Services: []Service{ServicePatent}},
	"CTNP":   {Code: "CTNP", Name: "Entgegenhaltungen NPL", Description: "non-patent literature citations", Input: InputText, Services: []Service{ServicePatent}},
	"DPRN":   {Code: "DPRN", Name: "Designpriorität Aktenzeichen", Description: "design priority application number", Input: InputText, Services: []Service{ServicePatent}},
	"DPRC":   {Code: "DPRC", Name: "Designpriorität Land", Description: "design priority country", Input: InputText, Services: []Service{ServicePatent}},
	"DPRD":   {Code: "DPRD", Name: "Designpriorität Datum", Description: "design priority date", Input: InputDate, Services: []Service{ServicePatent}},
	"PRN":    {Code: "PRN", Name: "Inländische Priorität Aktenzeichen", Description: "domestic priority application number", Input: InputText, Services: []Service{ServicePatent}},
	"PRD":    {Code: "PRD", Name: "Inländische Priorität Datum", Description: "domestic priority date", Input: InputDate, Services: []Service{ServicePatent}},
	"TAAKZ":  {Code: "TAAKZ", Name: "Teilung/Ausscheidung aus Aktenzeichen", Description: "divisional application file number", Input: InputText, Services: []Service{ServicePatent}},
	"TIAKZ":  {Code: "TIAKZ", Name: "Teilung/Ausscheidung in Aktenzeichen", Description: "split-off application file number", Input: InputText, Services: []Service{ServicePatent}},
	"DEPN":   {Code: "DEPN", Name: "Veröffentlichte DE-Dokumente", Description: "published DE document numbers", Input: InputText, Services: []Service{ServicePatent}},
	"EPWOPN": {Code: "EPWOPN", Name: "Veröffentlichte EP-/WO-Dokumente", Description: "published EP/WO document numbers", Input: InputText, Services: []Service{ServicePatent}},
	"PC":     {Code: "PC", Name: "Veröffentlichte Dokumente Ländercode", Description: "publication country code", Input: InputText, Services: []Service{ServicePatent}},
	"PCOD":   {Code: "PCOD", Name: "Veröffentlichte Dokumente Schriftartencode", Description: "publication kind code", Input: InputText, Services: []Service{ServicePatent}},
	"EPN":    {Code: "EPN", Name: "Veröffentlichungsnummer EP", Description: "EP publication number", Input: InputText, Services: []Service{ServicePatent}},
	"WPN":    {Code: "WPN", Name: "Veröffentlichungsnummer WO", Description: "WO/PCT publication number", Input: InputText, Services: []Service{ServicePatent}},
	"LANG":   {Code: "LANG", Name: "Veröffentlichungssprache", Description: "publication language (ELANG + WLANG combined)", Input: InputText, Services: []Service{ServicePatent}},
	"ELANG":  {Code: "ELANG", Name: "Veröffentlichungssprache EP", Description: "EP publication language", Input: InputText, Services: []Service{ServicePatent}},
	"WLANG":  {Code: "WLANG", Name: "Veröffentlichungssprache WO", Description: "WO publication language", Input: InputText, Services: []Service{ServicePatent}},

	// Utility model specific (Gebrauchsmuster)
	"AAAKZ":  {Code: "AAAKZ", Name: "Abzweigung aus Aktenzeichen", Description: "branched-off from file number", Input: InputText, Services: []Service{ServicePatent}},
	"EAKZ":   {Code: "EAKZ", Name: "Aktenzeichen EP", Description: "EP file number linked to utility model", Input: InputText, Services: []Service{ServicePatent}},
	"ET":     {Code: "ET", Name: "Eintragungstag", Description: "registration date (utility model/topography)", Input: InputDate, Services: []Service{ServicePatent}},
	"GBMNAT": {Code: "GBMNAT", Name: "PCT Eintritt nationale Phase Gebrauchsmuster", Description: "PCT national phase entry date for utility model", Input: InputDate, Services: []Service{ServicePatent}},
	"NVWT":   {Code: "NVWT", Name: "Tag der ersten nicht vertraulichen Verwertung", Description: "date of first non-confidential use (topography)", Input: InputText, Services: []Service{ServicePatent}},

	// Schutzzertifikate (supplementary protection certificates)
	"TIEZ":  {Code: "TIEZ", Name: "Bezeichnung des Erzeugnisses", Description: "product designation (SPC)", Input: InputText, Services: []Service{ServicePatent}},
	"TIGZ":  {Code: "TIGZ", Name: "Bezeichnung des Grundpatents", Description: "basic patent title (SPC)", Input: InputText, Services: []Service{ServicePatent}},
	"GEC":   {Code: "GEC", Name: "Genehmigung EU Land", Description: "EU marketing authorization country", Input: InputText, Services: []Service{ServicePatent}},
	"GEN":   {Code: "GEN", Name: "Genehmigung EU Nummer", Description: "EU marketing authorization number", Input: InputText, Services: []Service{ServicePatent}},
	"GES":   {Code: "GES", Name: "Genehmigung EU Stelle", Description: "EU marketing authorization authority", Input: InputText, Services: []Service{ServicePatent}},
	"GET":   {Code: "GET", Name: "Genehmigung EU Tag", Description: "EU marketing authorization date", Input: InputDate, Services: []Service{ServicePatent}},
	"GIN":   {Code: "GIN", Name: "Genehmigung Inland Nummer", Description: "domestic marketing authorization number", Input: InputText, Services: []Service{ServicePatent}},
	"GIS":   {Code: "GIS", Name: "Genehmigung Inland Stelle", Description: "domestic marketing authorization authority", Input: InputText, Services: []Service{ServicePatent}},
	"GIT":   {Code: "GIT", Name: "Genehmigung Inland Tag", Description: "domestic marketing authorization date", Input: InputDate, Services: []Service{ServicePatent}},
	"GAKZ":  {Code: "GAKZ", Name: "Grundpatent Aktenzeichen", Description: "basic patent file number (SPC)", Input: InputText, Services: []Service{ServicePatent}},
	"GEAKZ": {Code: "GEAKZ", Name: "Grundpatent Aktenzeichen EP", Description: "basic patent EP file number (SPC)", Input: InputText, Services: []Service{ServicePatent}},
	"LB":    {Code: "LB", Name: "Laufzeit Beginn", Description: "SPC term start date", Input: InputDate, Services: []Service{ServicePatent}},
	"LE":    {Code: "LE", Name: "Laufzeit Ende", Description: "SPC term end date", Input: InputDate, Services: []Service{ServicePatent}},
	"OT":    {Code: "OT", Name: "Veröffentlichungstag der Zertifikatsanmeldung", Description: "SPC application publication date", Input: InputDate, Services: []Service{ServicePatent}},
	"VZ":    {Code: "VZ", Name: "Verwendung des Zertifikats", Description: "SPC use (Arzneimittel/Pflanzenschutzmittel)", Input: InputText, Services: []Service{ServicePatent}},

	// Procedure data fields (used inside curly braces)
	"PART": {Code: "PART", Name: "Verfahrensart (Patentblatt)", Description: "procedure type for patent gazette", Input: InputText, Services: []Service{ServicePatent}},
	"VT":   {Code: "VT", Name: "Verfahrensstand-Monat", Description: "procedure status month (format: Mnn-yyyy)", Input: InputText, Services: []Service{ServicePatent}},
	"HN":   {Code: "HN", Name: "Heftnummer", Description: "gazette issue number", Input: InputText, Services: []Service{ServicePatent}},
	"PJ":   {Code: "PJ", Name: "Publikationsjahr", Description: "publication year", Input: InputText, Services: []Service{ServicePatent}},
}

// designFields contains fields available for design searches (Geschmacksmuster).
// Source: DPMAregister Schnittstellenbeschreibung, section 4.4
var designFields = map[string]Field{
	"AKZ":   {Code: "AKZ", Name: "Aktenzeichen/Designnummer", Description: "file number / design number", Input: InputText, Services: []Service{ServiceDesign}},
	"INH":   {Code: "INH", Name: "Inhaber", Description: "proprietor", Input: InputText, Services: []Service{ServiceDesign}},
	"ENTW":  {Code: "ENTW", Name: "Entwerfer", Description: "designer", Input: InputText, Services: []Service{ServiceDesign}},
	"ERZ":   {Code: "ERZ", Name: "Erzeugnis(se)", Description: "product(s)", Input: InputText, Services: []Service{ServiceDesign}},
	"TI":    {Code: "TI", Name: "Bezeichnung/Titel", Description: "title / designation", Input: InputText, Services: []Service{ServiceDesign}},
	"WKL":   {Code: "WKL", Name: "Warenklasse", Description: "Locarno class", Input: InputText, Services: []Service{ServiceDesign}},
	"RN":    {Code: "RN", Name: "Registernummer", Description: "registration number", Input: InputText, Services: []Service{ServiceDesign}},
	"DNR":   {Code: "DNR", Name: "Designnummer", Description: "design number", Input: InputText, Services: []Service{ServiceDesign}},
	"AT":    {Code: "AT", Name: "Anmeldetag", Description: "filing date", Input: InputDate, Services: []Service{ServiceDesign}},
	"ET":    {Code: "ET", Name: "Eintragungstag", Description: "registration date", Input: InputDate, Services: []Service{ServiceDesign}},
	"ST":    {Code: "ST", Name: "Status", Description: "current status", Input: InputText, Services: []Service{ServiceDesign}},
	"VTR":   {Code: "VTR", Name: "Vertreter", Description: "representative", Input: InputText, Services: []Service{ServiceDesign}},
	"LIZ":   {Code: "LIZ", Name: "Lizenz", Description: "license declaration", Input: InputText, Services: []Service{ServiceDesign}},
	"BA":    {Code: "BA", Name: "Bestandsart", Description: "stock type (aktiv/nicht-aktiv)", Input: InputText, Services: []Service{ServiceDesign}},
	"DB":    {Code: "DB", Name: "Datenbestand", Description: "database (DE/EM)", Input: InputText, Services: []Service{ServiceDesign}},
	"PUB":   {Code: "PUB", Name: "Publikationstag", Description: "publication date", Input: InputDate, Services: []Service{ServiceDesign}},
	"EREGT": {Code: "EREGT", Name: "Tag der ersten Übernahme in DPMAregister", Description: "date first entered in DPMAregister", Input: InputDate, Services: []Service{ServiceDesign}},
	"REGT":  {Code: "REGT", Name: "Tag der letzten Aktualisierung", Description: "date of last update in DPMAregister", Input: InputDate, Services: []Service{ServiceDesign}},

	// Procedure data fields (used inside curly braces)
	"VART": {Code: "VART", Name: "Verfahrensart", Description: "procedure type", Input: InputText, Services: []Service{ServiceDesign}},
	"VSTT": {Code: "VSTT", Name: "Verfahrensstandtag", Description: "procedure status date", Input: InputDate, Services: []Service{ServiceDesign}},
	"VST":  {Code: "VST", Name: "Verfahrensstand", Description: "procedure status", Input: InputText, Services: []Service{ServiceDesign}},
}

// trademarkFields contains fields available for trademark searches (Marken).
// Source: DPMAregister Schnittstellenbeschreibung, section 3.4
var trademarkFields = map[string]Field{
	"MARKE": {Code: "MARKE", Name: "Marke", Description: "trademark text", Input: InputText, Services: []Service{ServiceTrademark}},
	"md":    {Code: "md", Name: "Marke", Description: "trademark text (alias for MARKE)", Input: InputText, Services: []Service{ServiceTrademark}},
	"INH":   {Code: "INH", Name: "Anmelder/Inhaber", Description: "applicant / proprietor", Input: InputText, Services: []Service{ServiceTrademark}},
	"AT":    {Code: "AT", Name: "Anmeldetag", Description: "filing date", Input: InputDate, Services: []Service{ServiceTrademark}},
	"RN":    {Code: "RN", Name: "Registernummer/Aktenzeichen", Description: "registration number / file number", Input: InputText, Services: []Service{ServiceTrademark}},
	"BA":    {Code: "BA", Name: "Bestandsart", Description: "status type (eingetragen, eintragung-nicht-moeglich, etc.)", Input: InputText, Services: []Service{ServiceTrademark}},
	"DB":    {Code: "DB", Name: "Datenbestand", Description: "database (DE/EM/IR)", Input: InputText, Services: []Service{ServiceTrademark}},
	"KL":    {Code: "KL", Name: "Klasse(n)", Description: "Nice class(es)", Input: InputText, Services: []Service{ServiceTrademark}},
	"BKL":   {Code: "BKL", Name: "Bildklasse(n)", Description: "Vienna (image) class(es)", Input: InputText, Services: []Service{ServiceTrademark}},
	"MF":    {Code: "MF", Name: "Markenform", Description: "trademark form (Wort-/Bildmarke, etc.)", Input: InputText, Services: []Service{ServiceTrademark}},
	"LIZ":   {Code: "LIZ", Name: "Lizenzerklärung", Description: "license declaration", Input: InputText, Services: []Service{ServiceTrademark}},
	"VTR":   {Code: "VTR", Name: "Vertreter", Description: "representative", Input: InputText, Services: []Service{ServiceTrademark}},
	"ST":    {Code: "ST", Name: "Status", Description: "current status", Input: InputText, Services: []Service{ServiceTrademark}},

	// Procedure data fields (used inside curly braces)
	"VART": {Code: "VART", Name: "Verfahrensart", Description: "procedure type", Input: InputText, Services: []Service{ServiceTrademark}},
	"VSTT": {Code: "VSTT", Name: "Verfahrensstandtag", Description: "procedure status date", Input: InputDate, Services: []Service{ServiceTrademark}},
	"VST":  {Code: "VST", Name: "Verfahrensstand", Description: "procedure status", Input: InputText, Services: []Service{ServiceTrademark}},
}

// validOperators maps recognized Boolean operators.
// DPMAregister supports both English and German operators.
var validOperators = map[string]bool{
	"AND":   true,
	"OR":    true,
	"NOT":   true,
	"UND":   true,
	"ODER":  true,
	"NICHT": true,
	// Lowercase variants
	"and":   true,
	"or":    true,
	"not":   true,
	"und":   true,
	"oder":  true,
	"nicht": true,
}

// fieldsByService returns the field map for a given service.
func fieldsByService(svc Service) map[string]Field {
	switch svc {
	case ServicePatent:
		return patentFields
	case ServiceDesign:
		return designFields
	case ServiceTrademark:
		return trademarkFields
	default:
		return nil
	}
}

// IsValidField checks if a field code is valid for the given service.
// If service is ServiceAny, the field is checked against all services.
func IsValidField(field string, svc Service) bool {
	if svc == ServiceAny {
		if _, ok := patentFields[field]; ok {
			return true
		}
		if _, ok := designFields[field]; ok {
			return true
		}
		if _, ok := trademarkFields[field]; ok {
			return true
		}
		return field == "exists"
	}
	fields := fieldsByService(svc)
	if fields == nil {
		return false
	}
	_, ok := fields[field]
	return ok
}

// IsValidOperator checks if a string is a recognized Boolean operator.
func IsValidOperator(op string) bool {
	return validOperators[op]
}

// GetField returns the field definition for a code and service.
// Returns the Field and true if found, or zero value and false if not.
func GetField(code string, svc Service) (Field, bool) {
	if svc == ServiceAny {
		if f, ok := patentFields[code]; ok {
			return f, true
		}
		if f, ok := designFields[code]; ok {
			return f, true
		}
		if f, ok := trademarkFields[code]; ok {
			return f, true
		}
		return Field{}, false
	}
	fields := fieldsByService(svc)
	if fields == nil {
		return Field{}, false
	}
	f, ok := fields[code]
	return f, ok
}

// GetValidFields returns all valid field codes for a service.
// If service is ServiceAny, returns fields from all services (deduplicated).
func GetValidFields(svc Service) []string {
	if svc == ServiceAny {
		seen := make(map[string]bool)
		var result []string
		for _, fields := range []map[string]Field{patentFields, designFields, trademarkFields} {
			for code := range fields {
				if !seen[code] {
					seen[code] = true
					result = append(result, code)
				}
			}
		}
		slices.Sort(result)
		return result
	}
	fields := fieldsByService(svc)
	if fields == nil {
		return nil
	}
	result := make([]string, 0, len(fields))
	for code := range fields {
		result = append(result, code)
	}
	slices.Sort(result)
	return result
}
