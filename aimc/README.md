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

## Why ThaiFA Uses AIMC Data

This package enriches fund data from Finnomena with AIMC mappings to provide:

1. **English fund names** - The Finnomena API returns Thai names; AIMC provides English translations
2. **Thai fund names** - Stored from AIMC data for display toggle via `UseEnglishNames` setting
3. **Asset manager names** - Firm names in English
4. **Fund categories** - Standardized classification (e.g., "Equity Fund - Large Cap")
5. **Category search** - Browse funds by AIMC category 

## Data Source

The mappings are generated from:

- **AIMC category Excel**: <https://association.aimc.or.th/wp-content/uploads/2025/Aimccategory/AIMC-Category-Q3-2025.xlsx>
- **Finnomena API**: Used to link fund codes with AIMC category IDs

## Generation

Generate mappings by running:
```bash
go run ./cmd/import_aimc/main.go -output ${DATA_DIR}/aimc_mappings.json
```

```bash
./bin/import_aimc
```

The tool fetches fund data from Finnomena API and merges with the AIMC category list to produce the mapping file.

## File Location

- **Generated mappings**: `${DATA_DIR}/aimc_mappings.json`
- **Source CSV**: `data/aimc_data.csv` (in repository)
