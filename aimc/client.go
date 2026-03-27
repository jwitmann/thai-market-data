package aimc

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thedatashed/xlsxreader"
)

// Client provides AIMC (Association of Investment Management Companies) data lookup
type Client struct {
	dataDir        string
	mappings       *Mappings
	mappingsPath   string
	supplement     *Supplement
	supplementPath string
}

// Supplement stores user-defined overrides for AIMC mappings
// This allows local supplementation of fund-to-company and fund-to-category mappings
type Supplement struct {
	Categories map[string]string             `json:"categories"` // category_id -> category_name
	Funds      map[string]SupplementFundInfo `json:"funds"`      // fund_code -> fund info
}

// SupplementFundInfo represents supplement data for a single fund
type SupplementFundInfo struct {
	LegalName      string `json:"legal_name,omitempty"`
	ThaiName       string `json:"thai_name,omitempty"`
	FirmName       string `json:"firm_name"`                  // Required: company name
	AIMCCategoryID string `json:"aimc_category_id,omitempty"` // Optional: category ID
}

// Mappings stores AIMC category and fund information
type Mappings struct {
	Categories map[string]string   `json:"categories"`
	Funds      map[string]FundInfo `json:"funds"`
}

// FundInfo represents AIMC data for a single fund
type FundInfo struct {
	LegalName      string `json:"legal_name"`
	ThaiName       string `json:"thai_name"`
	FirmName       string `json:"firm_name"`
	AIMCCategoryID string `json:"aimc_category_id,omitempty"`
	AIMCCategory   string `json:"-"` // Temporary field for parsing, not serialized
}

// NewClient creates a new AIMC client with the given data directory
func NewClient(dataDir string) (*Client, error) {
	client := &Client{
		dataDir: dataDir,
	}

	if err := client.loadMappings(); err != nil {
		return nil, fmt.Errorf("failed to load AIMC mappings: %w", err)
	}

	// Load supplement (optional - don't fail if missing)
	if err := client.loadSupplement(); err != nil {
		log.Printf("[AIMC] No supplement file found or failed to load: %v", err)
		client.supplement = &Supplement{
			Categories: make(map[string]string),
			Funds:      make(map[string]SupplementFundInfo),
		}
	}

	return client, nil
}

// loadMappings loads AIMC data from JSON file
func (c *Client) loadMappings() error {
	c.mappingsPath = filepath.Join(c.dataDir, "aimc_mappings.json")
	data, err := os.ReadFile(c.mappingsPath)
	if err != nil {
		return fmt.Errorf("failed to read aimc_mappings.json: %w", err)
	}

	var mappings Mappings
	if err := json.Unmarshal(data, &mappings); err != nil {
		return fmt.Errorf("failed to parse aimc_mappings.json: %w", err)
	}

	c.mappings = &mappings
	return nil
}

// loadSupplement loads supplement data from JSON file (optional)
func (c *Client) loadSupplement() error {
	c.supplementPath = filepath.Join(c.dataDir, "company_supplement.json")
	data, err := os.ReadFile(c.supplementPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("supplement file not found: %w", err)
		}
		return fmt.Errorf("failed to read company_supplement.json: %w", err)
	}

	var supplement Supplement
	if err := json.Unmarshal(data, &supplement); err != nil {
		return fmt.Errorf("failed to parse company_supplement.json: %w", err)
	}

	// Initialize maps if nil
	if supplement.Categories == nil {
		supplement.Categories = make(map[string]string)
	}
	if supplement.Funds == nil {
		supplement.Funds = make(map[string]SupplementFundInfo)
	}

	c.supplement = &supplement
	log.Printf("[AIMC] Loaded supplement with %d funds and %d categories",
		len(supplement.Funds), len(supplement.Categories))
	return nil
}

// GetCategoryName returns the category name for a given category ID
// Checks supplement first, then falls back to AIMC mappings
func (c *Client) GetCategoryName(categoryID string) string {
	// Check supplement first
	if c.supplement != nil {
		if name, ok := c.supplement.Categories[categoryID]; ok {
			return name
		}
	}
	// Fall back to AIMC mappings
	if c.mappings == nil {
		return ""
	}
	return c.mappings.Categories[categoryID]
}

// GetFundInfo returns legal name, thai name, firm name, and category for a fund code
// Checks supplement first, then falls back to AIMC mappings
func (c *Client) GetFundInfo(fundCode string) (legalName, thaiName, firmName, category string) {
	// Check supplement first
	if c.supplement != nil {
		if info, ok := c.supplement.Funds[fundCode]; ok {
			legalName = info.LegalName
			thaiName = info.ThaiName
			firmName = info.FirmName
			category = c.GetCategoryName(info.AIMCCategoryID)
			return
		}
	}

	// Fall back to AIMC mappings
	if c.mappings == nil {
		return "", "", "", ""
	}

	info, ok := c.mappings.Funds[fundCode]
	if !ok {
		return "", "", "", ""
	}

	legalName = info.LegalName
	thaiName = info.ThaiName
	firmName = info.FirmName
	category = c.mappings.Categories[info.AIMCCategoryID]
	return
}

// GetCategories returns all available category names from both AIMC and supplement
func (c *Client) GetCategories() []string {
	seen := make(map[string]bool)
	var categories []string

	// Add supplement categories first
	if c.supplement != nil {
		for _, name := range c.supplement.Categories {
			if !seen[name] {
				seen[name] = true
				categories = append(categories, name)
			}
		}
	}

	// Add AIMC categories
	if c.mappings != nil {
		for _, name := range c.mappings.Categories {
			if !seen[name] {
				seen[name] = true
				categories = append(categories, name)
			}
		}
	}

	return categories
}

// GetFundsByCategory returns all fund codes in a given category
// Merges results from both AIMC and supplement
func (c *Client) GetFundsByCategory(categoryName string) []string {
	seen := make(map[string]bool)
	var funds []string

	// Check supplement first
	if c.supplement != nil {
		for id, name := range c.supplement.Categories {
			if name == categoryName {
				// Find all funds in this category from supplement
				for code, info := range c.supplement.Funds {
					if info.AIMCCategoryID == id && !seen[code] {
						seen[code] = true
						funds = append(funds, code)
					}
				}
				break
			}
		}
	}

	// Check AIMC mappings
	if c.mappings != nil {
		// Find category ID from name
		var categoryID string
		for id, name := range c.mappings.Categories {
			if name == categoryName {
				categoryID = id
				break
			}
		}

		// Find all funds in this category
		if categoryID != "" {
			for code, info := range c.mappings.Funds {
				if info.AIMCCategoryID == categoryID && !seen[code] {
					seen[code] = true
					funds = append(funds, code)
				}
			}
		}
	}

	return funds
}

// GetFundsByCompany returns all fund codes for a given company/firm
// Merges results from both AIMC and supplement
func (c *Client) GetFundsByCompany(companyName string) []string {
	seen := make(map[string]bool)
	var funds []string
	companyUpper := strings.ToUpper(companyName)

	// Check supplement first
	if c.supplement != nil {
		for code, info := range c.supplement.Funds {
			if strings.ToUpper(info.FirmName) == companyUpper && !seen[code] {
				seen[code] = true
				funds = append(funds, code)
			}
		}
	}

	// Check AIMC mappings
	if c.mappings != nil {
		for code, info := range c.mappings.Funds {
			if strings.ToUpper(info.FirmName) == companyUpper && !seen[code] {
				seen[code] = true
				funds = append(funds, code)
			}
		}
	}

	return funds
}

// GetFundsByCompanyFuzzy returns fund codes for partial company name matches
// Merges results from both AIMC and supplement
func (c *Client) GetFundsByCompanyFuzzy(partialName string) []string {
	seen := make(map[string]bool)
	var funds []string
	partialUpper := strings.ToUpper(partialName)

	// Check supplement first
	if c.supplement != nil {
		for code, info := range c.supplement.Funds {
			if strings.Contains(strings.ToUpper(info.FirmName), partialUpper) && !seen[code] {
				seen[code] = true
				funds = append(funds, code)
			}
		}
	}

	// Check AIMC mappings
	if c.mappings != nil {
		for code, info := range c.mappings.Funds {
			if strings.Contains(strings.ToUpper(info.FirmName), partialUpper) && !seen[code] {
				seen[code] = true
				funds = append(funds, code)
			}
		}
	}

	return funds
}

// GetAllFunds returns all fund codes from both AIMC and supplement
func (c *Client) GetAllFunds() []string {
	seen := make(map[string]bool)
	var funds []string

	// Add supplement funds first
	if c.supplement != nil {
		for code := range c.supplement.Funds {
			if !seen[code] {
				seen[code] = true
				funds = append(funds, code)
			}
		}
	}

	// Add AIMC funds
	if c.mappings != nil {
		for code := range c.mappings.Funds {
			if !seen[code] {
				seen[code] = true
				funds = append(funds, code)
			}
		}
	}

	return funds
}

// GetMappings returns the raw mappings (for advanced use)
func (c *Client) GetMappings() *Mappings {
	return c.mappings
}

// GetSupplement returns the current supplement data
func (c *Client) GetSupplement() *Supplement {
	return c.supplement
}

// HasSupplement checks if a fund code exists in the supplement
func (c *Client) HasSupplement(fundCode string) bool {
	if c.supplement == nil {
		return false
	}
	_, ok := c.supplement.Funds[fundCode]
	return ok
}

// SaveSupplementEntry adds or updates a supplement entry and persists to disk
func (c *Client) SaveSupplementEntry(fundCode, firmName, categoryID, legalName, thaiName string) error {
	if c.supplement == nil {
		c.supplement = &Supplement{
			Categories: make(map[string]string),
			Funds:      make(map[string]SupplementFundInfo),
		}
	}

	// Update fund entry
	c.supplement.Funds[fundCode] = SupplementFundInfo{
		LegalName:      legalName,
		ThaiName:       thaiName,
		FirmName:       firmName,
		AIMCCategoryID: categoryID,
	}

	// Save to disk
	return c.saveSupplement()
}

// DeleteSupplementEntry removes a fund from the supplement and persists to disk
func (c *Client) DeleteSupplementEntry(fundCode string) error {
	if c.supplement == nil {
		return nil
	}

	delete(c.supplement.Funds, fundCode)
	return c.saveSupplement()
}

// saveSupplement persists the supplement to disk
func (c *Client) saveSupplement() error {
	if c.supplement == nil {
		return nil
	}

	data, err := json.MarshalIndent(c.supplement, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal supplement: %w", err)
	}

	if c.supplementPath == "" {
		c.supplementPath = filepath.Join(c.dataDir, "company_supplement.json")
	}

	if err := os.WriteFile(c.supplementPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write supplement file: %w", err)
	}

	return nil
}

// GetCategoryIDByName finds a category ID by its name (checks supplement first)
func (c *Client) GetCategoryIDByName(categoryName string) string {
	// Check supplement first
	if c.supplement != nil {
		for id, name := range c.supplement.Categories {
			if name == categoryName {
				return id
			}
		}
	}

	// Check AIMC mappings
	if c.mappings != nil {
		for id, name := range c.mappings.Categories {
			if name == categoryName {
				return id
			}
		}
	}

	return ""
}

// AIMCMetadata contains metadata about the AIMC data source
type AIMCMetadata struct {
	LastUpdate  string `json:"last_update"`
	SourceURL   string `json:"source_url,omitempty"`
	Quarter     string `json:"quarter,omitempty"`
	RecordCount int    `json:"record_count"`
}

// MappingsWithMetadata stores AIMC data with metadata
type MappingsWithMetadata struct {
	AIMCMappings
	Metadata AIMCMetadata `json:"metadata"`
}

// AIMCMappings stores AIMC category and fund information (legacy structure for compatibility)
type AIMCMappings struct {
	Categories map[string]string   `json:"categories"`
	Funds      map[string]FundInfo `json:"funds"`
}

// NeedsUpdate checks if AIMC data needs to be updated based on last update timestamp
// Returns true if data is older than 90 days (quarterly) or has never been updated
func (c *Client) NeedsUpdate() bool {
	// Read metadata if available
	metaPath := filepath.Join(c.dataDir, "aimc_mappings.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return true
	}

	var mappings struct {
		Metadata AIMCMetadata `json:"metadata"`
	}
	if err := json.Unmarshal(data, &mappings); err != nil {
		return true
	}

	if mappings.Metadata.LastUpdate == "" {
		return true
	}

	lastUpdate, err := time.Parse("2006-01-02", mappings.Metadata.LastUpdate)
	if err != nil {
		return true
	}

	// Update quarterly (90 days)
	daysSinceUpdate := time.Since(lastUpdate).Hours() / 24
	return daysSinceUpdate > 90
}

// FetchAndUpdate downloads latest AIMC data and updates the mappings
// Returns error on failure but existing data remains usable
func (c *Client) FetchAndUpdate() error {
	log.Println("[AIMC] Starting data update from AIMC website...")

	// Find the latest AIMC URL (current quarter - 1)
	url, quarter := findLatestAIMCURL()
	if url == "" {
		return fmt.Errorf("could not find valid AIMC URL")
	}

	log.Printf("[AIMC] Downloading from %s (quarter: %s)...", url, quarter)
	data, err := downloadWithRetry(url)
	if err != nil {
		return fmt.Errorf("failed to download AIMC data: %w", err)
	}

	log.Println("[AIMC] Parsing XLSX data...")
	parsedFunds, err := parseAIMCXLSX(data)
	if err != nil {
		return fmt.Errorf("failed to parse XLSX: %w", err)
	}
	log.Printf("[AIMC] Found %d funds in XLSX", len(parsedFunds))

	// Fetch Finnomena API data for category IDs
	log.Println("[AIMC] Fetching category IDs from Finnomena API...")
	categoryMap, err := fetchCategoryIDs()
	if err != nil {
		log.Printf("[AIMC] Warning: Could not fetch category IDs: %v", err)
		categoryMap = make(map[string]string)
	}

	// Merge data
	mappings := AIMCMappings{
		Categories: make(map[string]string),
		Funds:      make(map[string]FundInfo),
	}

	for code, info := range parsedFunds {
		// Add category ID from API if available
		if catID, ok := categoryMap[code]; ok {
			info.AIMCCategoryID = catID
			// Build category map
			if info.AIMCCategory != "" && catID != "" {
				mappings.Categories[catID] = info.AIMCCategory
			}
		}
		mappings.Funds[code] = info
	}

	// Save with metadata
	metadata := AIMCMetadata{
		LastUpdate:  time.Now().Format("2006-01-02"),
		SourceURL:   url,
		Quarter:     quarter,
		RecordCount: len(mappings.Funds),
	}

	if err := c.saveMappingsWithMetadata(mappings, metadata); err != nil {
		return fmt.Errorf("failed to save mappings: %w", err)
	}

	// Reload
	c.loadMappings()

	log.Printf("[AIMC] Update complete! %d companies saved", len(mappings.Funds))
	return nil
}

// FetchAndSaveNew fetches AIMC data and saves to the given data directory
// This is a standalone function to handle the case where no existing data exists
func FetchAndSaveNew(dataDir string) error {
	client := &Client{dataDir: dataDir}
	return client.FetchAndUpdate()
}

// findLatestAIMCURL finds the latest AIMC URL by going back quarter by quarter until finding a valid one
func findLatestAIMCURL() (url, quarter string) {
	now := time.Now()
	year := now.Year()
	currentMonth := int(now.Month())

	// Determine current quarter (1-4)
	currentQuarter := (currentMonth-1)/3 + 1

	// Start from previous quarter
	targetQuarter := currentQuarter - 1
	targetYear := year

	if targetQuarter < 1 {
		targetQuarter = 4
		targetYear--
	}

	// Try up to 12 quarters back (3 years)
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	for attempts := 0; attempts < 12; attempts++ {
		url = fmt.Sprintf("https://association.aimc.or.th/wp-content/uploads/%d/Aimccategory/AIMC-Category-Q%d-%d.xlsx",
			targetYear, targetQuarter, targetYear)
		quarter = fmt.Sprintf("Q%d-%d", targetQuarter, targetYear)

		// Check if URL exists with HEAD request
		resp, err := client.Head(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			log.Printf("[AIMC] Found valid URL for %s: %s", quarter, url)
			return url, quarter
		}
		if resp != nil {
			resp.Body.Close()
		}

		// Go back one more quarter
		targetQuarter--
		if targetQuarter < 1 {
			targetQuarter = 4
			targetYear--
		}
	}

	// If no URL found in 3 years, return empty
	log.Printf("[AIMC] Could not find valid AIMC URL after 12 quarters")
	return "", ""
}

// downloadWithRetry downloads AIMC XLSX with retry logic
func downloadWithRetry(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	var resp *http.Response
	var err error

	for attempt := 1; attempt <= 3; attempt++ {
		resp, err = client.Get(url)
		if err == nil {
			break
		}
		log.Printf("[AIMC] Attempt %d failed: %v", attempt, err)
		if attempt < 3 {
			sleepDuration := time.Duration(attempt) * 2 * time.Second
			log.Printf("[AIMC] Retrying in %v...", sleepDuration)
			time.Sleep(sleepDuration)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to download after 3 attempts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// parseAIMCXLSX parses AIMC XLSX data using xlsxreader
func parseAIMCXLSX(data []byte) (map[string]FundInfo, error) {
	// Write to temp file (xlsxreader needs file path)
	tmpFile, err := os.CreateTemp("", "aimc_*.xlsx")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Parse with xlsxreader
	xl, err := xlsxreader.OpenFile(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to open XLSX: %w", err)
	}
	defer xl.Close()

	funds := make(map[string]FundInfo)
	rowCount := 0

	for row := range xl.ReadRows(xl.Sheets[0]) {
		rowCount++

		// Skip header row
		if rowCount == 1 {
			continue
		}

		// Need at least 5 columns
		if len(row.Cells) < 5 {
			continue
		}

		legalName := row.Cells[0].Value
		thaiName := row.Cells[1].Value
		code := row.Cells[2].Value
		firmName := row.Cells[3].Value
		category := row.Cells[4].Value

		if code == "" {
			continue
		}

		funds[code] = FundInfo{
			LegalName:    legalName,
			ThaiName:     thaiName,
			FirmName:     firmName,
			AIMCCategory: category, // Temporary storage, will use for building categories map
		}
	}

	return funds, nil
}

// fetchCategoryIDs fetches fund codes and their AIMC category IDs from Finnomena API
func fetchCategoryIDs() (map[string]string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get("https://www.finnomena.com/fn3/api/fund/v2/public/funds")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch funds: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var response struct {
		Data []struct {
			ShortCode      string `json:"short_code"`
			AIMCCategoryID string `json:"aimc_category_id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	categoryMap := make(map[string]string)
	for _, fund := range response.Data {
		if fund.ShortCode != "" && fund.AIMCCategoryID != "" {
			categoryMap[fund.ShortCode] = fund.AIMCCategoryID
		}
	}

	return categoryMap, nil
}

// saveMappingsWithMetadata saves mappings with metadata to JSON
func (c *Client) saveMappingsWithMetadata(mappings AIMCMappings, metadata AIMCMetadata) error {
	output := struct {
		AIMCMappings
		Metadata AIMCMetadata `json:"metadata"`
	}{
		AIMCMappings: mappings,
		Metadata:     metadata,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	path := filepath.Join(c.dataDir, "aimc_mappings.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	return nil
}
