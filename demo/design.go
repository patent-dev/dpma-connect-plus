package main

import (
	"fmt"

	dpma "github.com/patent-dev/dpma-connect-plus"
	dpmaquery "github.com/patent-dev/dpma-connect-plus/query"
)

func demoDesign(dctx *DemoContext) {
	printHeader("Design Service Demonstrations")

	demoGetDesignVersion(dctx)
	demoSearchDesigns(dctx)
	demoGetDesignInfo(dctx)
	demoGetDesignImage(dctx)
	demoGetDesignThumbnail(dctx)
}

func demoGetDesignVersion(dctx *DemoContext) {
	printSubHeader("GetVersion (Design Service)")

	version, err := dctx.Client.GetVersion(dctx.Ctx, dpma.ServiceDesign)
	if err != nil {
		printError(err)
		return
	}

	printResult("Version", version)
	dctx.saveExample("get_version_design", map[string]string{
		"service": dpma.ServiceDesign,
	}, []byte(version), FormatText)
}

func demoSearchDesigns(dctx *DemoContext) {
	printSubHeader("SearchDesigns")

	searchQuery := "INH=Samsung"

	// Validate query before sending
	q, err := dpmaquery.ParseQuery(searchQuery, dpmaquery.ServiceDesign)
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

	data, err := dctx.Client.SearchDesigns(dctx.Ctx, searchQuery)
	if err != nil {
		printError(err)
		return
	}

	printResult("Query", searchQuery)
	printResult("Response size", fmt.Sprintf("%d bytes", len(data)))
	printResult("Preview", truncate(string(data), 200))

	dctx.saveExample("search_designs", map[string]string{
		"query": searchQuery,
	}, data, DetectFormat(data))
}

func demoGetDesignImage(dctx *DemoContext) {
	printSubHeader("GetDesignImage")

	data, err := dctx.Client.GetDesignImage(dctx.Ctx, TestDesignNumber, "1")
	if err != nil {
		printError(err)
		return
	}

	printResult("Design number", TestDesignNumber)
	printResult("Image number", "1")
	printResult("Response size", fmt.Sprintf("%d bytes", len(data)))

	dctx.saveExample("get_design_image", map[string]string{
		"designNumber": TestDesignNumber,
		"imageNumber":  "1",
	}, data, FormatBin)
}

func demoGetDesignThumbnail(dctx *DemoContext) {
	printSubHeader("GetDesignThumbnail")

	data, err := dctx.Client.GetDesignThumbnail(dctx.Ctx, TestDesignNumber, "1")
	if err != nil {
		printError(err)
		return
	}

	printResult("Design number", TestDesignNumber)
	printResult("Thumbnail number", "1")
	printResult("Response size", fmt.Sprintf("%d bytes", len(data)))

	dctx.saveExample("get_design_thumbnail", map[string]string{
		"designNumber":    TestDesignNumber,
		"thumbnailNumber": "1",
	}, data, FormatBin)
}

func demoGetDesignInfo(dctx *DemoContext) {
	printSubHeader("GetDesignInfo")

	data, err := dctx.Client.GetDesignInfo(dctx.Ctx, TestDesignNumber)
	if err != nil {
		printError(err)
		return
	}

	printResult("Design number", TestDesignNumber)
	printResult("Response size", fmt.Sprintf("%d bytes", len(data)))
	printResult("Preview", truncate(string(data), 200))

	dctx.saveExample("get_design_info", map[string]string{
		"designNumber": TestDesignNumber,
	}, data, DetectFormat(data))
}
