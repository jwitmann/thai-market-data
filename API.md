# API Documentation

Complete API reference for the thai-market-data package.

## Table of Contents

- [AIMC Package](#aimc-package)
- [SET Package](#set-package)

---

## AIMC Package

Package `aimc` provides access to AIMC (Association of Investment Management Companies Thailand) fund category data.

### Types

```go
type Client struct {
    // AIMC client - create with NewClient()
}

type Mappings struct {
    Categories map[string]string   // category_id -> category_name
    Funds      map[string]FundInfo // fund_code -> fund_info
}

type FundInfo struct {
    LegalName      string // Legal/fund name in English
    ThaiName       string // Fund name in Thai
    FirmName       string // Asset management company name
    AIMCCategoryID string // AIMC category identifier
    AIMCCategory   string // Human-readable category name
}

type Supplement struct {
    Funds map[string]SupplementFundInfo // fund_code -> supplement_info
}

type SupplementFundInfo struct {
    FundCode string // Fund code (e.g., "TDEFENSE")
    Company  string // Asset management company name
    Category string // AIMC category name
    Source   string // Source of entry (e.g., "manual", "auto-detected")
}
```

### Constructor

#### NewClient

```go
func NewClient(dataDir string) (*Client, error)
```

Creates a new AIMC client.

**Parameters:**
- `dataDir`: Directory containing `aimc_mappings.json`

**Returns:**
- `*Client`: Configured client
- `error`: If mappings cannot be loaded

**Example:**
```go
client, err := aimc.NewClient("/path/to/data")
if err != nil {
    log.Fatal(err)
}
```

---

### Lookup Methods

#### GetCategoryName

```go
func (c *Client) GetCategoryName(categoryID string) string
```

Returns the human-readable name for a category ID.

**Example:**
```go
category := client.GetCategoryName("LC00002470")
// Returns: "Equity Fund - Large Cap"
```

#### GetFundInfo

```go
func (c *Client) GetFundInfo(fundCode string) (legalName, thaiName, firmName, category string)
```

Returns comprehensive information about a fund.

**Example:**
```go
legalName, thaiName, firmName, category := client.GetFundInfo("F000001")
fmt.Printf("Fund: %s (%s)\n", legalName, thaiName)
fmt.Printf("Company: %s\n", firmName)
fmt.Printf("Category: %s\n", category)
```

#### GetCategories

```go
func (c *Client) GetCategories() []string
```

Returns all available category names.

**Example:**
```go
categories := client.GetCategories()
for _, cat := range categories {
    fmt.Println(cat)
}
```

#### GetFundsByCategory

```go
func (c *Client) GetFundsByCategory(categoryName string) []string
```

Returns all fund codes in a specific category.

**Example:**
```go
funds := client.GetFundsByCategory("Equity Fund - Large Cap")
fmt.Printf("Found %d funds in category\n", len(funds))
```

#### GetFundsByCompany

```go
func (c *Client) GetFundsByCompany(companyName string) []string
```

Returns all funds managed by a specific company (exact match).

**Example:**
```go
funds := client.GetFundsByCompany("KASIKORN ASSET MANAGEMENT")
```

#### GetFundsByCompanyFuzzy

```go
func (c *Client) GetFundsByCompanyFuzzy(partialName string) []string
```

Returns all funds where company name contains the partial name (case-insensitive).

**Example:**
```go
funds := client.GetFundsByCompanyFuzzy("KASIKORN")
// Matches: "KASIKORN ASSET MANAGEMENT", "KASIKORN SECURITIES", etc.
```

#### GetAllFunds

```go
func (c *Client) GetAllFunds() []string
```

Returns all fund codes.

**Example:**
```go
allFunds := client.GetAllFunds()
fmt.Printf("Total funds: %d\n", len(allFunds))
```

#### GetMappings

```go
func (c *Client) GetMappings() *Mappings
```

Returns the complete mappings structure.

**Example:**
```go
mappings := client.GetMappings()
for fundCode, info := range mappings.Funds {
    fmt.Printf("%s: %s\n", fundCode, info.LegalName)
}
```

---

### Update Methods

#### NeedsUpdate

```go
func (c *Client) NeedsUpdate() bool
```

Checks if data is older than 90 days (quarterly check).

**Returns:**
- `true`: Data needs update
- `false`: Data is current

**Example:**
```go
if client.NeedsUpdate() {
    // Trigger update
}
```

#### FetchAndUpdate

```go
func (c *Client) FetchAndUpdate() error
```

Downloads fresh data from AIMC website and updates local mappings.

**Process:**
1. Detects latest quarterly Excel file URL
2. Downloads and parses XLSX
3. Merges with existing data
4. Saves updated mappings

**Example:**
```go
if client.NeedsUpdate() {
    log.Println("Updating AIMC data...")
    if err := client.FetchAndUpdate(); err != nil {
        log.Printf("Update failed: %v", err)
    }
}
```

#### GetAllCompanies

```go
func (c *Client) GetAllCompanies() []string
```

Returns a sorted list of all unique company names from both AIMC data and supplement entries.

**Example:**
```go
companies := client.GetAllCompanies()
for _, company := range companies {
    fmt.Println(company)
}
// Output: "KASIKORN ASSET MANAGEMENT Co., Ltd."
//         "TISCO Asset Management Co., Ltd."
//         ...
```

---

### Supplement Methods

#### SaveSupplementEntry

```go
func (c *Client) SaveSupplementEntry(fundCode, company, category string) error
```

Adds or updates a supplement entry for a fund. The supplement takes precedence over AIMC data.

**Parameters:**
- `fundCode`: Fund code (e.g., "TDEFENSE")
- `company`: Exact company name as it appears in AIMC data
- `category`: Category name (e.g., "Equity Fund - Sector")

**Returns:**
- `error`: If supplement cannot be saved

**Example:**
```go
err := client.SaveSupplementEntry("TDEFENSE", 
    "TISCO Asset Management Co., Ltd.", 
    "Equity Fund - Sector")
if err != nil {
    log.Fatal(err)
}
```

#### DeleteSupplementEntry

```go
func (c *Client) DeleteSupplementEntry(fundCode string) error
```

Removes a supplement entry for a fund.

**Example:**
```go
err := client.DeleteSupplementEntry("TDEFENSE")
```

#### HasSupplement

```go
func (c *Client) HasSupplement(fundCode string) bool
```

Checks if a fund has a supplement override.

**Example:**
```go
if client.HasSupplement("TDEFENSE") {
    fmt.Println("Fund has supplement override")
}
```

#### GetSupplement

```go
func (c *Client) GetSupplement(fundCode string) (*SupplementFundInfo, bool)
```

Retrieves the supplement entry for a fund.

**Returns:**
- `*SupplementFundInfo`: Supplement data
- `bool`: Whether supplement exists

**Example:**
```go
if info, ok := client.GetSupplement("TDEFENSE"); ok {
    fmt.Printf("Company: %s\n", info.Company)
    fmt.Printf("Category: %s\n", info.Category)
}
```

#### GetCategoryIDByName

```go
func (c *Client) GetCategoryIDByName(categoryName string) (string, bool)
```

Performs a reverse lookup to find the category ID from a category name.

**Example:**
```go
if categoryID, ok := client.GetCategoryIDByName("Equity Fund - Large Cap"); ok {
    fmt.Printf("Category ID: %s\n", categoryID)
}
```

---

## SET Package

Package `set` provides Thai Stock Exchange company name translation and lookup services.

### Types

```go
type Client struct {
    // SET client - create with NewClient()
}

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

type SETData struct {
    Companies map[string]Company  // symbol -> company
    Industries map[string]Industry // id -> industry
    Sectors map[string]Sector      // id -> sector
}

type CustomTranslations struct {
    Version      int                    // Translation version
    Sectors      map[string]Translation // Custom sector translations
    Industries   map[string]Translation // Custom industry translations
    Companies    map[string]Translation // Custom company translations
    Untranslated []UntranslatedEntry    // Log of untranslated items
}

type Translation struct {
    EN       string // English translation
    Verified bool   // Whether translation is verified
    Date     string // ISO date string
}

type UntranslatedEntry struct {
    Text   string // Untranslated text
    Type   string // Type: "sector", "industry", "company"
    Date   string // ISO date string
    FundID string // Associated fund ID (optional)
}
```

### Constructor

#### NewClient

```go
func NewClient(dataDir string) (*Client, error)
```

Creates a new SET client.

**Parameters:**
- `dataDir`: Directory containing `SET_mappings.json`

**Returns:**
- `*Client`: Configured client
- `error`: If data cannot be loaded

**Example:**
```go
client, err := set.NewClient("/path/to/data")
if err != nil {
    log.Fatal(err)
}
```

---

### Lookup Methods

#### GetBySymbol

```go
func (c *Client) GetBySymbol(symbol string) (*Company, error)
```

Returns company information by ticker symbol.

**Example:**
```go
company, err := client.GetBySymbol("PTT")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("%s / %s\n", company.NameEN, company.NameTH)
// Output: PTT PUBLIC COMPANY LIMITED / บริษัท ปตท. จำกัด (มหาชน)
```

#### GetByName

```go
func (c *Client) GetByName(name string) (*Company, error)
```

Finds company by name using fuzzy matching. Accepts partial matches and handles Thai names with prefixes.

**Fuzzy Matching Algorithm:**
1. Exact match (uppercase)
2. Case-insensitive exact match
3. Substring match
4. Progressive word removal for Thai names

**Example:**
```go
// By English name
company, _ := client.GetByName("PTT PUBLIC COMPANY LIMITED")

// By Thai name (with prefix)
company, _ := client.GetByName("หุ้นสามัญของบริษัท ปตท. จำกัด (มหาชน)")
// Strips "หุ้นสามัญของ" prefix automatically

// Partial match
company, _ := client.GetByName("KASIKORNBANK")
// Matches: "KASIKORNBANK PUBLIC COMPANY LIMITED"
```

#### IsThaiName

```go
func (c *Client) IsThaiName(name string) bool
```

Checks if a string contains Thai characters.

**Example:**
```go
if client.IsThaiName("บริษัท ปตท.") {
    // Handle Thai text
}
```

---

### Translation Methods

#### TranslateName

```go
func (c *Client) TranslateName(thaiName string) string
```

Translates a Thai company name to English. Handles common prefixes like "หุ้นสามัญของ" (ordinary shares of).

**Example:**
```go
english := client.TranslateName("หุ้นสามัญของบริษัท ปตท. จำกัด (มหาชน)")
// Returns: "PTT PUBLIC COMPANY LIMITED"

// Also works without prefix
english := client.TranslateName("บริษัท ปตท. จำกัด (มหาชน)")
// Returns: "PTT PUBLIC COMPANY LIMITED"
```

#### TranslateIndustry

```go
func (c *Client) TranslateIndustry(thaiIndustry string) string
```

Translates Thai industry name to English.

**Example:**
```go
english := client.TranslateIndustry("ทรัพยากร")
// Returns: "Resources"
```

#### TranslateSector

```go
func (c *Client) TranslateSector(thaiSector string) string
```

Translates Thai sector name to English.

**Example:**
```go
english := client.TranslateSector("พลังงานและสาธารณูปโภค")
// Returns: "Energy & Utilities"
```

#### TranslateWithFallback

```go
func (c *Client) TranslateWithFallback(thaiText, transType, fundID string) string
```

Translates with multiple fallback strategies:
1. Check custom translations
2. Check SET data
3. Log as untranslated
4. Return original if all fail

**Parameters:**
- `thaiText`: Text to translate
- `transType`: Type - "sector", "industry", or "company"
- `fundID`: Associated fund ID for logging

**Example:**
```go
// Translate with full fallback chain
english := client.TranslateWithFallback(
    "บริการด้านการเงิน",
    "sector",
    "F000001",
)
```

---

### Data Access

#### GetSETData

```go
func (c *Client) GetSETData() *SETData
```

Returns the complete SET data structure.

**Example:**
```go
data := client.GetSETData()
for symbol, company := range data.Companies {
    fmt.Printf("%s: %s\n", symbol, company.NameEN)
}
```

---

### Custom Translations

#### GetCustomTranslation

```go
func (c *Client) GetCustomTranslation(thaiText, transType string) (string, bool)
```

Retrieves a custom translation if it exists.

**Returns:**
- `translation`: The English translation
- `exists`: Whether translation was found

**Example:**
```go
if trans, ok := client.GetCustomTranslation("บริษัทใหม่", "company"); ok {
    fmt.Println("Found custom translation:", trans)
}
```

#### SetCustomTranslation

```go
func (c *Client) SetCustomTranslation(thaiText, englishText, transType string, verified bool)
```

Adds or updates a custom translation.

**Example:**
```go
client.SetCustomTranslation(
    "บริษัทใหม่",
    "NEW COMPANY LIMITED",
    "company",
    true, // verified
)
```

#### LogUntranslated

```go
func (c *Client) LogUntranslated(thaiText, transType, fundID string)
```

Logs an untranslated item for later review.

**Example:**
```go
client.LogUntranslated("ข้อความใหม่", "sector", "F000001")
```

---

### Update Methods

#### NeedsUpdate

```go
func (c *Client) NeedsUpdate() bool
```

Checks if data is older than 30 days (monthly check).

**Returns:**
- `true`: Data needs update
- `false`: Data is current

**Example:**
```go
if client.NeedsUpdate() {
    // Trigger update
}
```

#### FetchAndUpdate

```go
func (c *Client) FetchAndUpdate() error
```

Downloads fresh data from SET website and updates local mappings.

**Process:**
1. Downloads English and Thai HTML tables from SET.or.th
2. Converts TIS-620 encoding to UTF-8
3. Parses HTML tables
4. Merges English and Thai data by symbol
5. Saves updated mappings
6. Preserves custom translations

**Example:**
```go
if client.NeedsUpdate() {
    log.Println("Updating SET data...")
    if err := client.FetchAndUpdate(); err != nil {
        log.Printf("Update failed: %v", err)
    }
}
```

---

## Error Handling

Both packages return errors for:
- Missing or corrupted data files
- Network failures during updates
- Invalid parameters

**Best Practice:**
```go
client, err := aimc.NewClient(dataDir)
if err != nil {
    // Handle initialization error
    log.Fatal(err)
}

// Updates fail gracefully - existing data is preserved
if err := client.FetchAndUpdate(); err != nil {
    log.Printf("Update failed, using existing data: %v", err)
}
```

---

## Thread Safety

- **AIMC Client**: Safe for concurrent read operations
- **SET Client**: Uses RWMutex for safe concurrent access
- **Updates**: Should be triggered from a single goroutine

**Recommended Pattern:**
```go
// Background update goroutine
go func() {
    for {
        if client.NeedsUpdate() {
            client.FetchAndUpdate()
        }
        time.Sleep(24 * time.Hour)
    }
}()
```
