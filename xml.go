package dpmaconnect

import (
	"encoding/xml"
	"fmt"
)

// XMLParseError indicates a failure to parse XML response data.
type XMLParseError struct {
	Operation string // e.g. "ParsePatentSearch"
	Err       error  // underlying xml.Unmarshal error
}

func (e *XMLParseError) Error() string {
	return fmt.Sprintf("%s: failed to parse XML: %v", e.Operation, e.Err)
}

func (e *XMLParseError) Unwrap() error {
	return e.Err
}

// --- Public types ---

// Party represents a person or organization (applicant, inventor, etc.)
type Party struct {
	Name    string
	Address string
	Country string
}

// PatentSearchResult holds parsed patent search results.
type PatentSearchResult struct {
	TotalHits int
	Hits      []PatentHit
	RawXML    []byte // original XML response bytes
}

// PatentHit represents a single patent search result entry.
type PatentHit struct {
	LeadingRegisteredNumber string
	RegisteredNumber        string
	Type                    string // "Patent", "Utility model"
	LegalStatus             string
	Title                   string
	Applicants              []string
	IPCClasses              []string // main + further classifications
	ApplicationDate         string
	PublicationDate         string
	RegistrationDate        string
}

// PatentInfo holds parsed patent register info (ST36 bibliographic data).
type PatentInfo struct {
	Publications   []PatentPublication
	ApplicationRef PatentDocumentRef
	Title          string
	IPCClasses     []string
	Applicants     []Party
	Inventors      []Party
	IPRightType    string // patent, gebrauchsmuster, schutzzertifikat, etc.
	Status         string // anhaengig-in-kraft, nicht-anhaengig-erloschen
	Abstract       string
	FilingDate     string
	FirstPubDate   string
}

// PatentPublication represents a publication reference within patent info.
type PatentPublication struct {
	Country string
	Number  string
	Kind    string
}

// PatentDocumentRef represents an application or document reference.
type PatentDocumentRef struct {
	Country string
	Number  string
	Date    string
}

// TrademarkSearchResult holds parsed trademark search results.
type TrademarkSearchResult struct {
	TotalHits int
	Hits      []TrademarkHit
	RawXML    []byte // original XML response bytes
}

// TrademarkHit represents a single trademark search result entry.
type TrademarkHit struct {
	ApplicationNumber string
	MarkText          string
	MarkFeature       string // wortmarke, wort-bildmarke, etc.
	Classification    string
	Status            string
	ApplicationDate   string
	RegistrationDate  string
	Applicant         string
}

// TrademarkInfo holds parsed trademark register info (ST66).
type TrademarkInfo struct {
	ApplicationNumber  string
	RegistrationNumber string
	ApplicationDate    string
	RegistrationDate   string
	ExpiryDate         string
	Status             string
	MarkFeature        string
	MarkText           string
	Applicants         []Party
	Classifications    []TrademarkClass
}

// TrademarkClass represents a Nice classification entry.
type TrademarkClass struct {
	Number      string
	Description string
}

// DesignSearchResult holds parsed design search results.
type DesignSearchResult struct {
	TotalHits int
	Hits      []DesignHit
	RawXML    []byte // original XML response bytes
}

// DesignHit represents a single design search result entry.
type DesignHit struct {
	DesignIdentifier   string
	ApplicationNumber  string
	RegistrationNumber string
	Title              string
	Applicant          string
	ClassNumber        string
	Status             string
	ApplicationDate    string
	RegistrationDate   string
	PublicationDate    string
}

// DesignInfo holds parsed design register info (ST86).
type DesignInfo struct {
	DesignIdentifier   string
	ApplicationNumber  string
	RegistrationNumber string
	ApplicationDate    string
	RegistrationDate   string
	Title              string
	Status             string
	Applicants         []Party
	ClassNumber        string
	ClassDescription   string
}

// --- Private XML structs ---

// Patent search XML
type xmlPatentHitList struct {
	XMLName  xml.Name                 `xml:"PatentHitList"`
	HitCount int                      `xml:"HitCount,attr"`
	Records  []xmlPatentHitListRecord `xml:"PatentHitListRecord"`
}

type xmlPatentHitListRecord struct {
	LeadingRegisteredNumber string                    `xml:"leading-registered-number"`
	RegisteredNumber        string                    `xml:"registered-number"`
	Type                    string                    `xml:"type"`
	LegalStatus             string                    `xml:"legalstatus"`
	Title                   string                    `xml:"invention-title"`
	MainClassification      xmlPatentClassification   `xml:"main-classification"`
	FurtherClassifications  []xmlPatentClassification `xml:"further-classification"`
	ApplicationDate         string                    `xml:"applicationDate"`
	PublicationDate         string                    `xml:"publicationDate"`
	RegistrationDate        string                    `xml:"registrationDate"`
	Applicants              xmlPatentApplicants       `xml:"applicants"`
}

type xmlPatentClassification struct {
	Classification string `xml:"classification"`
}

type xmlPatentApplicants struct {
	Applicant []string `xml:"applicant"`
}

// Patent info XML (ST36 with namespace)
type xmlDPMAPatentDocument struct {
	XMLName           xml.Name         `xml:"dpma-patent-document"`
	BibliographicData xmlPatentBibData `xml:"bibliographic-data"`
}

type xmlPatentBibData struct {
	PublicationRefs     xmlPublicationRefs     `xml:"publication-references"`
	ApplicationRef      xmlApplicationRef      `xml:"application-reference"`
	InventionTitle      xmlInventionTitle      `xml:"invention-title"`
	ClassificationsIPCR xmlClassificationsIPCR `xml:"classifications-ipcr"`
	Parties             xmlParties             `xml:"parties"`
	OfficeSpecific      xmlOfficeSpecificBib   `xml:"office-specific-bib-data"`
}

type xmlPublicationRefs struct {
	Refs []xmlPublicationRef `xml:"publication-reference"`
}

type xmlPublicationRef struct {
	DocumentID xmlDocumentID `xml:"document-id"`
}

type xmlApplicationRef struct {
	DocumentID xmlDocumentID `xml:"document-id"`
}

type xmlDocumentID struct {
	Country   string `xml:"country"`
	DocNumber string `xml:"doc-number"`
	Kind      string `xml:"kind"`
	Date      string `xml:"date"`
}

type xmlInventionTitle struct {
	Lang  string `xml:"lang,attr"`
	Value string `xml:",chardata"`
}

type xmlClassificationsIPCR struct {
	Classifications []xmlClassificationIPCR `xml:"classification-ipcr"`
}

type xmlClassificationIPCR struct {
	Text string `xml:"text"`
}

type xmlParties struct {
	Applicants xmlStructuredApplicants `xml:"applicants"`
	Inventors  xmlStructuredInventors  `xml:"inventors"`
}

type xmlStructuredApplicants struct {
	Applicants []xmlStructuredParty `xml:"applicant"`
}

type xmlStructuredInventors struct {
	Inventors []xmlStructuredParty `xml:"inventor"`
}

type xmlStructuredParty struct {
	Addressbook xmlAddressbook `xml:"addressbook"`
}

type xmlAddressbook struct {
	Text    string     `xml:"text"`
	Name    string     `xml:"name"`
	Address xmlAddress `xml:"address"`
}

type xmlAddress struct {
	Address1 string `xml:"address-1"`
	Country  string `xml:"country"`
}

type xmlOfficeSpecificBib struct {
	TypeOfIPRight    string `xml:"type-of-ip-right"`
	Status           string `xml:"status"`
	NationalFiling   string `xml:"national-filing-data"`
	FirstPublication string `xml:"date-of-first-publication"`
	Abstract         string `xml:"abstract"`
}

// Trademark search XML
type xmlTrademarkHitList struct {
	XMLName  xml.Name            `xml:"HitList"`
	HitCount int                 `xml:"HitCount,attr"`
	Entries  []xmlTrademarkEntry `xml:"Entries"`
}

type xmlTrademarkEntry struct {
	ApplicationNumber string `xml:"ApplicationNumber"`
	MarkVerbalText    string `xml:"MarkVerbalElementText"`
	MarkFeature       string `xml:"MarkFeature"`
	Classification    string `xml:"Classification"`
	Status            string `xml:"MarkCurrentStatusCode"`
	ApplicationDate   string `xml:"ApplicationDate"`
	RegistrationDate  string `xml:"RegistrationDate"`
	Applicant         string `xml:"Applicant"`
}

// Trademark info XML (ST66 with namespace)
type xmlTrademarkTransaction struct {
	XMLName xml.Name         `xml:"Transaction"`
	Body    xmlTrademarkBody `xml:"TradeMarkTransactionBody"`
}

type xmlTrademarkBody struct {
	Content xmlTrademarkContent `xml:"TransactionContentDetails"`
}

type xmlTrademarkContent struct {
	Data xmlTrademarkData `xml:"TransactionData"`
}

type xmlTrademarkData struct {
	Details xmlTrademarkDetails `xml:"TradeMarkDetails"`
}

type xmlTrademarkDetails struct {
	TradeMark xmlTradeMark `xml:"TradeMark"`
}

type xmlTradeMark struct {
	ApplicationNumber  string                   `xml:"ApplicationNumber"`
	ApplicationDate    string                   `xml:"ApplicationDate"`
	RegistrationNumber string                   `xml:"RegistrationNumber"`
	RegistrationDate   string                   `xml:"RegistrationDate"`
	ExpiryDate         string                   `xml:"ExpiryDate"`
	Status             string                   `xml:"MarkCurrentStatusCode"`
	MarkFeature        string                   `xml:"MarkFeature"`
	WordMark           xmlWordMarkSpecification `xml:"WordMarkSpecification"`
	GoodsServices      xmlGoodsServicesDetails  `xml:"GoodsServicesDetails"`
	ApplicantDetails   xmlTMApplicantDetails    `xml:"ApplicantDetails"`
}

type xmlWordMarkSpecification struct {
	MarkVerbalElementText string `xml:"MarkVerbalElementText"`
}

type xmlGoodsServicesDetails struct {
	GoodsServices xmlGoodsServices `xml:"GoodsServices"`
}

type xmlGoodsServices struct {
	ClassDescriptions xmlClassDescriptionDetails `xml:"ClassDescriptionDetails"`
}

type xmlClassDescriptionDetails struct {
	ClassDescriptions []xmlClassDescription `xml:"ClassDescription"`
}

type xmlClassDescription struct {
	ClassNumber string `xml:"ClassNumber"`
	Description string `xml:"GoodsServicesDescription"`
}

type xmlTMApplicantDetails struct {
	Applicants []xmlTMApplicant `xml:"Applicant"`
}

type xmlTMApplicant struct {
	AddressBook xmlTMApplicantAddressBook `xml:"ApplicantAddressBook"`
}

type xmlTMApplicantAddressBook struct {
	FormattedNameAddress xmlFormattedNameAddress `xml:"FormattedNameAddress"`
}

type xmlFormattedNameAddress struct {
	Name    xmlFreeFormatNameContainer   `xml:"Name"`
	Address xmlFormattedAddressContainer `xml:"Address"`
}

type xmlFreeFormatNameContainer struct {
	FreeFormatName xmlFreeFormatName `xml:"FreeFormatName"`
}

type xmlFreeFormatName struct {
	Details xmlFreeFormatNameDetails `xml:"FreeFormatNameDetails"`
}

type xmlFreeFormatNameDetails struct {
	NameLine string `xml:"FreeFormatNameLine"`
}

type xmlFormattedAddressContainer struct {
	FormattedAddress xmlFormattedAddress `xml:"FormattedAddress"`
}

type xmlFormattedAddress struct {
	City     string `xml:"AddressCity"`
	County   string `xml:"AddressCounty"` // often used for country code
	State    string `xml:"AddressState"`
	Postcode string `xml:"AddressPostcode"`
}

// Design search XML
type xmlDesignHitList struct {
	XMLName  xml.Name                 `xml:"DesignHitList"`
	HitCount int                      `xml:"HitCount,attr"`
	Records  []xmlDesignHitListRecord `xml:"DesignHitListRecord"`
}

type xmlDesignHitListRecord struct {
	DesignIdentifier   string `xml:"DesignIdentifier"`
	ApplicationNumber  string `xml:"DesignApplicationNumber"`
	RegistrationNumber string `xml:"RegistrationNumber"`
	Title              string `xml:"ProductDescription"`
	Applicant          string `xml:"OwnerOrApplicant"`
	ClassNumber        string `xml:"ClassNumber"`
	Status             string `xml:"DesignStatus"`
	ApplicationDate    string `xml:"DesignApplicationDate"`
	RegistrationDate   string `xml:"RegistrationDate"`
	PublicationDate    string `xml:"PublicationDate"`
}

// Design info XML (ST86 with namespace)
type xmlDesignTransaction struct {
	XMLName xml.Name      `xml:"Transaction"`
	Body    xmlDesignBody `xml:"DesignTransactionBody"`
}

type xmlDesignBody struct {
	Content xmlDesignContent `xml:"TransactionContentDetails"`
}

type xmlDesignContent struct {
	Data xmlDesignTransactionData `xml:"TransactionData"`
}

type xmlDesignTransactionData struct {
	ApplicationDetails xmlDesignApplicationDetails `xml:"DesignApplicationDetails"`
}

type xmlDesignApplicationDetails struct {
	Application xmlDesignApplication `xml:"DesignApplication"`
}

type xmlDesignApplication struct {
	ApplicationNumber string              `xml:"DesignApplicationNumber"`
	ApplicationDate   string              `xml:"DesignApplicationDate"`
	Designs           xmlDesignDetails    `xml:"DesignDetails"`
	Applicants        xmlDesignApplicants `xml:"ApplicantDetails"`
}

type xmlDesignDetails struct {
	Designs []xmlDesign `xml:"Design"`
}

type xmlDesign struct {
	DesignIdentifier   string                      `xml:"DesignIdentifier"`
	RegistrationNumber string                      `xml:"RegistrationNumber"`
	RegistrationDate   string                      `xml:"RegistrationDate"`
	Title              string                      `xml:"DesignTitle"`
	Extension          xmlDesignExtension          `xml:"DE_ST86Extension"`
	Products           xmlIndicationProductDetails `xml:"IndicationProductDetails"`
	Applicants         xmlDesignApplicants         `xml:"ApplicantDetails"`
}

type xmlDesignExtension struct {
	Status string `xml:"DesignStatus"`
}

type xmlIndicationProductDetails struct {
	Products []xmlIndicationProduct `xml:"IndicationProduct"`
}

type xmlIndicationProduct struct {
	ClassDescriptions xmlDesignClassDescriptionDetails `xml:"ClassDescriptionDetails"`
}

type xmlDesignClassDescriptionDetails struct {
	Descriptions []xmlDesignClassDescription `xml:"ClassDescription"`
}

type xmlDesignClassDescription struct {
	ClassNumber string `xml:"ClassNumber"`
	Description string `xml:"ProductDescription"`
}

type xmlDesignApplicants struct {
	Applicants []xmlDesignApplicant `xml:"Applicant"`
}

type xmlDesignApplicant struct {
	AddressBook xmlDesignApplicantAddressBook `xml:"ApplicantAddressBook"`
}

type xmlDesignApplicantAddressBook struct {
	FormattedNameAddress xmlDesignFormattedNameAddress `xml:"FormattedNameAddress"`
}

type xmlDesignFormattedNameAddress struct {
	Name    xmlDesignNameContainer    `xml:"Name"`
	Address xmlDesignAddressContainer `xml:"Address"`
}

type xmlDesignNameContainer struct {
	FormattedName xmlDesignFormattedName `xml:"FormattedName"`
}

type xmlDesignFormattedName struct {
	OrgName string `xml:"OrganizationName"`
}

type xmlDesignAddressContainer struct {
	FormattedAddress xmlDesignFormattedAddress `xml:"FormattedAddress"`
}

type xmlDesignFormattedAddress struct {
	City   string `xml:"AddressCity"`
	County string `xml:"AddressCounty"`
	State  string `xml:"AddressState"`
}

// extractDesignApplicants converts raw XML design applicants to Party values.
func extractDesignApplicants(raw []xmlDesignApplicant) []Party {
	parties := make([]Party, 0, len(raw))
	for _, a := range raw {
		fna := a.AddressBook.FormattedNameAddress
		country := fna.Address.FormattedAddress.County
		if country == "" {
			country = fna.Address.FormattedAddress.State
		}
		parties = append(parties, Party{
			Name:    fna.Name.FormattedName.OrgName,
			Address: fna.Address.FormattedAddress.City,
			Country: country,
		})
	}
	return parties
}

// --- Parse functions ---

// ParsePatentSearch parses a patent search XML response.
func ParsePatentSearch(data []byte) (*PatentSearchResult, error) {
	var raw xmlPatentHitList
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, &XMLParseError{Operation: "ParsePatentSearch", Err: err}
	}

	result := &PatentSearchResult{
		TotalHits: raw.HitCount,
		Hits:      make([]PatentHit, len(raw.Records)),
		RawXML:    data,
	}

	for i := range raw.Records {
		r := &raw.Records[i]
		var ipcClasses []string
		if r.MainClassification.Classification != "" {
			ipcClasses = append(ipcClasses, r.MainClassification.Classification)
		}
		for _, fc := range r.FurtherClassifications {
			if fc.Classification != "" {
				ipcClasses = append(ipcClasses, fc.Classification)
			}
		}

		result.Hits[i] = PatentHit{
			LeadingRegisteredNumber: r.LeadingRegisteredNumber,
			RegisteredNumber:        r.RegisteredNumber,
			Type:                    r.Type,
			LegalStatus:             r.LegalStatus,
			Title:                   r.Title,
			Applicants:              r.Applicants.Applicant,
			IPCClasses:              ipcClasses,
			ApplicationDate:         r.ApplicationDate,
			PublicationDate:         r.PublicationDate,
			RegistrationDate:        r.RegistrationDate,
		}
	}

	return result, nil
}

// ParsePatentInfo parses a patent info XML response (ST36 format).
func ParsePatentInfo(data []byte) (*PatentInfo, error) {
	var raw xmlDPMAPatentDocument
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, &XMLParseError{Operation: "ParsePatentInfo", Err: err}
	}

	bib := raw.BibliographicData

	var publications []PatentPublication
	for _, ref := range bib.PublicationRefs.Refs {
		publications = append(publications, PatentPublication{
			Country: ref.DocumentID.Country,
			Number:  ref.DocumentID.DocNumber,
			Kind:    ref.DocumentID.Kind,
		})
	}

	var ipcClasses []string
	for _, c := range bib.ClassificationsIPCR.Classifications {
		if c.Text != "" {
			ipcClasses = append(ipcClasses, c.Text)
		}
	}

	var applicants []Party
	for _, a := range bib.Parties.Applicants.Applicants {
		applicants = append(applicants, Party{
			Name:    a.Addressbook.Name,
			Address: a.Addressbook.Address.Address1,
			Country: a.Addressbook.Address.Country,
		})
	}

	var inventors []Party
	for _, inv := range bib.Parties.Inventors.Inventors {
		inventors = append(inventors, Party{
			Name:    inv.Addressbook.Name,
			Address: inv.Addressbook.Address.Address1,
			Country: inv.Addressbook.Address.Country,
		})
	}

	return &PatentInfo{
		Publications: publications,
		ApplicationRef: PatentDocumentRef{
			Country: bib.ApplicationRef.DocumentID.Country,
			Number:  bib.ApplicationRef.DocumentID.DocNumber,
			Date:    bib.ApplicationRef.DocumentID.Date,
		},
		Title:        bib.InventionTitle.Value,
		IPCClasses:   ipcClasses,
		Applicants:   applicants,
		Inventors:    inventors,
		IPRightType:  bib.OfficeSpecific.TypeOfIPRight,
		Status:       bib.OfficeSpecific.Status,
		Abstract:     bib.OfficeSpecific.Abstract,
		FilingDate:   bib.OfficeSpecific.NationalFiling,
		FirstPubDate: bib.OfficeSpecific.FirstPublication,
	}, nil
}

// ParseTrademarkSearch parses a trademark search XML response.
func ParseTrademarkSearch(data []byte) (*TrademarkSearchResult, error) {
	var raw xmlTrademarkHitList
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, &XMLParseError{Operation: "ParseTrademarkSearch", Err: err}
	}

	result := &TrademarkSearchResult{
		TotalHits: raw.HitCount,
		Hits:      make([]TrademarkHit, len(raw.Entries)),
		RawXML:    data,
	}

	for i := range raw.Entries {
		e := &raw.Entries[i]
		result.Hits[i] = TrademarkHit{
			ApplicationNumber: e.ApplicationNumber,
			MarkText:          e.MarkVerbalText,
			MarkFeature:       e.MarkFeature,
			Classification:    e.Classification,
			Status:            e.Status,
			ApplicationDate:   e.ApplicationDate,
			RegistrationDate:  e.RegistrationDate,
			Applicant:         e.Applicant,
		}
	}

	return result, nil
}

// ParseTrademarkInfo parses a trademark info XML response (ST66 format).
func ParseTrademarkInfo(data []byte) (*TrademarkInfo, error) {
	var raw xmlTrademarkTransaction
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, &XMLParseError{Operation: "ParseTrademarkInfo", Err: err}
	}

	tm := raw.Body.Content.Data.Details.TradeMark

	var applicants []Party
	for _, a := range tm.ApplicantDetails.Applicants {
		fna := a.AddressBook.FormattedNameAddress
		country := fna.Address.FormattedAddress.County
		if country == "" {
			country = fna.Address.FormattedAddress.State
		}
		applicants = append(applicants, Party{
			Name:    fna.Name.FreeFormatName.Details.NameLine,
			Address: fna.Address.FormattedAddress.City,
			Country: country,
		})
	}

	var classes []TrademarkClass
	for _, cd := range tm.GoodsServices.GoodsServices.ClassDescriptions.ClassDescriptions {
		classes = append(classes, TrademarkClass{
			Number:      cd.ClassNumber,
			Description: cd.Description,
		})
	}

	return &TrademarkInfo{
		ApplicationNumber:  tm.ApplicationNumber,
		RegistrationNumber: tm.RegistrationNumber,
		ApplicationDate:    tm.ApplicationDate,
		RegistrationDate:   tm.RegistrationDate,
		ExpiryDate:         tm.ExpiryDate,
		Status:             tm.Status,
		MarkFeature:        tm.MarkFeature,
		MarkText:           tm.WordMark.MarkVerbalElementText,
		Applicants:         applicants,
		Classifications:    classes,
	}, nil
}

// ParseDesignSearch parses a design search XML response.
func ParseDesignSearch(data []byte) (*DesignSearchResult, error) {
	var raw xmlDesignHitList
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, &XMLParseError{Operation: "ParseDesignSearch", Err: err}
	}

	result := &DesignSearchResult{
		TotalHits: raw.HitCount,
		Hits:      make([]DesignHit, len(raw.Records)),
		RawXML:    data,
	}

	for i := range raw.Records {
		result.Hits[i] = DesignHit(raw.Records[i])
	}

	return result, nil
}

// ParseDesignInfo parses a design info XML response (ST86 format).
func ParseDesignInfo(data []byte) (*DesignInfo, error) {
	var raw xmlDesignTransaction
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, &XMLParseError{Operation: "ParseDesignInfo", Err: err}
	}

	app := raw.Body.Content.Data.ApplicationDetails.Application

	info := &DesignInfo{
		ApplicationNumber: app.ApplicationNumber,
		ApplicationDate:   app.ApplicationDate,
	}

	// Extract from first Design element
	if len(app.Designs.Designs) > 0 {
		d := app.Designs.Designs[0]
		info.DesignIdentifier = d.DesignIdentifier
		info.RegistrationNumber = d.RegistrationNumber
		info.RegistrationDate = d.RegistrationDate
		info.Title = d.Title
		info.Status = d.Extension.Status

		// Class info from IndicationProductDetails
		if len(d.Products.Products) > 0 {
			p := d.Products.Products[0]
			if len(p.ClassDescriptions.Descriptions) > 0 {
				cd := p.ClassDescriptions.Descriptions[0]
				info.ClassNumber = cd.ClassNumber
				info.ClassDescription = cd.Description
			}
		}

		// Design-level applicants take priority
		if len(d.Applicants.Applicants) > 0 {
			info.Applicants = extractDesignApplicants(d.Applicants.Applicants)
			return info, nil
		}
	}

	// Fall back to application-level applicants
	info.Applicants = extractDesignApplicants(app.Applicants.Applicants)

	return info, nil
}
