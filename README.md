# DPMA Connect Plus Go Client

[![Go Reference](https://pkg.go.dev/badge/github.com/patent-dev/dpma-connect-plus.svg)](https://pkg.go.dev/github.com/patent-dev/dpma-connect-plus)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go client library for the German Patent and Trademark Office (DPMA) Connect Plus REST API.

## Getting Started

### Authentication

DPMA Connect Plus is a paid service requiring authentication:

- **Service Details**: https://www.dpma.de/recherche/datenabgabe/dpmaconnect/index.html
- **Access**: Contact DPMA to obtain credentials
- **Authentication Method**: HTTP Basic Authentication

## Installation

```bash
go get github.com/patent-dev/dpma-connect-plus
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    dpma "github.com/patent-dev/dpma-connect-plus"
)

func main() {
    client, err := dpma.NewClient(&dpma.Config{
        Username: "your-dpma-username",
        Password: "your-dpma-password",
    })
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Get service version
    version, err := client.GetVersion(ctx, dpma.ServicePatent)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Version: %s\n", version)

    // Validate query before sending (optional but recommended)
    if err := dpma.ValidatePatentQuery("TI=Elektrofahrzeug"); err != nil {
        log.Fatal(err)
    }

    // Search patents (uses DPMAregister expert search syntax)
    results, err := client.SearchPatents(ctx, "TI=Elektrofahrzeug")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Results: %d bytes\n", len(results))
}
```

## API Methods

All methods accept `context.Context` for timeout and cancellation support.

### Common Operations

```go
// Get service version
GetVersion(ctx, service string) (string, error)
```

Use the service constants: `dpma.ServicePatent`, `dpma.ServiceDesign`, `dpma.ServiceTrademark`.

### Query Validation

```go
// Validate a query against service-specific field codes (returns nil or error)
ValidatePatentQuery(query string) error
ValidateDesignQuery(query string) error
ValidateTrademarkQuery(query string) error
```

For advanced usage (tokenization, field inspection), use the `query` sub-package directly.

### Patent Service

```go
// Search patents and utility models
SearchPatents(ctx, query string) ([]byte, error)

// Get patent info by registered number
GetPatentInfo(ctx, registeredNumber string) ([]byte, error)

// Get searchable full text for a document
GetSearchableFullText(ctx, documentID string) ([]byte, error)

// Download single patent publication PDF
GetPatentPublicationPDF(ctx, documentID string) ([]byte, error)

// Weekly bulk downloads (XML)
GetDisclosureDocumentsXML(ctx, year, week int) ([]byte, error)
GetPatentSpecificationsXML(ctx, year, week int) ([]byte, error)
GetUtilityModelsXML(ctx, year, week int) ([]byte, error)
GetPublicationDataXML(ctx, year, week int) ([]byte, error)
GetApplicantCitationsXML(ctx, year, week int) ([]byte, error)
GetEuropeanPatentSpecificationsXML(ctx, year, week int) ([]byte, error)

// Weekly bulk downloads (PDF)
GetDisclosureDocumentsPDF(ctx, year, week int) ([]byte, error)
GetPatentSpecificationsPDF(ctx, year, week int) ([]byte, error)
GetEuropeanPatentSpecificationsPDF(ctx, year, week int) ([]byte, error)
GetUtilityModelsPDF(ctx, year, week int) ([]byte, error)

// Register extract
GetPatentRegisterExtract(ctx, date time.Time, period string) ([]byte, error)
```

### Design Service

```go
// Search designs
SearchDesigns(ctx, query string) ([]byte, error)

// Get design info by design number
GetDesignInfo(ctx, designNumber string) ([]byte, error)

// Get design image/thumbnail
GetDesignImage(ctx, designNumber, imageNumber string) ([]byte, error)
GetDesignThumbnail(ctx, designNumber, thumbnailNumber string) ([]byte, error)

// Weekly bulk downloads
GetDesignBibliographicDataXML(ctx, year, week int) ([]byte, error)
GetDesignImages(ctx, year, week int) ([]byte, error)

// Register extract
GetDesignRegisterExtract(ctx, date time.Time, period string) ([]byte, error)
```

### Trademark Service

```go
// Search trademarks
SearchTrademarks(ctx, query string) ([]byte, error)

// Get trademark info by application number
GetTrademarkInfo(ctx, applicationNumber string) ([]byte, error)

// Get trademark image/thumbnail
GetTrademarkImage(ctx, applicationNumber string) ([]byte, error)
GetTrademarkThumbnail(ctx, applicationNumber string) ([]byte, error)

// Weekly bulk downloads
GetTrademarkBibDataApplied(ctx, year, week int) ([]byte, error)
GetTrademarkBibDataRegistered(ctx, year, week int) ([]byte, error)
GetTrademarkBibDataRejected(ctx, year, week int) ([]byte, error)

// Register extract
GetTrademarkRegisterExtract(ctx, date time.Time, period string) ([]byte, error)
```

### Streaming Downloads (Memory-Efficient)

Every bulk download and register extract method has a `*Stream` variant that writes to an `io.Writer`:

```go
// Patent streams
GetDisclosureDocumentsXMLStream(ctx, year, week int, dst io.Writer) error
GetPatentSpecificationsXMLStream(ctx, year, week int, dst io.Writer) error
GetUtilityModelsXMLStream(ctx, year, week int, dst io.Writer) error
GetPublicationDataXMLStream(ctx, year, week int, dst io.Writer) error
GetApplicantCitationsXMLStream(ctx, year, week int, dst io.Writer) error
GetEuropeanPatentSpecificationsXMLStream(ctx, year, week int, dst io.Writer) error
GetDisclosureDocumentsPDFStream(ctx, year, week int, dst io.Writer) error
GetPatentSpecificationsPDFStream(ctx, year, week int, dst io.Writer) error
GetEuropeanPatentSpecificationsPDFStream(ctx, year, week int, dst io.Writer) error
GetUtilityModelsPDFStream(ctx, year, week int, dst io.Writer) error
GetPatentRegisterExtractStream(ctx, date time.Time, period string, dst io.Writer) error

// Design streams
GetDesignBibliographicDataXMLStream(ctx, year, week int, dst io.Writer) error
GetDesignImagesStream(ctx, year, week int, dst io.Writer) error
GetDesignRegisterExtractStream(ctx, date time.Time, period string, dst io.Writer) error

// Trademark streams
GetTrademarkBibDataAppliedStream(ctx, year, week int, dst io.Writer) error
GetTrademarkBibDataRegisteredStream(ctx, year, week int, dst io.Writer) error
GetTrademarkBibDataRejectedStream(ctx, year, week int, dst io.Writer) error
GetTrademarkRegisterExtractStream(ctx, date time.Time, period string, dst io.Writer) error
```

## Configuration

```go
config := &dpma.Config{
    Username: "your-username",           // Required
    Password: "your-password",           // Required
    BaseURL:  "https://dpmaconnect.dpma.de/dpmaws/rest-services", // Default
    Timeout:  20 * time.Minute,           // Request timeout (default: 20 minutes)
    HTTPClient: myCustomHTTPClient,      // Optional: provide your own *http.Client
}

client, err := dpma.NewClient(config)
```

## Usage Examples

### Download Patent PDF

```go
ctx := context.Background()
pdf, err := client.GetPatentPublicationPDF(ctx, "DE102023000001A1")
if err != nil {
    log.Fatal(err)
}

err = os.WriteFile("patent.pdf", pdf, 0644)
if err != nil {
    log.Fatal(err)
}
```

### Download Publication Week Data

```go
ctx := context.Background()

zipData, err := client.GetDisclosureDocumentsXML(ctx, 2024, 45)
if err != nil {
    if _, ok := err.(*dpma.DataNotAvailableError); ok {
        fmt.Println("Data not available for this week")
        return
    }
    log.Fatal(err)
}

err = os.WriteFile("disclosure_202445.zip", zipData, 0644)
if err != nil {
    log.Fatal(err)
}
```

### Stream Large Files

```go
ctx := context.Background()

file, err := os.Create("patents_202445.zip")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

err = client.GetPatentSpecificationsXMLStream(ctx, 2024, 45, file)
if err != nil {
    log.Fatal(err)
}
```

### Search Designs

```go
ctx := context.Background()

// Validate query against design field codes
if err := dpma.ValidateDesignQuery("INH=Samsung"); err != nil {
    log.Fatal(err)
}

results, err := client.SearchDesigns(ctx, "INH=Samsung")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Received %d bytes of XML results\n", len(results))
```

### Register Extract

```go
ctx := context.Background()

date := time.Date(2024, 10, 23, 0, 0, 0, 0, time.UTC)
data, err := client.GetPatentRegisterExtract(ctx, date, dpma.PeriodDaily)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Register extract: %d bytes\n", len(data))
```

## Query Validation

The `query` sub-package provides a parser and validator for DPMAregister expert search syntax:

```go
import "github.com/patent-dev/dpma-connect-plus/query"

// Parse and validate a patent query
q, err := query.ParseQuery("TI=Elektrofahrzeug AND INH=Siemens", query.ServicePatent)
if err != nil {
    log.Fatal(err)
}
if err := q.Validate(); err != nil {
    log.Fatal(err) // e.g. unknown field, unmatched parentheses
}

// Inspect the query
fmt.Println(q.GetFields())  // ["TI", "INH"]
fmt.Println(q.HasField("TI")) // true

// Validate against a specific service
q, _ = query.ParseQuery("MARKE=test", query.ServiceTrademark) // valid
q, _ = query.ParseQuery("MARKE=test", query.ServicePatent)    // invalid: unknown field

// Look up field definitions
f, ok := query.GetField("TI", query.ServicePatent)
fmt.Println(f.Description) // "title / designation"

// List all valid fields for a service
fields := query.GetValidFields(query.ServiceDesign)
```

### Features

- Tokenizer with support for quoted values, comparison operators (`=`, `>=`, `<=`, `>`, `<`), parentheses, and curly braces (procedure data)
- Field validation per service (Patent, Design, Trademark) based on official DPMAregister field codes
- Bracket and brace matching
- Recognizes both English (`AND`, `OR`, `NOT`) and German (`UND`, `ODER`, `NICHT`) Boolean operators
- Field lookup with German name, English description, and input type (text/date)

## Search Query Syntax

All search methods use [DPMAregister expert search syntax](https://register.dpma.de/DPMAregister/pat/experte). The format is `FIELD=value` with Boolean operators `AND`, `OR`, `NOT`.

### Patent Field Codes

| Code | Description |
|------|-------------|
| `TI` | Title/designation |
| `INH` | Applicant/proprietor |
| `IN` | Inventor |
| `IC` | IPC classification |
| `AKZ` | File number |
| `PN` | Publication number |
| `PUB` | Publication date |
| `AB` | Abstract |

### Design Field Codes

| Code | Description |
|------|-------------|
| `TI` | Designation |
| `INH` | Proprietor |
| `ENTW` | Designer |
| `ERZ` | Product(s) |
| `WKL` | Commodity class |
| `RN` | Registration number |
| `DNR` | Design number |

### Trademark Field Codes

| Code | Description |
|------|-------------|
| `md` | Trademark text |
| `INH` | Proprietor |
| `WKL` | Class(es) |

For a full reference, see the DPMAregister help pages:
- [Patent field codes](https://register.dpma.de/register/htdocs/prod/de/hilfe/recherchefelder/patgbm/index.html)
- [Design field codes](https://register.dpma.de/register/htdocs/prod/de/hilfe/recherchefelder/gsm/index.html)
- [Trademark field codes](https://register.dpma.de/register/htdocs/prod/de/hilfe/recherchefelder/marken/index.html)

## Error Handling

The library provides custom error types for different scenarios:

```go
// Resource not found (404)
if notFoundErr, ok := err.(*dpma.NotFoundError); ok {
    fmt.Printf("Not found: %s %s\n", notFoundErr.Resource, notFoundErr.ID)
}

// Data not available (common for old/future publication weeks)
if _, ok := err.(*dpma.DataNotAvailableError); ok {
    fmt.Println("Data not available for the requested period")
}

// Generic API errors
if apiErr, ok := err.(*dpma.APIError); ok {
    fmt.Printf("API error: %s (code: %s, HTTP %d)\n", apiErr.Message, apiErr.Code, apiErr.StatusCode)
}

// XML parsing failures (malformed response data)
if xmlErr, ok := err.(*dpma.XMLParseError); ok {
    fmt.Printf("Failed to parse response in %s: %v\n", xmlErr.Operation, xmlErr.Unwrap())
}
```

## Date and Week Formatting

### Publication Week Format

Publication weeks use `YYYYWW` format (6 digits):

```go
pubWeek, err := dpma.FormatPublicationWeek(2024, 45)  // Returns "202445", nil

year, week, err := dpma.ParsePublicationWeek("202445")  // Returns 2024, 45, nil
```

### Register Extract Date Format

Register extracts use `YYYY-MM-DD` format:

```go
date := time.Date(2024, 10, 23, 0, 0, 0, 0, time.UTC)
dateStr := dpma.FormatDate(date)  // Returns "2024-10-23"
```

## Services

DPMA Connect Plus provides three services:

| Service | Description | Constant |
|---------|-------------|----------|
| Patents & Utility Models | Patent applications, grants, utility models | `dpma.ServicePatent` |
| Designs | Design applications and registrations | `dpma.ServiceDesign` |
| Trademarks | Trademark applications and registrations | `dpma.ServiceTrademark` |

## Package Structure

```
├── client.go              # Core client (Config, NewClient, GetVersion)
├── client_patent.go       # Patent service methods
├── client_design.go       # Design service methods
├── client_trademark.go    # Trademark service methods
├── errors.go              # Custom error types
├── helpers.go             # Date/week formatting, constants
├── query/                 # Query parser and field validation
│   ├── query.go           # Parser, tokenizer, validator
│   ├── fields.go          # Field definitions per service
│   └── query_test.go      # Query package tests
├── client_test.go         # Core unit tests
├── client_patent_test.go  # Patent unit tests
├── client_design_test.go  # Design unit tests
├── client_trademark_test.go # Trademark unit tests
├── integration_test.go    # Integration tests (real API)
├── generated/             # Auto-generated OpenAPI code
│   ├── types_gen.go       # Generated types
│   └── client_gen.go      # Generated client
├── openapi.yaml           # OpenAPI 3.0 specification
└── demo/                  # Interactive demo application
    └── demo.go
```

## Implementation

This library follows a clean architecture:

1. **OpenAPI Specification**: Hand-crafted `openapi.yaml` based on official DPMA documentation
2. **Code Generation**: Types and client generated using [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen)
3. **Idiomatic Wrapper**: Clean Go client wrapping generated code with error handling and convenience methods

## Testing

### Unit Tests (Mock Server)

Offline tests using mock HTTP server with realistic responses:

```bash
go test -v
go test -v -cover
```

### Integration Tests (Real API)

Tests that make actual requests to the DPMA API:

```bash
export DPMA_CONNECT_PLUS_USERNAME=your-username
export DPMA_CONNECT_PLUS_PASSWORD=your-password

go test -tags=integration -v
```

**Note**: Integration tests require valid DPMA Connect Plus credentials and will skip if environment variables are not set.

## Demo Application

An interactive demo application is included:

```bash
export DPMA_CONNECT_PLUS_USERNAME=your-username
export DPMA_CONNECT_PLUS_PASSWORD=your-password

cd demo
go run demo.go
```

The demo provides an interactive menu for testing all three services: patents, designs, and trademarks.

## Implementation Status

All endpoints defined in the DPMA Connect Plus API are implemented. Verification status depends on the permissions granted to the test account.

### Fully Verified

The following endpoints have been tested against the live API with real data:

| Service | Endpoint | Description |
|---------|----------|-------------|
| General | `GetVersion` | Service version info |
| Patent | `SearchPatent` | Expert search |
| Patent | `GetRegisterInfo` | Register data for a patent |
| Patent | `GetFulltextXML` / `Stream` | Full-text XML |
| Patent | `GetFulltextPDF` / `Stream` | Full-text PDF |
| Patent | `GetDocumentIdList` | Document IDs for a publication week |
| Patent | `GetBulkFulltextXML` / `Stream` | Bulk full-text XML |
| Patent | `GetBulkFulltextPDF` / `Stream` | Bulk full-text PDF |
| Patent | `GetRegisterExtractXML` / `Stream` | Register extract |
| Patent | `GetDisclosureDocumentsXML` / `Stream` | Weekly disclosure documents |
| Design | `SearchDesign` | Expert search |
| Design | `GetDesignRegisterInfo` | Register data for a design |
| Trademark | `SearchTrademark` | Expert search |
| Trademark | `GetTrademarkRegisterInfo` | Register data for a trademark |

### Not Yet Verified (Permission-Restricted)

The following endpoints are implemented but could not be verified with the current test account. We are in talks with DPMA to achieve broader test coverage.

| Service | Endpoint | Description |
|---------|----------|-------------|
| Patent | `GetPublicationDataXML` / `Stream` | Weekly publication data |
| Patent | `GetApplicantCitationsXML` / `Stream` | Applicant citations |
| Design | `GetDesignBibliographicDataXML` / `Stream` | Bibliographic bulk data |
| Design | `GetDesignImages` / `Stream` | Design images bulk download |
| Trademark | `GetTrademarkBibDataApplied` / `Stream` | Applied trademarks bulk data |
| Trademark | `GetTrademarkBibDataRegistered` / `Stream` | Registered trademarks bulk data |
| Trademark | `GetTrademarkBibDataRejected` / `Stream` | Rejected trademarks bulk data |

## Data Availability

- **Publication Data**: Updated weekly
  - Patents/Utility Models: Thursdays
  - Designs/Trademarks: Fridays
- **Register Data**: Updated daily
- **Historical Data**: Varies by document type

**Note**: "Data not available" errors are common for:
- Future publication weeks
- Old publication weeks (before digital archiving)
- Weeks with no publications

## Development

### Regenerating from OpenAPI

If the OpenAPI spec is updated:

```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

oapi-codegen -package generated -generate types openapi.yaml > generated/types_gen.go
oapi-codegen -package generated -generate client openapi.yaml > generated/client_gen.go
```

### Code Quality

```bash
go fmt ./...
go vet ./...
go test -v ./...
```

## Similar Projects

This project follows the style and quality standards of:
- [patent-dev/uspto-odp](https://github.com/patent-dev/uspto-odp) - USPTO Open Data Portal Go client
- [patent-dev/epo-bdds](https://github.com/patent-dev/epo-bdds) - EPO BDDS Go client

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Credits

**Developed by:**
- Wolfgang Stark - [patent.dev](https://patent.dev) - [Funktionslust GmbH](https://funktionslust.digital)
