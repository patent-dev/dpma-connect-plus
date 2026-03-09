package main

import (
	"fmt"

	dpma "github.com/patent-dev/dpma-connect-plus"
	dpmaquery "github.com/patent-dev/dpma-connect-plus/query"
)

func demoTrademark(dctx *DemoContext) {
	printHeader("Trademark Service Demonstrations")

	demoGetTrademarkVersion(dctx)
	demoSearchTrademarks(dctx)
	demoGetTrademarkInfo(dctx)
	demoGetTrademarkImage(dctx)
	demoGetTrademarkThumbnail(dctx)
}

func demoGetTrademarkVersion(dctx *DemoContext) {
	printSubHeader("GetVersion (Trademark Service)")

	version, err := dctx.Client.GetVersion(dctx.Ctx, dpma.ServiceTrademark)
	if err != nil {
		printError(err)
		return
	}

	printResult("Version", version)
	dctx.saveExample("get_version_trademark", map[string]string{
		"service": dpma.ServiceTrademark,
	}, []byte(version), FormatText)
}

func demoSearchTrademarks(dctx *DemoContext) {
	printSubHeader("SearchTrademarks")

	searchQuery := "md=Apple"

	// Validate query before sending
	q, err := dpmaquery.ParseQuery(searchQuery, dpmaquery.ServiceTrademark)
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

	data, err := dctx.Client.SearchTrademarks(dctx.Ctx, searchQuery)
	if err != nil {
		printError(err)
		return
	}

	printResult("Query", searchQuery)
	printResult("Response size", fmt.Sprintf("%d bytes", len(data)))
	printResult("Preview", truncate(string(data), 200))

	dctx.saveExample("search_trademarks", map[string]string{
		"query": searchQuery,
	}, data, DetectFormat(data))
}

func demoGetTrademarkImage(dctx *DemoContext) {
	printSubHeader("GetTrademarkImage")

	data, err := dctx.Client.GetTrademarkImage(dctx.Ctx, TestTrademarkAppNum)
	if err != nil {
		printError(err)
		return
	}

	printResult("Application number", TestTrademarkAppNum)
	printResult("Response size", fmt.Sprintf("%d bytes", len(data)))

	dctx.saveExample("get_trademark_image", map[string]string{
		"applicationNumber": TestTrademarkAppNum,
	}, data, FormatBin)
}

func demoGetTrademarkThumbnail(dctx *DemoContext) {
	printSubHeader("GetTrademarkThumbnail")

	data, err := dctx.Client.GetTrademarkThumbnail(dctx.Ctx, TestTrademarkAppNum)
	if err != nil {
		printError(err)
		return
	}

	printResult("Application number", TestTrademarkAppNum)
	printResult("Response size", fmt.Sprintf("%d bytes", len(data)))

	dctx.saveExample("get_trademark_thumbnail", map[string]string{
		"applicationNumber": TestTrademarkAppNum,
	}, data, FormatBin)
}

func demoGetTrademarkInfo(dctx *DemoContext) {
	printSubHeader("GetTrademarkInfo")

	data, err := dctx.Client.GetTrademarkInfo(dctx.Ctx, TestTrademarkAppNum)
	if err != nil {
		printError(err)
		return
	}

	printResult("Application number", TestTrademarkAppNum)
	printResult("Response size", fmt.Sprintf("%d bytes", len(data)))
	printResult("Preview", truncate(string(data), 200))

	dctx.saveExample("get_trademark_info", map[string]string{
		"applicationNumber": TestTrademarkAppNum,
	}, data, DetectFormat(data))
}
