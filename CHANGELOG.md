# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[1.0.0]: https://github.com/jwitmann/thai-market-data/releases/tag/v1.0.0
