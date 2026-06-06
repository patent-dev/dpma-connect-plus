# DPMA Connect Plus Go Client

[![CI](https://github.com/patent-dev/dpma-connect-plus/actions/workflows/ci.yml/badge.svg)](https://github.com/patent-dev/dpma-connect-plus/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/patent-dev/dpma-connect-plus.svg)](https://pkg.go.dev/github.com/patent-dev/dpma-connect-plus)
[![Go Report Card](https://goreportcard.com/badge/github.com/patent-dev/dpma-connect-plus)](https://goreportcard.com/report/github.com/patent-dev/dpma-connect-plus)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A Go client for the German Patent and Trade Mark Office (DPMA) Connect Plus REST
API (DPMAregister web services), covering German patent, design, and trademark
data over HTTP Basic authentication.

## Overview

- **Three services** - patents and utility models, designs, and trademarks,
  selected with the `dpma.ServicePatent`, `dpma.ServiceDesign`, and
  `dpma.ServiceTrademark` constants.
- **Expert search** - `SearchPatents`, `SearchDesigns`, `SearchTrademarks` using
  DPMAregister expert search syntax, with optional client-side query validation.
- **Register data** - register info by number, register extracts, and searchable
  full text.
- **Bulk and single downloads** - weekly XML/PDF bulk archives, single
  publication PDFs, and design/trademark images.
- **Streaming variants** - every bulk and register-extract method has a `*Stream`
  variant that writes to an `io.Writer` without buffering the whole response.
- **Typed errors** - `NotFoundError`, `DataNotAvailableError`, `APIError`, and
  `XMLParseError` for `errors.As` handling.

## Installation

```bash
go get github.com/patent-dev/dpma-connect-plus
```

## Getting access

DPMAconnectPlus is a contractual data-supply service. The DPMA issues an HTTP
Basic **username** and **password** after a signed standard agreement is in place;
there is no self-service web portal. A one-time connection fee applies (200 EUR at
time of writing); data retrieval itself is free.

1. Review the service and terms on the
   [DPMAconnectPlus page](https://www.dpma.de/english/search/data_supply_services/dpmaconnect/index.html).

2. Complete the standard agreement (Standardvertrag DPMAconnectPlus) including the
   purpose-of-use paragraph, and send two signed originals to
   *Deutsches Patent- und Markenamt, Referat 2.1.2 - Kundenservice Datenabgabe,
   80297 München, Germany*.

3. The DPMA creates an account with a username and a password of your choice.

4. Export the credentials for the client and demo:

   ```bash
   export DPMA_CONNECT_PLUS_USERNAME=your-username
   export DPMA_CONNECT_PLUS_PASSWORD=your-password
   ```


## Quick start

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

## Usage

All methods accept `context.Context` for timeout and cancellation support.

### Configuration

```go
config := &dpma.Config{
    Username: "your-username",                                    // Required
    Password: "your-password",                                    // Required
    BaseURL:  "https://dpmaconnect.dpma.de/dpmaws/rest-services", // Default
    Timeout:  20 * time.Minute,                                   // Default: 20 minutes
    HTTPClient: myCustomHTTPClient,                               // Optional: your own *http.Client
}

client, err := dpma.NewClient(config)
```

If `HTTPClient` is set, `Timeout` is ignored and the custom client's timeout
applies instead.

### Common operations

```go
// Get service version
GetVersion(ctx, service string) (string, error)
```

Use the service constants: `dpma.ServicePatent`, `dpma.ServiceDesign`,
`dpma.ServiceTrademark`.

### Query validation

```go
// Validate a query against service-specific field codes (returns nil or error)
ValidatePatentQuery(query string) error
ValidateDesignQuery(query string) error
ValidateTrademarkQuery(query string) error
```

For advanced usage (tokenization, field inspection), use the `query` sub-package
directly:

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
fmt.Println(q.GetFields())    // ["TI", "INH"]
fmt.Println(q.HasField("TI")) // true

// Look up field definitions
f, ok := query.GetField("TI", query.ServicePatent)
fmt.Println(f.Description) // "title / designation"

// List all valid fields for a service
fields := query.GetValidFields(query.ServiceDesign)
```

The parser supports quoted values, comparison operators (`=`, `>=`, `<=`, `>`,
`<`), parentheses and curly braces (procedure data), bracket/brace matching, and
both English (`AND`, `OR`, `NOT`) and German (`UND`, `ODER`, `NICHT`) Boolean
operators. Field validation is per service, based on the official DPMAregister
field codes.

### Patent service

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

### Design service

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

### Trademark service

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

### Streaming downloads (memory-efficient)

Every bulk download and register extract method has a `*Stream` variant that
writes to an `io.Writer`:

```go
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

The full set of `*Stream` methods mirrors the buffered bulk and register-extract
methods above (`GetDisclosureDocumentsXMLStream`, `GetDesignImagesStream`,
`GetTrademarkRegisterExtractStream`, and so on).

### Search query syntax

All search methods use
[DPMAregister expert search syntax](https://register.dpma.de/DPMAregister/pat/experte).
The format is `FIELD=value` with Boolean operators `AND`/`UND`, `OR`/`ODER`,
`NOT`/`NICHT` (English and German are both accepted), comparison operators
(`=`, `>=`, `<=`, `>`, `<`), wildcards (`?` any chars, `!` one char, `#` one or
no char), quoted values, and parentheses. Procedure-data fields are written
inside curly braces, for example `{VST=pub-offenlegungsschrift UND VSTT=05.01.2011}`.

The `query` subpackage parses and validates queries against per-service field
definitions (see [Query validation](#query-validation)). Each service supports a
different set of field codes; the most common are listed below.

Common patent field codes:

| Code | Description |
|------|-------------|
| `TI` | Bezeichnung/Titel (title / designation) |
| `INH` | Anmelder/Inhaber (applicant / proprietor) |
| `IN` | Erfinder (inventor) |
| `IC` | IPC-Klasse (IPC classification) |
| `AKZ` | Aktenzeichen (file number / publication number) |
| `PN` | Veröffentlichungsnummer (publication number) |
| `AT` | Anmeldetag (filing date) |
| `PUB` | Publikationstag (publication date) |
| `ST` | Status |
| `AB` | Abstract |

Common design field codes:

| Code | Description |
|------|-------------|
| `TI` | Bezeichnung/Titel (title / designation) |
| `INH` | Inhaber (proprietor) |
| `ENTW` | Entwerfer (designer) |
| `ERZ` | Erzeugnis(se) (product(s)) |
| `WKL` | Warenklasse (Locarno class) |
| `RN` | Registernummer (registration number) |
| `DNR` | Designnummer (design number) |
| `AT` | Anmeldetag (filing date) |
| `ET` | Eintragungstag (registration date) |

Common trademark field codes:

| Code | Description |
|------|-------------|
| `MARKE` / `md` | Marke (trademark text) |
| `INH` | Anmelder/Inhaber (applicant / proprietor) |
| `KL` | Klasse(n) (Nice class(es)) |
| `BKL` | Bildklasse(n) (Vienna / image class(es)) |
| `MF` | Markenform (trademark form) |
| `RN` | Registernummer/Aktenzeichen (registration number / file number) |
| `AT` | Anmeldetag (filing date) |
| `ST` | Status |

For full field references, see the DPMAregister help pages:
- [Patent field codes](https://register.dpma.de/register/htdocs/prod/de/hilfe/recherchefelder/patgbm/index.html)
- [Design field codes](https://register.dpma.de/register/htdocs/prod/de/hilfe/recherchefelder/gsm/index.html)
- [Trademark field codes](https://register.dpma.de/register/htdocs/prod/de/hilfe/recherchefelder/marken/index.html)

### Date and week formatting

Publication weeks use `YYYYWW` format (6 digits); register extracts use
`YYYY-MM-DD`:

```go
pubWeek, err := dpma.FormatPublicationWeek(2024, 45) // "202445", nil
year, week, err := dpma.ParsePublicationWeek("202445") // 2024, 45, nil

date := time.Date(2024, 10, 23, 0, 0, 0, 0, time.UTC)
dateStr := dpma.FormatDate(date) // "2024-10-23"
```

### Data availability

- **Publication data** is updated weekly: patents/utility models on Thursdays,
  designs/trademarks on Fridays.
- **Register data** is updated daily.

`DataNotAvailableError` is common for future publication weeks, very old weeks
(before digital archiving), and weeks with no publications.

## Error handling

The library returns typed errors you can match with `errors.As`:

```go
// Resource not found (404)
var notFoundErr *dpma.NotFoundError
if errors.As(err, &notFoundErr) {
    fmt.Printf("not found: %s %s\n", notFoundErr.Resource, notFoundErr.ID)
}

// Data not available (common for old/future publication weeks)
var dataErr *dpma.DataNotAvailableError
if errors.As(err, &dataErr) {
    fmt.Println("data not available for the requested period")
}

// Generic API errors
var apiErr *dpma.APIError
if errors.As(err, &apiErr) {
    fmt.Printf("API error: %s (code: %s, HTTP %d)\n", apiErr.Message, apiErr.Code, apiErr.StatusCode)
}

// XML parsing failures (malformed response data)
var xmlErr *dpma.XMLParseError
if errors.As(err, &xmlErr) {
    fmt.Printf("failed to parse response in %s: %v\n", xmlErr.Operation, xmlErr.Unwrap())
}
```

## Testing

```bash
make test              # unit tests (race), mock HTTP server
make test-integration  # integration tests (//go:build integration), needs credentials
make lint
```

Integration tests hit the live DPMA API and skip unless credentials are set:

```bash
export DPMA_CONNECT_PLUS_USERNAME=your-username
export DPMA_CONNECT_PLUS_PASSWORD=your-password
make test-integration
```

An interactive demo application is included under `demo/`; it reads the same
environment variables and offers a menu for the patent, design, and trademark
services.

## Implementation status

Every endpoint of the DPMA Connect Plus API is implemented across the patent,
design, and trademark services (see the [Usage](#usage) method lists). How much
of that surface a given account can exercise depends on the permissions DPMA
grants it.

A subset is covered by the integration test suite and verified against the live
API, including `GetVersion`, the three `Search*` methods (and their `*Parsed`
variants), `GetPatentInfo` / `GetDesignInfo` / `GetTrademarkInfo`,
`GetPatentRegisterExtract`, `GetPatentPublicationPDF`,
`GetDisclosureDocumentsXML`, and `GetDesignBibliographicDataXML`.

The remaining bulk endpoints (for example `GetPublicationDataXML`,
`GetApplicantCitationsXML`, `GetDesignImages`, and the
`GetTrademarkBibData*` methods) are implemented but require permissions the test
account does not currently hold, so they are not yet exercised end to end.

## Development

Regenerate the typed client when the OpenAPI spec (`openapi.yaml`) changes:

```bash
make generate
```

This re-applies any local OpenAPI fixes and runs `go generate ./...`, which
drives [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) to rewrite
`generated/types_gen.go` and `generated/client_gen.go`. Do not edit the files
under `generated/` by hand.

Other common targets:

```bash
make fmt     # gofmt
make lint    # golangci-lint
make tidy    # go mod tidy
```

## Related projects

Part of the [patent.dev](https://patent.dev) open-source patent data ecosystem:

- [epo-ops](https://github.com/patent-dev/epo-ops) - EPO Open Patent Services client (bibliographic, full text, families, legal status, images)
- [epo-bdds](https://github.com/patent-dev/epo-bdds) - EPO Bulk Data Distribution Service client (DOCDB, INPADOC, EP full text)
- [uspto-odp](https://github.com/patent-dev/uspto-odp) - USPTO Open Data Portal client (patents, PTAB, TSDR, full text)

The [bulk-file-loader](https://github.com/patent-dev/bulk-file-loader) uses these libraries for automated patent data downloads.

## License

MIT - Funktionslust GmbH / patent.dev.
