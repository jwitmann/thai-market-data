# thai-market-data

Go package for Thai financial market data sources.

## Packages

- **[aimc](aimc/)** - Association of Investment Management Companies fund categories
  - Auto-updating fund category mappings from AIMC website
  - XLSX parsing for quarterly data updates
  
- **[set](set/)** - Stock Exchange of Thailand company translations
  - Thai-to-English company name translation
  - Auto-updating from SET.or.th (HTML table parsing)
  - TIS-620 to UTF-8 encoding conversion

## Installation

```bash
go get github.com/jwitmann/thai-market-data
```

## Quick Start

### AIMC

```go
import "github.com/jwitmann/thai-market-data/aimc"

// Create client
client, err := aimc.NewClient(dataDir)
if err != nil {
    log.Fatal(err)
}

// Auto-update if data is older than 90 days
if client.NeedsUpdate() {
    go client.FetchAndUpdate() // Background update
}

// Lookup fund category
info := client.GetFundInfo("F000001")
fmt.Println(info.AIMCCategory) // "Equity Fund - Large Cap"
```

### SET

```go
import "github.com/jwitmann/thai-market-data/set"

// Create client
client, err := set.NewClient(dataDir)
if err != nil {
    log.Fatal(err)
}

// Translate Thai company name to English
englishName := client.TranslateName("บริษัท ปตท. จำกัด (มหาชน)")
// Returns: "PTT PUBLIC COMPANY LIMITED"

// Get company by symbol
company, _ := client.GetBySymbol("PTT")
fmt.Println(company.NameEN)
```

## Features

- **Auto-updating**: Both packages check for stale data and update automatically
- **Quarterly updates (AIMC)**: Checks every 90 days
- **Monthly updates (SET)**: Checks every 30 days
- **HTTP retry**: Exponential backoff for network resilience
- **Zero external setup**: Works out of the box with embedded data

## Dependencies

- `github.com/thedatashed/xlsxreader` - AIMC XLSX parsing
- `github.com/tupkung/tis620` - SET TIS-620 encoding conversion

## Documentation

- [AIMC Package Documentation](aimc/README.md)
- [SET Package Documentation](set/README.md)

## License

MIT
