package dpmaconnect

import "strings"

// PatentNumberSystem names the numbering system of a register record's
// leading-registered-number.
type PatentNumberSystem string

// Number systems the patent register reports in search hits.
const (
	// NumberSystemDE is a German file number: digits only, check digit last
	// ("100273602" for 100 27 360.2, "82337144" for G 82 33 714.4).
	NumberSystemDE PatentNumberSystem = "DE"
	// NumberSystemDD is a patent of the former GDR: "DD" and the patent number
	// ("DD136869").
	NumberSystemDD PatentNumberSystem = "DD"
	// NumberSystemEP is a European application: "E" and the EPO application
	// number with its check digit ("E247834666" for 24 78 3466.6).
	NumberSystemEP PatentNumberSystem = "EP"
	// NumberSystemPCT is a PCT application: receiving office, year and serial
	// ("CN2019089592" for PCT/CN2019/089592).
	NumberSystemPCT PatentNumberSystem = "PCT"
)

// RegisteredNumberClass is the result of ClassifyRegisteredNumber.
type RegisteredNumberClass struct {
	// System is the numbering system, empty when the form is not recognised.
	System PatentNumberSystem
	// Country is the code the register itself uses for such a record: "DE",
	// "DD", "EP", or "WO" for a PCT application.
	Country string
	// ReceivingOffice is the PCT receiving office ("CN"), PCT only.
	ReceivingOffice string
}

// ClassifyRegisteredNumber classifies a leading-registered-number AS REPORTED
// BY THE REGISTER in a search hit (PatentHit.LeadingRegisteredNumber). It is a
// pure function of the number's form:
//
//	digits only                -> German file      (DE)
//	"E" + digits               -> European file    (EP)
//	two letters + ten digits   -> PCT application  (WO, receiving office = the letters)
//	"DD" + digits              -> GDR patent       (DD)
//
// Do not feed it user input: a publication number such as "DE10027362" has the
// shape of a PCT registered number, which is why the register's own value is
// required. Anything else yields the zero value.
func ClassifyRegisteredNumber(leading string) RegisteredNumberClass {
	s := strings.TrimSpace(leading)
	i := strings.IndexFunc(s, isASCIIDigit)
	if i < 0 || strings.IndexFunc(s[i:], isNotASCIIDigit) >= 0 {
		return RegisteredNumberClass{}
	}
	prefix, digits := s[:i], s[i:]
	switch {
	case prefix == "":
		return RegisteredNumberClass{System: NumberSystemDE, Country: "DE"}
	case prefix == "E":
		return RegisteredNumberClass{System: NumberSystemEP, Country: "EP"}
	case len(prefix) == 2 && isASCIIUpper(prefix) && len(digits) == 10:
		return RegisteredNumberClass{System: NumberSystemPCT, Country: "WO", ReceivingOffice: prefix}
	case prefix == "DD":
		return RegisteredNumberClass{System: NumberSystemDD, Country: "DD"}
	}
	return RegisteredNumberClass{}
}

// PublicationNumber is a publication number split into its parts. Number holds
// the digits without leading zeros, so the register's zero-padded form
// ("000010027360") and the usual written form ("10027360") compare equal.
type PublicationNumber struct {
	Country string // two-letter office code, empty when the input had none
	Number  string // digits, leading zeros removed
	Kind    string // kind code ("A1", "B4", "U1"), empty when the input had none
}

// ParsePublicationNumber splits a publication number into country, number and
// kind code. It accepts the forms callers write ("DE10027360A1",
// "de 100 27 360 A1", "EP4780303", "10027360") and the forms the register
// writes: a zero-padded doc-number ("000010027360") and the combined
// published-documents / citation form ("DD000000136869A1", "US000005371412A").
// Whitespace is ignored and letters are upper-cased. ok is false when the input
// is not an optional two-letter country, digits, and an optional kind code (one
// letter, optionally followed by one digit).
func ParsePublicationNumber(s string) (pn PublicationNumber, ok bool) {
	s = strings.ToUpper(strings.Join(strings.Fields(s), ""))
	if len(s) >= 2 && isASCIIUpper(s[:2]) {
		pn.Country, s = s[:2], s[2:]
	}
	end := strings.IndexFunc(s, isNotASCIIDigit)
	if end < 0 {
		end = len(s)
	}
	digits, kind := s[:end], s[end:]
	if digits == "" {
		return PublicationNumber{}, false
	}
	switch {
	case kind == "":
	case len(kind) == 1 && isASCIIUpper(kind):
	case len(kind) == 2 && isASCIIUpper(kind[:1]) && isASCIIDigit(rune(kind[1])):
	default:
		return PublicationNumber{}, false
	}
	pn.Number = NormalizePublicationDocNumber(digits)
	pn.Kind = kind
	return pn, true
}

// NormalizePublicationDocNumber removes whitespace and the leading zeros the
// register pads a publication doc-number with: "000010027360" becomes
// "10027360", "112019001234" stays as it is. A value of zeros only becomes "0".
func NormalizePublicationDocNumber(docNumber string) string {
	s := strings.Join(strings.Fields(docNumber), "")
	trimmed := strings.TrimLeft(s, "0")
	if trimmed == "" && s != "" {
		return "0"
	}
	return trimmed
}

// Matches reports whether two publication numbers denote the same publication:
// the numbers are equal, and country and kind code are equal wherever BOTH
// sides state them. A side that omits the country or the kind code does not
// constrain it, so "DE10027360" matches both the A1 and the B4 publication.
func (pn PublicationNumber) Matches(other PublicationNumber) bool {
	if pn.Number == "" || pn.Number != other.Number {
		return false
	}
	if pn.Country != "" && other.Country != "" && pn.Country != other.Country {
		return false
	}
	return pn.Kind == "" || other.Kind == "" || pn.Kind == other.Kind
}

// MatchesNumber reports whether the publication reference is the publication a
// caller wrote as publicationNumber (e.g. "DE10027360A1"), see
// PublicationNumber.Matches. The register's zero padding is ignored.
func (p PatentPublication) MatchesNumber(publicationNumber string) bool {
	return matchesPublication(p.Country, p.Number, p.Kind, publicationNumber)
}

// MatchesPublicationNumber is PatentPublication.MatchesNumber for a document
// reference such as a PatentInfo.RegionalPublications entry (EP and WO
// publications, which the register lists without kind code).
func (r PatentDocumentRef) MatchesPublicationNumber(publicationNumber string) bool {
	return matchesPublication(r.Country, r.Number, r.Kind, publicationNumber)
}

func matchesPublication(country, number, kind, requested string) bool {
	own, ok := ParsePublicationNumber(country + number + kind)
	if !ok {
		return false
	}
	req, ok := ParsePublicationNumber(requested)
	return ok && own.Matches(req)
}

func isASCIIDigit(r rune) bool    { return r >= '0' && r <= '9' }
func isNotASCIIDigit(r rune) bool { return !isASCIIDigit(r) }

func isASCIIUpper(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 'A' || s[i] > 'Z' {
			return false
		}
	}
	return s != ""
}
