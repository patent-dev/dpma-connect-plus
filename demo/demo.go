// Package main provides a comprehensive demo of the DPMA Connect Plus Go client.
//
// This demo showcases all three DPMA services:
//   - Patent Service (search, info, PDF, weekly bulk, register extract)
//   - Design Service (search, info, images, weekly bulk, register extract)
//   - Trademark Service (search, info, images, weekly bulk, register extract)
//
// All request/response pairs are saved to demo/examples/ for reference.
//
// Usage:
//
//	export DPMA_CONNECT_PLUS_USERNAME="your-username"
//	export DPMA_CONNECT_PLUS_PASSWORD="your-password"
//
//	# Run all demos
//	./demo
//
//	# Run specific service
//	./demo -service=patent
//	./demo -service=design
//	./demo -service=trademark
//
//	# Run in interactive mode
//	./demo -interactive
//
//	# Skip saving examples
//	./demo -no-save
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	dpma "github.com/patent-dev/dpma-connect-plus"
)

var (
	username    = flag.String("username", os.Getenv("DPMA_CONNECT_PLUS_USERNAME"), "DPMA username")
	password    = flag.String("password", os.Getenv("DPMA_CONNECT_PLUS_PASSWORD"), "DPMA password")
	service     = flag.String("service", "", "Service filter (patent|design|trademark)")
	interactive = flag.Bool("interactive", false, "Run in interactive mode")
	examplesDir = flag.String("examples", "examples", "Directory to save examples")
	skipSave    = flag.Bool("no-save", false, "Skip saving request/response files")
)

// DemoContext holds shared context for all demos
type DemoContext struct {
	Client   *dpma.Client
	Ctx      context.Context
	Saver    *ExampleSaver
	SkipSave bool
}

// saveExample saves a request/response example if saving is enabled
func (dctx *DemoContext) saveExample(name string, params map[string]string, response []byte, format FileFormat) {
	if dctx.Saver == nil {
		return
	}
	requestDesc := FormatRequestDescription(name, params)
	if err := dctx.Saver.SaveExample(name, requestDesc, response, format); err != nil {
		fmt.Printf("Warning: failed to save example for %s: %v\n", name, err)
	}
}

func main() {
	flag.Parse()

	u := *username
	p := *password

	if u == "" || p == "" {
		reader := bufio.NewReader(os.Stdin)
		if u == "" {
			fmt.Print("Enter your DPMA username: ")
			u, _ = reader.ReadString('\n')
			u = strings.TrimSpace(u)
		}
		if p == "" {
			fmt.Print("Enter your DPMA password: ")
			p, _ = reader.ReadString('\n')
			p = strings.TrimSpace(p)
		}
	} else {
		fmt.Println("Using credentials from environment (DPMA_CONNECT_PLUS_USERNAME/PASSWORD)")
	}

	config := dpma.DefaultConfig()
	config.Username = u
	config.Password = p
	client, err := dpma.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	dctx := &DemoContext{
		Client:   client,
		Ctx:      ctx,
		SkipSave: *skipSave,
	}
	if !*skipSave {
		dctx.Saver = NewExampleSaver(*examplesDir)
	}

	if *interactive {
		runInteractive(ctx, client)
		return
	}

	switch *service {
	case "patent":
		demoPatent(dctx)
	case "design":
		demoDesign(dctx)
	case "trademark":
		demoTrademark(dctx)
	case "":
		runAll(dctx)
	default:
		log.Fatalf("Unknown service: %s", *service)
	}
}

func runAll(dctx *DemoContext) {
	demoPatent(dctx)
	demoDesign(dctx)
	demoTrademark(dctx)
}

func runInteractive(ctx context.Context, client *dpma.Client) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n=== DPMA Connect Plus Demo ===")
		fmt.Println("--- Patents ---")
		fmt.Println("1. Test service version")
		fmt.Println("2. Download single patent PDF")
		fmt.Println("3. Download publication week XML (Disclosure Documents)")
		fmt.Println("4. Download publication week XML (Patent Specifications)")
		fmt.Println("5. Download publication week XML (Utility Models)")
		fmt.Println("6. Search patents")
		fmt.Println("7. Get patent info")
		fmt.Println("--- Designs ---")
		fmt.Println("8. Search designs")
		fmt.Println("9. Get design info")
		fmt.Println("10. Download design bibliographic data XML")
		fmt.Println("--- Trademarks ---")
		fmt.Println("11. Search trademarks")
		fmt.Println("12. Get trademark info")
		fmt.Println("13. Download trademark bib data (applied)")
		fmt.Println("0. Exit")
		fmt.Print("\nSelect option: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		option, _ := strconv.Atoi(input)

		switch option {
		case 1:
			interactiveTestVersion(ctx, client, reader)
		case 2:
			interactiveDownloadPatentPDF(ctx, client, reader)
		case 3:
			interactiveDownloadWeekly(ctx, client, reader, "disclosure documents", func(y, w int) ([]byte, error) { return client.GetDisclosureDocumentsXML(ctx, y, w) })
		case 4:
			interactiveDownloadWeekly(ctx, client, reader, "patent specifications", func(y, w int) ([]byte, error) { return client.GetPatentSpecificationsXML(ctx, y, w) })
		case 5:
			interactiveDownloadWeekly(ctx, client, reader, "utility models", func(y, w int) ([]byte, error) { return client.GetUtilityModelsXML(ctx, y, w) })
		case 6:
			interactiveSearch(ctx, client, reader, "patents", func(q string) ([]byte, error) { return client.SearchPatents(ctx, q) })
		case 7:
			interactiveGetInfo(ctx, client, reader, "patent", "registered number", func(id string) ([]byte, error) { return client.GetPatentInfo(ctx, id) })
		case 8:
			interactiveSearch(ctx, client, reader, "designs", func(q string) ([]byte, error) { return client.SearchDesigns(ctx, q) })
		case 9:
			interactiveGetInfo(ctx, client, reader, "design", "design number", func(id string) ([]byte, error) { return client.GetDesignInfo(ctx, id) })
		case 10:
			interactiveDownloadWeekly(ctx, client, reader, "design bib data", func(y, w int) ([]byte, error) { return client.GetDesignBibliographicDataXML(ctx, y, w) })
		case 11:
			interactiveSearch(ctx, client, reader, "trademarks", func(q string) ([]byte, error) { return client.SearchTrademarks(ctx, q) })
		case 12:
			interactiveGetInfo(ctx, client, reader, "trademark", "application number", func(id string) ([]byte, error) { return client.GetTrademarkInfo(ctx, id) })
		case 13:
			interactiveDownloadWeekly(ctx, client, reader, "trademark bib data (applied)", func(y, w int) ([]byte, error) { return client.GetTrademarkBibDataApplied(ctx, y, w) })
		case 0:
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid option")
		}
	}
}

func interactiveTestVersion(ctx context.Context, client *dpma.Client, reader *bufio.Reader) {
	fmt.Println("\n=== Test Service Version ===")
	fmt.Println("1. Patent Service")
	fmt.Println("2. Design Service")
	fmt.Println("3. Trademark Service")
	fmt.Print("Select service: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	option, _ := strconv.Atoi(input)

	services := map[int]string{
		1: dpma.ServicePatent,
		2: dpma.ServiceDesign,
		3: dpma.ServiceTrademark,
	}

	svc, ok := services[option]
	if !ok {
		fmt.Println("Invalid option")
		return
	}

	fmt.Printf("\nFetching version for %s...\n", svc)
	version, err := client.GetVersion(ctx, svc)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Version: %s\n", version)
}

func interactiveDownloadPatentPDF(ctx context.Context, client *dpma.Client, reader *bufio.Reader) {
	fmt.Println("\n=== Download Patent PDF ===")
	fmt.Print("Enter document ID (e.g., DE102023000001A1): ")

	docID, _ := reader.ReadString('\n')
	docID = strings.TrimSpace(docID)
	if docID == "" {
		fmt.Println("Document ID required")
		return
	}

	fmt.Printf("\nDownloading %s...\n", docID)
	start := time.Now()

	data, err := client.GetPatentPublicationPDF(ctx, docID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	filename := fmt.Sprintf("%s.pdf", docID)
	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	fmt.Printf("Downloaded %d bytes to %s in %v\n", len(data), filename, time.Since(start))
}

func interactiveDownloadWeekly(_ context.Context, _ *dpma.Client, reader *bufio.Reader, name string, fn func(int, int) ([]byte, error)) {
	year, week := getPublicationWeek(reader)
	if year == 0 {
		return
	}

	fmt.Printf("\nDownloading %s for %d week %d...\n", name, year, week)
	start := time.Now()

	data, err := fn(year, week)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	filename := fmt.Sprintf("%s_%04d%02d.zip", strings.ReplaceAll(name, " ", "_"), year, week)
	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	fmt.Printf("Downloaded %d bytes to %s in %v\n", len(data), filename, time.Since(start))
}

func interactiveSearch(_ context.Context, _ *dpma.Client, reader *bufio.Reader, name string, fn func(string) ([]byte, error)) {
	fmt.Printf("\n=== Search %s ===\n", name)
	fmt.Print("Enter search query: ")

	query, _ := reader.ReadString('\n')
	query = strings.TrimSpace(query)
	if query == "" {
		fmt.Println("Query required")
		return
	}

	fmt.Printf("\nSearching %s for: %s...\n", name, query)
	start := time.Now()

	data, err := fn(query)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Received %d bytes in %v\n", len(data), time.Since(start))

	filename := fmt.Sprintf("search_%s.xml", name)
	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}
	fmt.Printf("Results saved to %s\n", filename)
}

func interactiveGetInfo(_ context.Context, _ *dpma.Client, reader *bufio.Reader, name, idName string, fn func(string) ([]byte, error)) {
	fmt.Printf("\n=== Get %s Info ===\n", name)
	fmt.Printf("Enter %s: ", idName)

	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)
	if id == "" {
		fmt.Printf("%s required\n", idName)
		return
	}

	fmt.Printf("\nFetching %s info for %s...\n", name, id)
	data, err := fn(id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Received %d bytes of XML data\n", len(data))
}

func getPublicationWeek(reader *bufio.Reader) (year, week int) {
	fmt.Print("Enter year (e.g., 2024): ")
	yearStr, _ := reader.ReadString('\n')
	yearStr = strings.TrimSpace(yearStr)
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		fmt.Println("Invalid year")
		return 0, 0
	}

	fmt.Print("Enter week (1-53): ")
	weekStr, _ := reader.ReadString('\n')
	weekStr = strings.TrimSpace(weekStr)
	week, err = strconv.Atoi(weekStr)
	if err != nil || week < 1 || week > 53 {
		fmt.Println("Invalid week")
		return 0, 0
	}

	return year, week
}
