# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **AIMC Package**: Supplement system for local fund-to-company overrides
  - `Supplement` and `SupplementFundInfo` structs for local overrides
  - `SaveSupplementEntry()` - Add or update supplement entries
  - `DeleteSupplementEntry()` - Remove supplement entries
  - `HasSupplement()` - Check if fund has supplement override
  - `GetSupplement()` - Retrieve supplement entry
  - `GetAllCompanies()` - List all companies including supplement funds
  - `GetCategoryIDByName()` - Reverse lookup category by name
  - Automatic merge of AIMC + supplement data (supplement takes precedence)
  - Graceful degradation when supplement file is missing

### Changed
- All AIMC lookup methods now merge supplement data automatically
- `GetFundsByCompany()` and `GetFundsByCompanyFuzzy()` include supplement funds

## [1.0.0] - 2024-03-06

### Added
- Initial release
- **AIMC Package**: Association of Investment Management Companies fund data
  - Auto-update from AIMC website (quarterly, 90-day check)
  - XLSX parsing with `github.com/thedatashed/xlsxreader`
  - Fund category lookup
  - Company/fund search
  - Fuzzy matching for company names
  
- **SET Package**: Stock Exchange of Thailand company data
  - Auto-update from SET.or.th (monthly, 30-day check)
  - HTML table parsing
  - TIS-620 to UTF-8 encoding conversion
  - Thai-to-English company name translation
  - Sector and industry translation
  - Custom translation support
  - Fuzzy name matching

### Features
- Automatic data updates with `FetchAndUpdate()`
- HTTP retry with exponential backoff
- Background update support
- Zero setup required
- Comprehensive test coverage for both packages
- Complete API documentation (621 lines)

[Unreleased]: https://github.com/jwitmann/thai-market-data/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/jwitmann/thai-market-data/releases/tag/v1.0.0
