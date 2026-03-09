package main

import (
	"fmt"

	dpma "github.com/patent-dev/dpma-connect-plus"
	dpmaquery "github.com/patent-dev/dpma-connect-plus/query"
)

func demoPatent(dctx *DemoContext) {
	printHeader("Patent Service Demonstrations")

	demoGetPatentVersion(dctx)
	demoSearchPatents(dctx)
	demoGetPatentInfo(dctx)
	demoGetPatentPublicationPDF(dctx)
	demoGetSearchableFullText(dctx)
	demoGetDisclosureDocumentsXML(dctx)
}

func demoGetPatentVersion(dctx *DemoContext) {
	printSubHeader("GetVersion (Patent Service)")

	version, err := dctx.Client.GetVersion(dctx.Ctx, dpma.ServicePatent)
	if err != nil {
		printError(err)
		return
	}

	printResult("Version", version)
	dctx.saveExample("get_version_patent", map[string]string{
		"service": dpma.ServicePatent,
	}, []byte(version), FormatText)
}

func demoSearchPatents(dctx *DemoContext) {
	printSubHeader("SearchPatents")

	searchQuery := "TI=Elektrofahrzeug"

	// Validate query before sending
	q, err := dpmaquery.ParseQuery(searchQuery, dpmaquery.ServicePatent)
	if err != nil {
		printError(err)
		return
	}
	if err := q.Validate(); err != nil {
		printError(err)
		return
	}
	printResult("Query (validated)", searchQuery)
	printResult("Fields", fmt.Sprintf("%v", q.GetFields()))

	data, err := dctx.Client.SearchPatents(dctx.Ctx, searchQuery)
	if err != nil {
		printError(err)
		return
	}

	printResult("Query", searchQuery)
	printResult("Response size", fmt.Sprintf("%d bytes", len(data)))
	printResult("Preview", truncate(string(data), 200))

	dctx.saveExample("search_patents", map[string]string{
		"query": searchQuery,
	}, data, DetectFormat(data))
}

func demoGetPatentInfo(dctx *DemoContext) {
	printSubHeader("GetPatentInfo")

	data, err := dctx.Client.GetPatentInfo(dctx.Ctx, TestPatentRegNumber)
	if err != nil {
		printError(err)
		return
	}

	printResult("Registered number", TestPatentRegNumber)
	printResult("Response size", fmt.Sprintf("%d bytes", len(data)))
	printResult("Preview", truncate(string(data), 200))

	dctx.saveExample("get_patent_info", map[string]string{
		"registeredNumber": TestPatentRegNumber,
	}, data, DetectFormat(data))
}

func demoGetPatentPublicationPDF(dctx *DemoContext) {
	printSubHeader("GetPatentPublicationPDF")

	data, err := dctx.Client.GetPatentPublicationPDF(dctx.Ctx, TestPatentDocID)
	if err != nil {
		printError(err)
		return
	}

	printResult("Document ID", TestPatentDocID)
	printResult("Response size", fmt.Sprintf("%d bytes", len(data)))

	dctx.saveExample("get_patent_publication_pdf", map[string]string{
		"documentID": TestPatentDocID,
	}, data, FormatBin)
}

func demoGetSearchableFullText(dctx *DemoContext) {
	printSubHeader("GetSearchableFullText")

	data, err := dctx.Client.GetSearchableFullText(dctx.Ctx, TestPatentDocID)
	if err != nil {
		printError(err)
		return
	}

	printResult("Document ID", TestPatentDocID)
	printResult("Response size", fmt.Sprintf("%d bytes", len(data)))
	printResult("Preview", truncate(string(data), 200))

	dctx.saveExample("get_searchable_full_text", map[string]string{
		"documentID": TestPatentDocID,
	}, data, DetectFormat(data))
}

func demoGetDisclosureDocumentsXML(dctx *DemoContext) {
	printSubHeader("GetDisclosureDocumentsXML")

	year, week := 2026, 9
	data, err := dctx.Client.GetDisclosureDocumentsXML(dctx.Ctx, year, week)
	if err != nil {
		printError(err)
		return
	}

	printResult("Year/Week", fmt.Sprintf("%d/%d", year, week))
	printResult("Response size", fmt.Sprintf("%d bytes", len(data)))

	dctx.saveExample("get_disclosure_documents_xml", map[string]string{
		"year": fmt.Sprintf("%d", year),
		"week": fmt.Sprintf("%d", week),
	}, data, FormatBin)
}
