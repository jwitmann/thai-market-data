# SET data (Thai Stock Exchange)

Goal is to have some API allowing us to get the SET data of a given company name in english and thai from either it's symbol or company name. It should allow us to translate a thai name into its english equivalent.

## Installation

```bash
go get github.com/jwitmann/thai-market-data/set
```

## Usage

```go
import "github.com/jwitmann/thai-market-data/set"

// Create client
client, err := set.NewClient(dataDir)
if err != nil {
    log.Fatal(err)
}

// Get company by ticker symbol
company, err := client.GetBySymbol("PTT")
// Returns: {NameEN: "PTT PUBLIC COMPANY LIMITED", NameTH: "บริษัท ปตท. จำกัด (มหาชน)", ...}

// Search by name (fuzzy matching)
company, err := client.GetByName("PTT PUBLIC COMPANY LIMITED")
company, err := client.GetByName("บริษัท ปตท. จำกัด (มหาชน)")

// Check if a name contains Thai characters
isThai := client.IsThaiName("บริษัท ปตท.")

// Translate Thai company name to English
englishName := client.TranslateName("หุ้นสามัญของบริษัท ปตท. จำกัด (มหาชน)")
// Returns: "PTT PUBLIC COMPANY LIMITED"
// (Strips "หุ้นสามัญของ" prefix and looks up company)

// Translate Thai sector name to English
sectorEN := client.TranslateSector("พลังงานและสาธารณูปโภค")
// Returns: "Energy & Utilities"

// Translate Thai industry name to English
industryEN := client.TranslateIndustry("ทรัพยากร")
// Returns: "Resources"
```

## Auto-Update

The client supports automatic updates from the SET website:

```go
// Check if data needs update (older than 30 days)
if client.NeedsUpdate() {
    // Update in background
    go func() {
        if err := client.FetchAndUpdate(); err != nil {
            log.Printf("SET update failed: %v", err)
        }
    }()
}
```

## Architecture

### Data Flow

1. **Data Loading**:
   - Downloads HTML tables from SET.or.th (English and Thai versions)
   - Parses and converts TIS-620 encoding to UTF-8
   - Merges data by symbol (ticker)
   - Creates unique IDs for industries and sectors
   - Outputs `${DATA_DIR}/SET_mappings.json`

2. **JSON Structure**:

```json
{
  "companies": {
    "PTT": {
      "name_en": "PTT PUBLIC COMPANY LIMITED",
      "name_th": "บริษัท ปตท. จำกัด (มหาชน)",
      "market": "SET",
      "industry_id": "ind_007",
      "sector_id": "sec_005"
    }
  },
  "industries": {
    "ind_007": {
      "name_th": "ทรัพยากร",
      "name_en": "Resources"
    }
  },
  "sectors": {
    "sec_005": {
      "name_th": "พลังงานและสาธารณูปโภค",
      "name_en": "Energy & Utilities"
    }
  }
}
```

3. **Runtime API** (`set/client.go`):
   - Loads JSON file on startup
   - Provides lookup and translation functions
   - Maintains in-memory cache for fast access

### Company Structure

```go
type Company struct {
    NameEN     string // English company name
    NameTH     string // Thai company name
    Market     string // "SET" or "mai"
    IndustryID string // Reference to industries map
    SectorID   string // Reference to sectors map
}

type Industry struct {
    NameTH string // Thai industry name
    NameEN string // English industry name
}

type Sector struct {
    NameTH string // Thai sector name
    NameEN string // English sector name
}
```

## Data Sources

The SET client downloads data directly from SET.or.th. The HTML tables contain the following fields:

**English version (`SET_listedCompanies_en.csv`):**

- Symbol, Company, Market, Industry, Sector, Address, Zip code, Tel., Fax, Website

**Thai version (`SET_listedCompanies_th.csv`):**

- หลักทรัพย์, บริษัท, ตลาด, กลุ่มอุตสาหกรรม, หมวดธุรกิจ, ที่อยู่, รหัสไปรษณีย์, โทรศัพท์, โทรสาร, เว๊บไซต์

## Fuzzy Matching Algorithm

The `GetByName` function uses aggressive fuzzy matching:

1. **Exact match** on uppercase name
2. **Case-insensitive exact match**
3. **Substring match** (contains) - search query contained in company name
4. **Progressive word removal** for Thai names:
   - "บริษัท ปตท. จำกัด (มหาชน)" → not found
   - "บริษัท ปตท. จำกัด" → not found
   - "บริษัท ปตท." → not found
   - "บริษัท" → found (matches any company with "บริษัท")

## Examples

### Lookup by Symbol

```go
company, _ := client.GetBySymbol("KBANK")
fmt.Println(company.NameEN)  // "KASIKORNBANK PUBLIC COMPANY LIMITED"
fmt.Println(company.NameTH)  // "ธนาคารกสิกรไทย จำกัด (มหาชน)"
```

### Fuzzy Name Search

```go
// Thai name with prefix
company, _ := client.GetByName("หุ้นสามัญของบริษัท ปตท. จำกัด (มหาชน)")
// Works because fuzzy matching finds "บริษัท ปตท. จำกัด (มหาชน)"

// Partial match
company, _ := client.GetByName("KASIKORNBANK")
// Works because "KASIKORNBANK" is contained in "KASIKORNBANK PUBLIC COMPANY LIMITED"
```

### Portfolio Translation

```go
// Translate portfolio holdings
for i, holding := range portfolio.TopHoldings {
    if translated := client.TranslateName(holding.Name); translated != "" {
        portfolio.TopHoldings[i].Name = translated
    }
}
// Each Thai company name is replaced with its English equivalent
```
