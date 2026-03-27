# AIMC Fund Categories

## What is AIMC?

**AIMC** = Association of Investment Management Companies (Thailand) — the self-regulatory organization for Thai asset management companies.

AIMC classifies Thai mutual funds into standardized categories based on:

- Asset class (Equity, Fixed Income, Mixed, Money Market, etc.)
- Investment style/strategy
- Geographic focus

## Categories

Example categories:

- Equity Fund - Large Cap
- Equity Fund -中小盘 (China A-Shares)
- Equity Fund -中小型股 (Small Cap)
- Fixed Income Fund
- Mixed Fund
- Money Market Fund

## Use Cases

This package provides AIMC data for applications working with Thai mutual funds:

1. **English fund names** - Complement Finnomena API (which returns Thai names) with English translations
2. **Thai fund names** - Access both Thai and English versions of fund names
3. **Asset manager names** - Get firm names in English
4. **Fund categories** - Standardized classification (e.g., "Equity Fund - Large Cap")
5. **Category search** - Browse funds by AIMC category 

## Data Source

The mappings are generated from:

- **AIMC category Excel**: <https://association.aimc.or.th/wp-content/uploads/2025/Aimccategory/AIMC-Category-Q3-2025.xlsx>
- **Finnomena API**: Used to link fund codes with AIMC category IDs

Data is automatically downloaded and updated via `FetchAndUpdate()`.

## File Location

- **Generated mappings**: `${DATA_DIR}/aimc_mappings.json`
- **Source CSV**: `data/aimc_data.csv` (in repository)
- **Supplement overrides**: `${DATA_DIR}/company_supplement.json` (optional)

## Supplement System

The supplement system allows you to add or override fund-to-company mappings locally without modifying the AIMC data.

### Use Cases

- **New funds** not yet in AIMC quarterly data
- **Missing company mappings** for funds that exist in AIMC but lack firm information
- **Custom categorizations** for special cases

### How It Works

1. Create a `company_supplement.json` file in your data directory
2. The client automatically loads and merges supplement data with AIMC data
3. **Supplement takes precedence** - overrides any conflicting AIMC entries

### Supplement File Format

```json
{
  "TDEFENSE": {
    "fund_code": "TDEFENSE",
    "company": "TISCO Asset Management Co., Ltd.",
    "category": "Equity Fund - Sector",
    "source": "manual"
  },
  "TGOLD-UH": {
    "fund_code": "TGOLD-UH",
    "company": "TISCO Asset Management Co., Ltd.",
    "category": "Equity Fund - Sector",
    "source": "auto-detected"
  }
}
```

### Programmatic Management

```go
// Add or update a supplement entry
client.SaveSupplementEntry("NEWFUND", "Company Name", "Category Name")

// Delete a supplement entry
client.DeleteSupplementEntry("NEWFUND")

// Check if fund has supplement override
if client.HasSupplement("NEWFUND") {
    info := client.GetSupplement("NEWFUND")
    fmt.Println(info.Company)
}

// Get all companies (including supplement funds)
companies := client.GetAllCompanies()
```

### Company Name Matching

When adding supplements, use the **exact company name** as it appears in AIMC data:

- ✅ `TISCO Asset Management Co., Ltd.`
- ❌ `TISCO` (too short)
- ❌ `TISCO Asset Management` (missing "Co., Ltd.")

The client provides `GetAllCompanies()` to list all available company names for reference.
