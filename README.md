# thai-market-data

[![Go Reference](https://pkg.go.dev/badge/github.com/jwitmann/thai-market-data.svg)](https://pkg.go.dev/github.com/jwitmann/thai-market-data)
[![Go Report Card](https://goreportcard.com/badge/github.com/jwitmann/thai-market-data?t=1)](https://goreportcard.com/report/github.com/jwitmann/thai-market-data)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

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

## API Reference

See [API.md](API.md) for complete API documentation with all available methods.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure your code passes all tests:
```bash
go test ./...
```

## License

MIT License - see [LICENSE](LICENSE) file for details
