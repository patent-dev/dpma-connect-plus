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

// Party represents a person or organization (applicant, inventor, agent, etc.)
type Party struct {
	Text     string // full free-format line as returned by the office, if present
	Name     string
	Address  string
	Street   string
	Postcode string
	Country  string
	Sequence string // @sequence attribute, when present
	AppType  string // applicant @app-type attribute, when present
}

// PatentSearchResult holds parsed patent search results.
type PatentSearchResult struct {
	TotalHits    int
	DocumentHits int // <Counter><DocumentHits>
	DatabaseHits int // <Counter><DatabaseHits>
	Hits         []PatentHit
	RawXML       []byte // original XML response bytes
}

// PatentHit represents a single patent search result entry.
type PatentHit struct {
	LeadingRegisteredNumber string
	RegisteredNumber        string
	Type                    string // "Patent", "Utility model"
	LegalStatus             string
	Title                   string
	Applicants              []string
	Inventors               []string
	Agent                   string
	IPCClasses              []string          // main + further classification symbols
	Classifications         []PatentSearchIPC // main + further with edition
	MainClassification      PatentSearchIPC   // main classification with edition
	FurtherClassifications  []PatentSearchIPC // further classifications with edition
	ApplicationDate         string
	PublicationDate         string
	RegistrationDate        string
}

// PatentSearchIPC is a classification symbol with its edition (patent search).
type PatentSearchIPC struct {
	Classification string
	Edition        string
}

// PatentInfo holds parsed patent register info (ST36 bibliographic data).
type PatentInfo struct {
	SchemaVersion          string // @schema-version
	Publications           []PatentPublication
	ApplicationRef         PatentDocumentRef
	Title                  string
	TitleLang              string
	IPCClasses             []string // classification symbols (back-compat convenience)
	Classifications        []PatentClassification
	Applicants             []Party
	Inventors              []Party
	Agents                 []Party // representatives
	CorrespondenceParty    Party   // parties/correspondence-address
	PriorityClaims         []PatentPriorityClaim
	IPRightType            string // patent, gebrauchsmuster, schutzzertifikat, etc.
	Status                 string // anhaengig-in-kraft, nicht-anhaengig-erloschen
	AgentType              string // office-specific agent-type
	Abstract               string
	FilingDate             string
	FirstPubDate           string
	DateLastRegisterUpdate string
	Events                 []PatentEvent // legal-event / prosecution history
	RawXML                 []byte        // original XML response bytes
}

// PatentClassification represents a classification-ipcr entry with version.
type PatentClassification struct {
	Sequence         string
	Symbol           string // <text>
	VersionIndicator string // <ipc-version-indicator>
}

// PatentPublication represents a publication reference within patent info.
type PatentPublication struct {
	Sequence string
	Country  string
	Number   string
	Kind     string
	Date     string
}

// PatentDocumentRef represents an application or document reference.
type PatentDocumentRef struct {
	Country string
	Number  string
	Date    string
}

// PatentPriorityClaim represents a priority-claim entry.
type PatentPriorityClaim struct {
	Sequence  string
	Country   string
	Date      string
	DocNumber string
}

// PatentEvent represents one legal-event / prosecution-history entry.
type PatentEvent struct {
	TypeOfProcedure        string
	ProceduralStatus       string
	DateOfProceduralStatus string
	PublicationInfo        *PatentEventPublication
	IPCMainClass           *PatentIPCClass
	PreviousIPCMainClass   *PatentIPCClass
	PublishedDocuments     []string
}

// PatentEventPublication represents the publication-info of an event.
type PatentEventPublication struct {
	IssueNumber     string
	Year            string
	PublicationDate string
	PublicationType string
	Part            string
}

// PatentIPCClass represents an ipc-main-class / previous-ipc-main-class.
type PatentIPCClass struct {
	Symbol  string
	Version string
}

// TrademarkSearchResult holds parsed trademark search results.
type TrademarkSearchResult struct {
	TotalHits    int
	DocumentHits int
	DatabaseHits int
	Hits         []TrademarkHit
	RawXML       []byte // original XML response bytes
}

// TrademarkHit represents a single trademark search result entry.
type TrademarkHit struct {
	Number                 string
	RegistrationOfficeCode string
	ApplicationNumber      string
	MarkText               string
	MarkDescriptionText    string // MarkDescriptionDetails/DescriptionText
	MarkFeature            string // wortmarke, wort-bildmarke, etc.
	Classification         string
	Status                 string
	ApplicationDate        string
	RegistrationDate       string
	Applicant              string
	Representative         string
}

// TrademarkInfo holds parsed trademark register info (ST66).
type TrademarkInfo struct {
	RegistrationOfficeCode     string
	ApplicationNumber          string
	RegistrationNumber         string
	ApplicationDate            string
	RegistrationDate           string
	ExpiryDate                 string
	TerminationDate            string
	Status                     string
	KindMark                   string
	MarkFeature                string
	OppositionPeriodStartDate  string
	OppositionPeriodEndDate    string
	MarkText                   string
	MarkTextLang               string // MarkVerbalElementText @languageCode
	StandardCharacterIndicator string // MarkStandardCharacterIndicator
	ProposedLeadingClassNumber string
	Applicants                 []Party
	Representatives            []Party
	CorrespondenceParty        Party
	Classifications            []TrademarkClass
	Records                    []TrademarkRecord // register-event history
	FileNumber                 string            // de:FileNumber
	MarkFeatureDPMA            string            // de:MarkFeatureDPMA
	RawXML                     []byte            // original XML response bytes
}

// TrademarkClass represents a Nice classification entry.
type TrademarkClass struct {
	Number          string
	Description     string
	DescriptionLang string // GoodsServicesDescription @languageCode
}

// TrademarkRecord represents one register-event (MarkRecord) entry, including
// the de:-namespaced extension details (ownership transfer, cancellation,
// correction history).
type TrademarkRecord struct {
	RecordIdentifier       string
	BasicRecordKind        string
	PublicationIdentifier  string
	PublicationSection     string
	PublicationDate        string
	CurrentStatusCode      string
	CurrentStatusDate      string
	LegalGround            string // de:Record5f/de:LegalGround
	CancellationDate       string // de:Record5f/de:CancellationDate
	CorrectionText         string // de:CorrectionIncomplete/de:CorrectionText
	ImageCorrected         string // de:CorrectionIncomplete/de:ImageCorrected
	ReceiptDeclarationDate string // de:TransferIncomplete/de:ReceiptDeclarationDate
	PreviousHolder         *Party
	NewHolder              *Party
}

// DesignSearchResult holds parsed design search results.
type DesignSearchResult struct {
	TotalHits    int
	DocumentHits int
	DatabaseHits int
	Hits         []DesignHit
	RawXML       []byte // original XML response bytes
}

// DesignHit represents a single design search result entry.
type DesignHit struct {
	DesignIdentifier                 string
	ApplicationNumber                string
	RegistrationNumber               string
	TotalRepresentationSheet         string
	FirstRepresentationSheetFilename string
	Title                            string
	Applicant                        string
	Representative                   string
	StaffName                        string
	ClassNumber                      string
	Status                           string
	ApplicationDate                  string
	RegistrationDate                 string
	PublicationDate                  string
}

// DesignInfo holds parsed design register info (ST86).
type DesignInfo struct {
	RegistrationOfficeCode string
	ApplicationNumber      string
	ApplicationReference   string
	ApplicationDate        string
	TotalDesign            string
	Designs                []DesignEntry // all designs in a multi-design registration
	Applicants             []Party       // application-level applicants
	CorrespondenceParty    Party
	RawXML                 []byte // original XML response bytes

	// Convenience fields mirroring the first design (back-compat).
	DesignIdentifier   string
	RegistrationNumber string
	RegistrationDate   string
	Title              string
	Status             string
	ClassNumber        string
	ClassDescription   string
}

// DesignEntry represents one design within a registration.
type DesignEntry struct {
	DesignIdentifier   string
	RegistrationNumber string
	RegistrationDate   string
	Title              string
	TotalSpecimen      string
	Status             string
	Classes            []DesignClass
	Priorities         []DesignPriority
	Representations    []DesignView // representation sheet views
	PreferedView       *DesignView
	Records            []DesignRecord // register-event history
	Applicants         []Party        // design-level applicants
	Representatives    []Party
	Extension          DesignExtensionInfo
}

// DesignExtensionInfo holds the DE_ST86Extension scalar fields.
type DesignExtensionInfo struct {
	RegisterNumber                string
	DesignFileNumber              string
	DesignNumber                  string
	DateOfCaptureInSystem         string
	TypeOfDeposit                 string
	DefermentStatus               string
	TermProtectionObtained        string
	PaymentDeadline               string
	KindOfFee                     string
	DestructionPaperFile          string
	DesignDescriptionNotPublished string
	CancellationDate              string
}

// DesignClass represents an indication-product classification entry.
type DesignClass struct {
	ClassificationKindCode string
	ClassificationVersion  string
	Comment                string
	ClassNumber            string
	Description            string
}

// DesignPriority represents a priority claim of a design.
type DesignPriority struct {
	CountryCode string
	Number      string
	Date        string
}

// DesignView represents an image/representation view.
type DesignView struct {
	Filename string
	Format   string
	Number   string
	Height   string
	Width    string
	Unit     string
}

// DesignRecord represents one design register-event (DesignRecord) entry.
type DesignRecord struct {
	FilingDate            string
	LanguageCode          string
	PublicationIdentifier string
	PublicationSubsection string
	PublicationDate       string
	LegalStatus           string
}

// --- Private XML structs ---

// Patent search XML
type xmlPatentHitList struct {
	XMLName  xml.Name                 `xml:"PatentHitList"`
	HitCount int                      `xml:"HitCount,attr"`
	Counter  xmlHitCounter            `xml:"Counter"`
	Records  []xmlPatentHitListRecord `xml:"PatentHitListRecord"`
}

type xmlHitCounter struct {
	DocumentHits int `xml:"DocumentHits"`
	DatabaseHits int `xml:"DatabaseHits"`
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
	Inventors               xmlPatentSearchInventors  `xml:"inventors"`
	Agent                   string                    `xml:"agent"`
}

type xmlPatentClassification struct {
	Edition        string `xml:"edition"`
	Classification string `xml:"classification"`
}

type xmlPatentApplicants struct {
	Applicant []string `xml:"applicant"`
}

type xmlPatentSearchInventors struct {
	Inventor []string `xml:"inventor"`
}

// Patent info XML (ST36 with namespace)
type xmlDPMAPatentDocument struct {
	XMLName           xml.Name         `xml:"dpma-patent-document"`
	SchemaVersion     string           `xml:"schema-version,attr"`
	BibliographicData xmlPatentBibData `xml:"bibliographic-data"`
	Events            xmlPatentEvents  `xml:"events"`
}

type xmlPatentBibData struct {
	PublicationRefs     xmlPublicationRefs     `xml:"publication-references"`
	ApplicationRef      xmlApplicationRef      `xml:"application-reference"`
	InventionTitle      xmlInventionTitle      `xml:"invention-title"`
	ClassificationsIPCR xmlClassificationsIPCR `xml:"classifications-ipcr"`
	Parties             xmlParties             `xml:"parties"`
	PriorityClaims      xmlPriorityClaims      `xml:"priority-claims"`
	OfficeSpecific      xmlOfficeSpecificBib   `xml:"office-specific-bib-data"`
	RelatedDocuments    xmlRelatedDocuments    `xml:"related-documents"`
}

type xmlRelatedDocuments struct{}

type xmlPublicationRefs struct {
	Refs []xmlPublicationRef `xml:"publication-reference"`
}

type xmlPublicationRef struct {
	Sequence   string        `xml:"sequence,attr"`
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
	Sequence         string `xml:"sequence,attr"`
	Text             string `xml:"text"`
	VersionIndicator string `xml:"ipc-version-indicator"`
}

type xmlParties struct {
	Applicants            xmlStructuredApplicants `xml:"applicants"`
	Inventors             xmlStructuredInventors  `xml:"inventors"`
	CorrespondenceAddress xmlCorrespondence       `xml:"correspondence-address"`
	Agents                xmlPatentAgents         `xml:"agents"`
}

type xmlStructuredApplicants struct {
	Applicants []xmlStructuredParty `xml:"applicant"`
}

type xmlStructuredInventors struct {
	Inventors []xmlStructuredParty `xml:"inventor"`
}

type xmlPatentAgents struct {
	Agents []xmlStructuredParty `xml:"agent"`
}

type xmlCorrespondence struct {
	Addressbook xmlAddressbook `xml:"addressbook"`
}

type xmlStructuredParty struct {
	Sequence    string         `xml:"sequence,attr"`
	AppType     string         `xml:"app-type,attr"`
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

type xmlPriorityClaims struct {
	Claims []xmlPriorityClaim `xml:"priority-claim"`
}

type xmlPriorityClaim struct {
	Sequence  string `xml:"sequence,attr"`
	Country   string `xml:"country"`
	Date      string `xml:"date"`
	DocNumber string `xml:"doc-number"`
}

type xmlOfficeSpecificBib struct {
	TypeOfIPRight          string `xml:"type-of-ip-right"`
	Status                 string `xml:"status"`
	AgentType              string `xml:"agent-type"`
	NationalFiling         string `xml:"national-filing-data"`
	FirstPublication       string `xml:"date-of-first-publication"`
	Abstract               string `xml:"abstract"`
	DateLastRegisterUpdate string `xml:"date-last-register-update"`
}

type xmlPatentEvents struct {
	Events []xmlPatentEventData `xml:"event-data"`
}

type xmlPatentEventData struct {
	TypeOfProcedure        string                `xml:"type-of-procedure"`
	ProceduralStatus       string                `xml:"procedural-status"`
	DateOfProceduralStatus string                `xml:"date-of-procedural-status"`
	PublicationInfo        *xmlEventPublication  `xml:"publication-info"`
	IPCMainClass           *xmlEventIPCClass     `xml:"ipc-main-class"`
	PreviousIPCMainClass   *xmlEventIPCClass     `xml:"previous-ipc-main-class"`
	PublishedDocuments     xmlPublishedDocuments `xml:"published-documents"`
}

type xmlEventPublication struct {
	IssueNumber     string `xml:"issue-number"`
	Year            string `xml:"year"`
	PublicationDate string `xml:"publication-date"`
	PublicationType string `xml:"publication-type"`
	Part            string `xml:"part"`
}

type xmlEventIPCClass struct {
	Symbol  string `xml:"ipc-symbol"`
	Version string `xml:"ipc-version"`
}

type xmlPublishedDocuments struct {
	Documents []string `xml:"document"`
}

// Trademark search XML
type xmlTrademarkHitList struct {
	XMLName  xml.Name            `xml:"HitList"`
	HitCount int                 `xml:"HitCount,attr"`
	Counter  xmlHitCounter       `xml:"Counter"`
	Entries  []xmlTrademarkEntry `xml:"Entries"`
}

type xmlTrademarkEntry struct {
	Number                 string                     `xml:"Number"`
	RegistrationOfficeCode string                     `xml:"RegistrationOfficeCode"`
	ApplicationNumber      string                     `xml:"ApplicationNumber"`
	MarkDescription        xmlTMSearchMarkDescription `xml:"MarkDescriptionDetails"`
	MarkVerbalText         string                     `xml:"MarkVerbalElementText"`
	MarkFeature            string                     `xml:"MarkFeature"`
	Classification         string                     `xml:"Classification"`
	Status                 string                     `xml:"MarkCurrentStatusCode"`
	ApplicationDate        string                     `xml:"ApplicationDate"`
	RegistrationDate       string                     `xml:"RegistrationDate"`
	Applicant              string                     `xml:"Applicant"`
	Representative         string                     `xml:"Representative"`
}

type xmlTMSearchMarkDescription struct {
	DescriptionText string `xml:"DescriptionText"`
}

// Trademark info XML (ST66 with namespace)
type xmlTrademarkTransaction struct {
	XMLName        xml.Name         `xml:"Transaction"`
	SchemaLocation string           `xml:"schemaLocation,attr"`
	Body           xmlTrademarkBody `xml:"TradeMarkTransactionBody"`
}

type xmlTrademarkBody struct {
	Content xmlTrademarkContent `xml:"TransactionContentDetails"`
}

type xmlTrademarkContent struct {
	TransactionIdentifier string           `xml:"TransactionIdentifier"`
	TransactionCode       string           `xml:"TransactionCode"`
	Data                  xmlTrademarkData `xml:"TransactionData"`
}

type xmlTrademarkData struct {
	Details xmlTrademarkDetails `xml:"TradeMarkDetails"`
}

type xmlTrademarkDetails struct {
	TradeMark xmlTradeMark `xml:"TradeMark"`
}

type xmlTradeMark struct {
	RegistrationOfficeCode    string                     `xml:"RegistrationOfficeCode"`
	ApplicationNumber         string                     `xml:"ApplicationNumber"`
	ApplicationDate           string                     `xml:"ApplicationDate"`
	RegistrationNumber        string                     `xml:"RegistrationNumber"`
	RegistrationDate          string                     `xml:"RegistrationDate"`
	ExpiryDate                string                     `xml:"ExpiryDate"`
	TerminationDate           string                     `xml:"TerminationDate"`
	Status                    string                     `xml:"MarkCurrentStatusCode"`
	KindMark                  string                     `xml:"KindMark"`
	MarkFeature               string                     `xml:"MarkFeature"`
	OppositionPeriodStartDate string                     `xml:"OppositionPeriodStartDate"`
	OppositionPeriodEndDate   string                     `xml:"OppositionPeriodEndDate"`
	WordMark                  xmlWordMarkSpecification   `xml:"WordMarkSpecification"`
	GoodsServices             xmlGoodsServicesDetails    `xml:"GoodsServicesDetails"`
	PublicationDetails        xmlTMPublicationDetails    `xml:"PublicationDetails"`
	ApplicantDetails          xmlTMApplicantDetails      `xml:"ApplicantDetails"`
	RepresentativeDetails     xmlTMRepresentativeDetails `xml:"RepresentativeDetails"`
	CorrespondenceAddress     xmlTMCorrespondenceAddress `xml:"CorrespondenceAddress"`
	MarkRecordDetails         xmlMarkRecordDetails       `xml:"MarkRecordDetails"`
	FileNumber                string                     `xml:"FileNumber"`
	MarkFeatureDPMA           string                     `xml:"MarkFeatureDPMA"`
}

type xmlWordMarkSpecification struct {
	MarkVerbalElementText          xmlLangText `xml:"MarkVerbalElementText"`
	MarkStandardCharacterIndicator string      `xml:"MarkStandardCharacterIndicator"`
}

type xmlLangText struct {
	LanguageCode string `xml:"languageCode,attr"`
	Value        string `xml:",chardata"`
}

type xmlGoodsServicesDetails struct {
	GoodsServices xmlGoodsServices `xml:"GoodsServices"`
}

type xmlGoodsServices struct {
	ProposedLeadingClassNumber string                     `xml:"ProposedLeadingClassNumber"`
	ClassDescriptions          xmlClassDescriptionDetails `xml:"ClassDescriptionDetails"`
}

type xmlClassDescriptionDetails struct {
	ClassDescriptions []xmlClassDescription `xml:"ClassDescription"`
}

type xmlClassDescription struct {
	ClassNumber string      `xml:"ClassNumber"`
	Description xmlLangText `xml:"GoodsServicesDescription"`
}

type xmlTMPublicationDetails struct {
	Publications []xmlTMPublication `xml:"Publication"`
}

type xmlTMPublication struct {
	PublicationDate string `xml:"PublicationDate"`
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

type xmlTMRepresentativeDetails struct {
	Representatives []xmlTMRepresentative `xml:"Representative"`
}

type xmlTMRepresentative struct {
	AddressBook xmlTMRepresentativeAddressBook `xml:"RepresentativeAddressBook"`
}

type xmlTMRepresentativeAddressBook struct {
	FormattedNameAddress xmlFormattedNameAddress `xml:"FormattedNameAddress"`
}

type xmlTMCorrespondenceAddress struct {
	AddressBook xmlTMCorrespondenceAddressBook `xml:"CorrespondenceAddressBook"`
}

type xmlTMCorrespondenceAddressBook struct {
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
	Street   string `xml:"AddressStreet"`
	City     string `xml:"AddressCity"`
	County   string `xml:"AddressCounty"` // often used for country code
	State    string `xml:"AddressState"`
	Postcode string `xml:"AddressPostcode"`
}

// Trademark register-event tree (MarkRecordDetails) incl. de: extension records.
type xmlMarkRecordDetails struct {
	Records []xmlMarkRecord `xml:"MarkRecord"`
}

type xmlMarkRecord struct {
	BasicRecord               xmlBasicRecord               `xml:"BasicRecord"`
	Record5f                  xmlRecord5f                  `xml:"Record5f"`
	RecordProcedureIncomplete xmlRecordProcedureIncomplete `xml:"RecordProcedureIncomplete"`
}

type xmlBasicRecord struct {
	RecordIdentifier   string                      `xml:"RecordIdentifier"`
	BasicRecordKind    string                      `xml:"BasicRecordKind"`
	PublicationDetails xmlRecordPublicationDetails `xml:"RecordPublicationDetails"`
	CurrentStatusCode  string                      `xml:"RecordCurrentStatusCode"`
	CurrentStatusDate  string                      `xml:"RecordCurrentStatusDate"`
}

type xmlRecordPublicationDetails struct {
	Publication xmlRecordPublication `xml:"RecordPublication"`
}

type xmlRecordPublication struct {
	PublicationIdentifier string `xml:"PublicationIdentifier"`
	PublicationSection    string `xml:"PublicationSection"`
	PublicationDate       string `xml:"PublicationDate"`
}

type xmlRecord5f struct {
	LegalGround      string `xml:"LegalGround"`
	CancellationDate string `xml:"CancellationDate"`
}

type xmlRecordProcedureIncomplete struct {
	ApplicationIncompleteType xmlEmptyMarker          `xml:"ApplicationIncompleteType"`
	CancellationIncomplete    xmlEmptyMarker          `xml:"CancellationIncomplete"`
	CorrectionIncomplete      xmlCorrectionIncomplete `xml:"CorrectionIncomplete"`
	TransferIncomplete        xmlTransferIncomplete   `xml:"TransferIncomplete"`
}

type xmlEmptyMarker struct{}

type xmlCorrectionIncomplete struct {
	CorrectionText string `xml:"CorrectionText"`
	ImageCorrected string `xml:"ImageCorrected"`
}

type xmlTransferIncomplete struct {
	ReceiptDeclarationDate string                    `xml:"ReceiptDeclarationDate"`
	HolderDetails          xmlHolderDetails          `xml:"HolderDetails"`
	CorrespondenceDetails  xmlTransferCorrespondence `xml:"CorrespondenceAddressDetails"`
}

type xmlHolderDetails struct {
	PreviousHolder xmlTMApplicant `xml:"PreviousHolder"`
	NewHolder      xmlTMApplicant `xml:"NewHolder"`
}

type xmlTransferCorrespondence struct {
	NewCorrespondenceAddress xmlTMCorrespondenceAddress `xml:"NewCorrespondenceAddress"`
}

// Design search XML
type xmlDesignHitList struct {
	XMLName  xml.Name                 `xml:"DesignHitList"`
	HitCount int                      `xml:"HitCount,attr"`
	Counter  xmlHitCounter            `xml:"Counter"`
	Records  []xmlDesignHitListRecord `xml:"DesignHitListRecord"`
}

type xmlDesignHitListRecord struct {
	DesignIdentifier                 string `xml:"DesignIdentifier"`
	ApplicationNumber                string `xml:"DesignApplicationNumber"`
	RegistrationNumber               string `xml:"RegistrationNumber"`
	TotalRepresentationSheet         string `xml:"TotalRepresentationSheet"`
	FirstRepresentationSheetFilename string `xml:"FirstRepresentationSheetFilename"`
	Title                            string `xml:"ProductDescription"`
	Applicant                        string `xml:"OwnerOrApplicant"`
	Representative                   string `xml:"Representative"`
	StaffName                        string `xml:"StaffName"`
	ClassNumber                      string `xml:"ClassNumber"`
	Status                           string `xml:"DesignStatus"`
	ApplicationDate                  string `xml:"DesignApplicationDate"`
	RegistrationDate                 string `xml:"RegistrationDate"`
	PublicationDate                  string `xml:"PublicationDate"`
}

// Design info XML (ST86 with namespace)
type xmlDesignTransaction struct {
	XMLName        xml.Name      `xml:"Transaction"`
	SchemaLocation string        `xml:"schemaLocation,attr"`
	Body           xmlDesignBody `xml:"DesignTransactionBody"`
}

type xmlDesignBody struct {
	Content xmlDesignContent `xml:"TransactionContentDetails"`
}

type xmlDesignContent struct {
	TransactionIdentifier string                   `xml:"TransactionIdentifier"`
	TransactionCode       string                   `xml:"TransactionCode"`
	Data                  xmlDesignTransactionData `xml:"TransactionData"`
}

type xmlDesignTransactionData struct {
	ApplicationDetails xmlDesignApplicationDetails `xml:"DesignApplicationDetails"`
}

type xmlDesignApplicationDetails struct {
	Application xmlDesignApplication `xml:"DesignApplication"`
}

type xmlDesignApplication struct {
	RegistrationOfficeCode string                  `xml:"RegistrationOfficeCode"`
	ApplicationNumber      string                  `xml:"DesignApplicationNumber"`
	ApplicationDate        string                  `xml:"DesignApplicationDate"`
	ApplicationReference   string                  `xml:"DesignApplicationReference"`
	TotalDesign            string                  `xml:"TotalDesign"`
	Designs                xmlDesignDetails        `xml:"DesignDetails"`
	Applicants             xmlDesignApplicants     `xml:"ApplicantDetails"`
	CorrespondenceAddress  xmlDesignCorrespondence `xml:"CorrespondenceAddress"`
}

type xmlDesignDetails struct {
	Designs []xmlDesign `xml:"Design"`
}

type xmlDesign struct {
	DesignIdentifier   string                         `xml:"DesignIdentifier"`
	RegistrationNumber string                         `xml:"RegistrationNumber"`
	RegistrationDate   string                         `xml:"RegistrationDate"`
	Title              string                         `xml:"DesignTitle"`
	TotalSpecimen      string                         `xml:"TotalSpecimen"`
	Extension          xmlDesignExtension             `xml:"DE_ST86Extension"`
	Representations    xmlDesignRepresentationSheet   `xml:"DesignRepresentationSheetDetails"`
	Products           xmlIndicationProductDetails    `xml:"IndicationProductDetails"`
	Priorities         xmlDesignPriorityDetails       `xml:"PriorityDetails"`
	PublicationDetails xmlDesignPublicationDetails    `xml:"PublicationDetails"`
	Applicants         xmlDesignApplicants            `xml:"ApplicantDetails"`
	Representatives    xmlDesignRepresentativeDetails `xml:"RepresentativeDetails"`
}

type xmlDesignExtension struct {
	DesignInfo                    xmlDesignInfoExt       `xml:"DesignInfo"`
	Status                        string                 `xml:"DesignStatus"`
	DateOfCaptureInSystem         string                 `xml:"DateOfCaptureInSystem"`
	TypeOfDeposit                 string                 `xml:"TypeOfDeposit"`
	DefermentStatus               string                 `xml:"DefermentStatus"`
	TermProtectionObtained        string                 `xml:"TermProtectionObtained"`
	PaymentDeadline               string                 `xml:"PaymentDeadline"`
	KindOfFee                     string                 `xml:"KindOfFee"`
	DestructionPaperFile          string                 `xml:"DestructionPaperFile"`
	DesignDescriptionNotPublished string                 `xml:"DesignDescriptionNotPublished"`
	CancellationDate              string                 `xml:"CancellationDate"`
	PreferedRepresentation        xmlDesignView          `xml:"DesignPreferedRepresentation"`
	RecordDetails                 xmlDesignRecordDetails `xml:"DesignRecordDetails"`
}

type xmlDesignInfoExt struct {
	RegisterNumber   string `xml:"RegisterNumber"`
	DesignFileNumber string `xml:"DesignFileNumber"`
	DesignNumber     string `xml:"DesignNumber"`
}

type xmlDesignRepresentationSheet struct {
	Sheets []xmlDesignSheet `xml:"DesignRepresentationSheet"`
}

type xmlDesignSheet struct {
	ViewDetails xmlDesignViewDetails `xml:"ViewDetails"`
}

type xmlDesignViewDetails struct {
	Views []xmlDesignView `xml:"View"`
}

type xmlDesignView struct {
	Filename string               `xml:"ViewFilename"`
	Format   string               `xml:"ViewFileFormat"`
	Number   string               `xml:"ViewNumber"`
	Size     xmlDesignViewRepSize `xml:"ViewRepresentationSize"`
}

type xmlDesignViewRepSize struct {
	Height string `xml:"Height"`
	Width  string `xml:"Width"`
	Unit   string `xml:"Unit"`
}

type xmlDesignRecordDetails struct {
	Records []xmlDesignRecord `xml:"DesignRecord"`
}

type xmlDesignRecord struct {
	BasicRecord xmlDesignBasicRecord `xml:"DesignBasicRecord"`
}

type xmlDesignBasicRecord struct {
	FilingDate         string                            `xml:"RecordFilingDate"`
	LanguageCode       string                            `xml:"RecordLanguageCode"`
	PublicationDetails xmlDesignRecordPublicationDetails `xml:"RecordPublicationDetails"`
	LegalStatus        string                            `xml:"LegalStatus"`
}

type xmlDesignRecordPublicationDetails struct {
	Publication xmlDesignRecordPublication `xml:"RecordPublication"`
}

type xmlDesignRecordPublication struct {
	PublicationIdentifier string `xml:"PublicationIdentifier"`
	PublicationSubsection string `xml:"PublicationSubsection"`
	PublicationDate       string `xml:"PublicationDate"`
}

type xmlIndicationProductDetails struct {
	Products []xmlIndicationProduct `xml:"IndicationProduct"`
}

type xmlIndicationProduct struct {
	ClassificationKindCode string                           `xml:"ClassificationKindCode"`
	ClassificationVersion  string                           `xml:"ClassificationVersion"`
	Comment                string                           `xml:"Comment"`
	ClassDescriptions      xmlDesignClassDescriptionDetails `xml:"ClassDescriptionDetails"`
}

type xmlDesignClassDescriptionDetails struct {
	Descriptions []xmlDesignClassDescription `xml:"ClassDescription"`
}

type xmlDesignClassDescription struct {
	ClassNumber string `xml:"ClassNumber"`
	Description string `xml:"ProductDescription"`
}

type xmlDesignPriorityDetails struct {
	Priorities []xmlDesignPriority `xml:"Priority"`
}

type xmlDesignPriority struct {
	CountryCode string `xml:"PriorityCountryCode"`
	Number      string `xml:"PriorityNumber"`
	Date        string `xml:"PriorityDate"`
}

type xmlDesignPublicationDetails struct {
	Publications []xmlDesignPublication `xml:"Publication"`
}

type xmlDesignPublication struct {
	PublicationIdentifier string `xml:"PublicationIdentifier"`
	PublicationDate       string `xml:"PublicationDate"`
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

type xmlDesignRepresentativeDetails struct {
	Representatives []xmlDesignRepresentative `xml:"Representative"`
}

type xmlDesignRepresentative struct {
	AddressBook xmlDesignRepresentativeAddressBook `xml:"RepresentativeAddressBook"`
}

type xmlDesignRepresentativeAddressBook struct {
	FormattedNameAddress xmlDesignFormattedNameAddress `xml:"FormattedNameAddress"`
}

type xmlDesignCorrespondence struct {
	AddressBook xmlDesignCorrespondenceAddressBook `xml:"CorrespondenceAddressBook"`
}

type xmlDesignCorrespondenceAddressBook struct {
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
	PostOfficeBox string `xml:"AddressPostOfficeBox"`
	Street        string `xml:"AddressStreet"`
	City          string `xml:"AddressCity"`
	County        string `xml:"AddressCounty"`
	State         string `xml:"AddressState"`
	Postcode      string `xml:"AddressPostcode"`
	CountryCode   string `xml:"FormattedAddressCountryCode"`
}

// --- Conversion helpers ---

func partyFromFreeFormat(fna xmlFormattedNameAddress) Party {
	addr := fna.Address.FormattedAddress
	country := addr.County
	if country == "" {
		country = addr.State
	}
	return Party{
		Name:     fna.Name.FreeFormatName.Details.NameLine,
		Address:  addr.City,
		Street:   addr.Street,
		Postcode: addr.Postcode,
		Country:  country,
	}
}

func partyFromDesignFormatted(fna xmlDesignFormattedNameAddress) Party {
	addr := fna.Address.FormattedAddress
	country := addr.County
	if country == "" {
		country = addr.State
	}
	if country == "" {
		country = addr.CountryCode
	}
	return Party{
		Name:     fna.Name.FormattedName.OrgName,
		Address:  addr.City,
		Street:   addr.Street,
		Postcode: addr.Postcode,
		Country:  country,
	}
}

// extractDesignApplicants converts raw XML design applicants to Party values.
func extractDesignApplicants(raw []xmlDesignApplicant) []Party {
	parties := make([]Party, 0, len(raw))
	for _, a := range raw {
		parties = append(parties, partyFromDesignFormatted(a.AddressBook.FormattedNameAddress))
	}
	return parties
}

func extractDesignRepresentatives(raw []xmlDesignRepresentative) []Party {
	parties := make([]Party, 0, len(raw))
	for _, r := range raw {
		parties = append(parties, partyFromDesignFormatted(r.AddressBook.FormattedNameAddress))
	}
	return parties
}

// --- Parse functions ---

// ParsePatentSearch parses a patent search XML response.
func ParsePatentSearch(data []byte) (*PatentSearchResult, error) {
	if err := parseDPMAError(data, 0); err != nil {
		return nil, err
	}
	var raw xmlPatentHitList
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, &XMLParseError{Operation: "ParsePatentSearch", Err: err}
	}

	result := &PatentSearchResult{
		TotalHits:    raw.HitCount,
		DocumentHits: raw.Counter.DocumentHits,
		DatabaseHits: raw.Counter.DatabaseHits,
		Hits:         make([]PatentHit, len(raw.Records)),
		RawXML:       data,
	}

	for i := range raw.Records {
		r := &raw.Records[i]

		main := PatentSearchIPC{
			Classification: r.MainClassification.Classification,
			Edition:        r.MainClassification.Edition,
		}
		var classes []PatentSearchIPC
		var further []PatentSearchIPC
		var ipcSymbols []string
		if main.Classification != "" {
			classes = append(classes, main)
			ipcSymbols = append(ipcSymbols, main.Classification)
		}
		for _, fc := range r.FurtherClassifications {
			c := PatentSearchIPC{Classification: fc.Classification, Edition: fc.Edition}
			further = append(further, c)
			if c.Classification != "" {
				classes = append(classes, c)
				ipcSymbols = append(ipcSymbols, c.Classification)
			}
		}

		result.Hits[i] = PatentHit{
			LeadingRegisteredNumber: r.LeadingRegisteredNumber,
			RegisteredNumber:        r.RegisteredNumber,
			Type:                    r.Type,
			LegalStatus:             r.LegalStatus,
			Title:                   r.Title,
			Applicants:              r.Applicants.Applicant,
			Inventors:               r.Inventors.Inventor,
			Agent:                   r.Agent,
			IPCClasses:              ipcSymbols,
			Classifications:         classes,
			MainClassification:      main,
			FurtherClassifications:  further,
			ApplicationDate:         r.ApplicationDate,
			PublicationDate:         r.PublicationDate,
			RegistrationDate:        r.RegistrationDate,
		}
	}

	return result, nil
}

// ParsePatentInfo parses a patent info XML response (ST36 format).
func ParsePatentInfo(data []byte) (*PatentInfo, error) {
	if err := parseDPMAError(data, 0); err != nil {
		return nil, err
	}
	var raw xmlDPMAPatentDocument
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, &XMLParseError{Operation: "ParsePatentInfo", Err: err}
	}

	bib := raw.BibliographicData

	var publications []PatentPublication
	for _, ref := range bib.PublicationRefs.Refs {
		publications = append(publications, PatentPublication{
			Sequence: ref.Sequence,
			Country:  ref.DocumentID.Country,
			Number:   ref.DocumentID.DocNumber,
			Kind:     ref.DocumentID.Kind,
			Date:     ref.DocumentID.Date,
		})
	}

	var ipcClasses []string
	var classifications []PatentClassification
	for _, c := range bib.ClassificationsIPCR.Classifications {
		classifications = append(classifications, PatentClassification{
			Sequence:         c.Sequence,
			Symbol:           c.Text,
			VersionIndicator: c.VersionIndicator,
		})
		if c.Text != "" {
			ipcClasses = append(ipcClasses, c.Text)
		}
	}

	applicants := partiesFromStructured(bib.Parties.Applicants.Applicants)
	inventors := partiesFromStructured(bib.Parties.Inventors.Inventors)
	agents := partiesFromStructured(bib.Parties.Agents.Agents)

	var priorities []PatentPriorityClaim
	for _, pc := range bib.PriorityClaims.Claims {
		// Field layout mirrors xmlPriorityClaim exactly; convert directly.
		priorities = append(priorities, PatentPriorityClaim(pc))
	}

	var events []PatentEvent
	for _, e := range raw.Events.Events {
		ev := PatentEvent{
			TypeOfProcedure:        e.TypeOfProcedure,
			ProceduralStatus:       e.ProceduralStatus,
			DateOfProceduralStatus: e.DateOfProceduralStatus,
			PublishedDocuments:     e.PublishedDocuments.Documents,
		}
		if e.PublicationInfo != nil {
			ev.PublicationInfo = &PatentEventPublication{
				IssueNumber:     e.PublicationInfo.IssueNumber,
				Year:            e.PublicationInfo.Year,
				PublicationDate: e.PublicationInfo.PublicationDate,
				PublicationType: e.PublicationInfo.PublicationType,
				Part:            e.PublicationInfo.Part,
			}
		}
		if e.IPCMainClass != nil {
			ev.IPCMainClass = &PatentIPCClass{Symbol: e.IPCMainClass.Symbol, Version: e.IPCMainClass.Version}
		}
		if e.PreviousIPCMainClass != nil {
			ev.PreviousIPCMainClass = &PatentIPCClass{Symbol: e.PreviousIPCMainClass.Symbol, Version: e.PreviousIPCMainClass.Version}
		}
		events = append(events, ev)
	}

	info := &PatentInfo{
		SchemaVersion: raw.SchemaVersion,
		Publications:  publications,
		ApplicationRef: PatentDocumentRef{
			Country: bib.ApplicationRef.DocumentID.Country,
			Number:  bib.ApplicationRef.DocumentID.DocNumber,
			Date:    bib.ApplicationRef.DocumentID.Date,
		},
		Title:                  bib.InventionTitle.Value,
		TitleLang:              bib.InventionTitle.Lang,
		IPCClasses:             ipcClasses,
		Classifications:        classifications,
		Applicants:             applicants,
		Inventors:              inventors,
		Agents:                 agents,
		CorrespondenceParty:    partyFromAddressbook(bib.Parties.CorrespondenceAddress.Addressbook),
		PriorityClaims:         priorities,
		IPRightType:            bib.OfficeSpecific.TypeOfIPRight,
		Status:                 bib.OfficeSpecific.Status,
		AgentType:              bib.OfficeSpecific.AgentType,
		Abstract:               bib.OfficeSpecific.Abstract,
		FilingDate:             bib.OfficeSpecific.NationalFiling,
		FirstPubDate:           bib.OfficeSpecific.FirstPublication,
		DateLastRegisterUpdate: bib.OfficeSpecific.DateLastRegisterUpdate,
		Events:                 events,
		RawXML:                 data,
	}
	return info, nil
}

func partyFromAddressbook(ab xmlAddressbook) Party {
	return Party{
		Text:    ab.Text,
		Name:    ab.Name,
		Address: ab.Address.Address1,
		Country: ab.Address.Country,
	}
}

func partiesFromStructured(raw []xmlStructuredParty) []Party {
	if len(raw) == 0 {
		return nil
	}
	parties := make([]Party, 0, len(raw))
	for _, p := range raw {
		party := partyFromAddressbook(p.Addressbook)
		party.Sequence = p.Sequence
		party.AppType = p.AppType
		parties = append(parties, party)
	}
	return parties
}

// ParseTrademarkSearch parses a trademark search XML response.
func ParseTrademarkSearch(data []byte) (*TrademarkSearchResult, error) {
	if err := parseDPMAError(data, 0); err != nil {
		return nil, err
	}
	var raw xmlTrademarkHitList
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, &XMLParseError{Operation: "ParseTrademarkSearch", Err: err}
	}

	result := &TrademarkSearchResult{
		TotalHits:    raw.HitCount,
		DocumentHits: raw.Counter.DocumentHits,
		DatabaseHits: raw.Counter.DatabaseHits,
		Hits:         make([]TrademarkHit, len(raw.Entries)),
		RawXML:       data,
	}

	for i := range raw.Entries {
		e := &raw.Entries[i]
		result.Hits[i] = TrademarkHit{
			Number:                 e.Number,
			RegistrationOfficeCode: e.RegistrationOfficeCode,
			ApplicationNumber:      e.ApplicationNumber,
			MarkText:               e.MarkVerbalText,
			MarkDescriptionText:    e.MarkDescription.DescriptionText,
			MarkFeature:            e.MarkFeature,
			Classification:         e.Classification,
			Status:                 e.Status,
			ApplicationDate:        e.ApplicationDate,
			RegistrationDate:       e.RegistrationDate,
			Applicant:              e.Applicant,
			Representative:         e.Representative,
		}
	}

	return result, nil
}

// ParseTrademarkInfo parses a trademark info XML response (ST66 format).
func ParseTrademarkInfo(data []byte) (*TrademarkInfo, error) {
	if err := parseDPMAError(data, 0); err != nil {
		return nil, err
	}
	var raw xmlTrademarkTransaction
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, &XMLParseError{Operation: "ParseTrademarkInfo", Err: err}
	}

	tm := raw.Body.Content.Data.Details.TradeMark

	var applicants []Party
	for _, a := range tm.ApplicantDetails.Applicants {
		applicants = append(applicants, partyFromFreeFormat(a.AddressBook.FormattedNameAddress))
	}

	var representatives []Party
	for _, r := range tm.RepresentativeDetails.Representatives {
		representatives = append(representatives, partyFromFreeFormat(r.AddressBook.FormattedNameAddress))
	}

	var classes []TrademarkClass
	for _, cd := range tm.GoodsServices.GoodsServices.ClassDescriptions.ClassDescriptions {
		classes = append(classes, TrademarkClass{
			Number:          cd.ClassNumber,
			Description:     cd.Description.Value,
			DescriptionLang: cd.Description.LanguageCode,
		})
	}

	var records []TrademarkRecord
	for _, mr := range tm.MarkRecordDetails.Records {
		rec := TrademarkRecord{
			RecordIdentifier:       mr.BasicRecord.RecordIdentifier,
			BasicRecordKind:        mr.BasicRecord.BasicRecordKind,
			PublicationIdentifier:  mr.BasicRecord.PublicationDetails.Publication.PublicationIdentifier,
			PublicationSection:     mr.BasicRecord.PublicationDetails.Publication.PublicationSection,
			PublicationDate:        mr.BasicRecord.PublicationDetails.Publication.PublicationDate,
			CurrentStatusCode:      mr.BasicRecord.CurrentStatusCode,
			CurrentStatusDate:      mr.BasicRecord.CurrentStatusDate,
			LegalGround:            mr.Record5f.LegalGround,
			CancellationDate:       mr.Record5f.CancellationDate,
			CorrectionText:         mr.RecordProcedureIncomplete.CorrectionIncomplete.CorrectionText,
			ImageCorrected:         mr.RecordProcedureIncomplete.CorrectionIncomplete.ImageCorrected,
			ReceiptDeclarationDate: mr.RecordProcedureIncomplete.TransferIncomplete.ReceiptDeclarationDate,
		}
		holders := mr.RecordProcedureIncomplete.TransferIncomplete.HolderDetails
		if name := holders.PreviousHolder.AddressBook.FormattedNameAddress.Name.FreeFormatName.Details.NameLine; name != "" {
			p := partyFromFreeFormat(holders.PreviousHolder.AddressBook.FormattedNameAddress)
			rec.PreviousHolder = &p
		}
		if name := holders.NewHolder.AddressBook.FormattedNameAddress.Name.FreeFormatName.Details.NameLine; name != "" {
			p := partyFromFreeFormat(holders.NewHolder.AddressBook.FormattedNameAddress)
			rec.NewHolder = &p
		}
		records = append(records, rec)
	}

	info := &TrademarkInfo{
		RegistrationOfficeCode:     tm.RegistrationOfficeCode,
		ApplicationNumber:          tm.ApplicationNumber,
		RegistrationNumber:         tm.RegistrationNumber,
		ApplicationDate:            tm.ApplicationDate,
		RegistrationDate:           tm.RegistrationDate,
		ExpiryDate:                 tm.ExpiryDate,
		TerminationDate:            tm.TerminationDate,
		Status:                     tm.Status,
		KindMark:                   tm.KindMark,
		MarkFeature:                tm.MarkFeature,
		OppositionPeriodStartDate:  tm.OppositionPeriodStartDate,
		OppositionPeriodEndDate:    tm.OppositionPeriodEndDate,
		MarkText:                   tm.WordMark.MarkVerbalElementText.Value,
		MarkTextLang:               tm.WordMark.MarkVerbalElementText.LanguageCode,
		StandardCharacterIndicator: tm.WordMark.MarkStandardCharacterIndicator,
		ProposedLeadingClassNumber: tm.GoodsServices.GoodsServices.ProposedLeadingClassNumber,
		Applicants:                 applicants,
		Representatives:            representatives,
		CorrespondenceParty:        partyFromFreeFormat(tm.CorrespondenceAddress.AddressBook.FormattedNameAddress),
		Classifications:            classes,
		Records:                    records,
		FileNumber:                 tm.FileNumber,
		MarkFeatureDPMA:            tm.MarkFeatureDPMA,
		RawXML:                     data,
	}
	return info, nil
}

// ParseDesignSearch parses a design search XML response.
func ParseDesignSearch(data []byte) (*DesignSearchResult, error) {
	if err := parseDPMAError(data, 0); err != nil {
		return nil, err
	}
	var raw xmlDesignHitList
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, &XMLParseError{Operation: "ParseDesignSearch", Err: err}
	}

	result := &DesignSearchResult{
		TotalHits:    raw.HitCount,
		DocumentHits: raw.Counter.DocumentHits,
		DatabaseHits: raw.Counter.DatabaseHits,
		Hits:         make([]DesignHit, len(raw.Records)),
		RawXML:       data,
	}

	for i := range raw.Records {
		r := &raw.Records[i]
		result.Hits[i] = DesignHit{
			DesignIdentifier:                 r.DesignIdentifier,
			ApplicationNumber:                r.ApplicationNumber,
			RegistrationNumber:               r.RegistrationNumber,
			TotalRepresentationSheet:         r.TotalRepresentationSheet,
			FirstRepresentationSheetFilename: r.FirstRepresentationSheetFilename,
			Title:                            r.Title,
			Applicant:                        r.Applicant,
			Representative:                   r.Representative,
			StaffName:                        r.StaffName,
			ClassNumber:                      r.ClassNumber,
			Status:                           r.Status,
			ApplicationDate:                  r.ApplicationDate,
			RegistrationDate:                 r.RegistrationDate,
			PublicationDate:                  r.PublicationDate,
		}
	}

	return result, nil
}

// ParseDesignInfo parses a design info XML response (ST86 format).
func ParseDesignInfo(data []byte) (*DesignInfo, error) {
	if err := parseDPMAError(data, 0); err != nil {
		return nil, err
	}
	var raw xmlDesignTransaction
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, &XMLParseError{Operation: "ParseDesignInfo", Err: err}
	}

	app := raw.Body.Content.Data.ApplicationDetails.Application

	info := &DesignInfo{
		RegistrationOfficeCode: app.RegistrationOfficeCode,
		ApplicationNumber:      app.ApplicationNumber,
		ApplicationReference:   app.ApplicationReference,
		ApplicationDate:        app.ApplicationDate,
		TotalDesign:            app.TotalDesign,
		Applicants:             extractDesignApplicants(app.Applicants.Applicants),
		CorrespondenceParty:    partyFromDesignFormatted(app.CorrespondenceAddress.AddressBook.FormattedNameAddress),
		RawXML:                 data,
	}

	for i := range app.Designs.Designs {
		info.Designs = append(info.Designs, convertDesign(&app.Designs.Designs[i]))
	}

	// Convenience fields mirror the first design.
	if len(info.Designs) > 0 {
		d := info.Designs[0]
		info.DesignIdentifier = d.DesignIdentifier
		info.RegistrationNumber = d.RegistrationNumber
		info.RegistrationDate = d.RegistrationDate
		info.Title = d.Title
		info.Status = d.Status
		if len(d.Classes) > 0 {
			info.ClassNumber = d.Classes[0].ClassNumber
			info.ClassDescription = d.Classes[0].Description
		}
		// Design-level applicants take priority over application-level.
		if len(d.Applicants) > 0 {
			info.Applicants = d.Applicants
		}
	}

	return info, nil
}

func convertDesign(d *xmlDesign) DesignEntry {
	ext := d.Extension
	entry := DesignEntry{
		DesignIdentifier:   d.DesignIdentifier,
		RegistrationNumber: d.RegistrationNumber,
		RegistrationDate:   d.RegistrationDate,
		Title:              d.Title,
		TotalSpecimen:      d.TotalSpecimen,
		Status:             ext.Status,
		Applicants:         extractDesignApplicants(d.Applicants.Applicants),
		Representatives:    extractDesignRepresentatives(d.Representatives.Representatives),
		Extension: DesignExtensionInfo{
			RegisterNumber:                ext.DesignInfo.RegisterNumber,
			DesignFileNumber:              ext.DesignInfo.DesignFileNumber,
			DesignNumber:                  ext.DesignInfo.DesignNumber,
			DateOfCaptureInSystem:         ext.DateOfCaptureInSystem,
			TypeOfDeposit:                 ext.TypeOfDeposit,
			DefermentStatus:               ext.DefermentStatus,
			TermProtectionObtained:        ext.TermProtectionObtained,
			PaymentDeadline:               ext.PaymentDeadline,
			KindOfFee:                     ext.KindOfFee,
			DestructionPaperFile:          ext.DestructionPaperFile,
			DesignDescriptionNotPublished: ext.DesignDescriptionNotPublished,
			CancellationDate:              ext.CancellationDate,
		},
	}

	if v := ext.PreferedRepresentation; v.Filename != "" || v.Number != "" {
		view := convertDesignView(v)
		entry.PreferedView = &view
	}

	for _, sheet := range d.Representations.Sheets {
		for _, v := range sheet.ViewDetails.Views {
			entry.Representations = append(entry.Representations, convertDesignView(v))
		}
	}

	for _, p := range d.Products.Products {
		for _, cd := range p.ClassDescriptions.Descriptions {
			entry.Classes = append(entry.Classes, DesignClass{
				ClassificationKindCode: p.ClassificationKindCode,
				ClassificationVersion:  p.ClassificationVersion,
				Comment:                p.Comment,
				ClassNumber:            cd.ClassNumber,
				Description:            cd.Description,
			})
		}
	}

	for _, pr := range d.Priorities.Priorities {
		// Field layout mirrors xmlDesignPriority exactly; convert directly.
		entry.Priorities = append(entry.Priorities, DesignPriority(pr))
	}

	for _, rec := range ext.RecordDetails.Records {
		entry.Records = append(entry.Records, DesignRecord{
			FilingDate:            rec.BasicRecord.FilingDate,
			LanguageCode:          rec.BasicRecord.LanguageCode,
			PublicationIdentifier: rec.BasicRecord.PublicationDetails.Publication.PublicationIdentifier,
			PublicationSubsection: rec.BasicRecord.PublicationDetails.Publication.PublicationSubsection,
			PublicationDate:       rec.BasicRecord.PublicationDetails.Publication.PublicationDate,
			LegalStatus:           rec.BasicRecord.LegalStatus,
		})
	}

	return entry
}

func convertDesignView(v xmlDesignView) DesignView {
	return DesignView{
		Filename: v.Filename,
		Format:   v.Format,
		Number:   v.Number,
		Height:   v.Size.Height,
		Width:    v.Size.Width,
		Unit:     v.Size.Unit,
	}
}
