package set

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/tupkung/tis620"
)

// Client provides SET data lookup and translation services
type Client struct {
	dataDir                string
	setData                *SETData
	customTranslations     *CustomTranslations
	customTranslationsPath string
	mu                     sync.RWMutex
}

// CustomTranslations stores user-defined translations
type CustomTranslations struct {
	Version      int                    `json:"version"`
	Sectors      map[string]Translation `json:"sectors"`
	Industries   map[string]Translation `json:"industries"`
	Companies    map[string]Translation `json:"companies"`
	Untranslated []UntranslatedEntry    `json:"untranslated_log,omitempty"`
}

// Translation represents a single translation entry
type Translation struct {
	EN       string `json:"en"`
	Verified bool   `json:"verified"`
	Date     string `json:"date"`
}

// UntranslatedEntry tracks items that couldn't be translated
type UntranslatedEntry struct {
	Text   string `json:"text"`
	Type   string `json:"type"`
	Date   string `json:"date"`
	FundID string `json:"fund_id,omitempty"`
}

// Company represents a SET-listed company
type Company struct {
	NameEN     string `json:"name_en"`
	NameTH     string `json:"name_th"`
	Market     string `json:"market"`
	IndustryID string `json:"industry_id"`
	SectorID   string `json:"sector_id"`
}

// Industry represents an industry classification
type Industry struct {
	NameTH string `json:"name_th"`
	NameEN string `json:"name_en"`
}

// Sector represents a sector classification
type Sector struct {
	NameTH string `json:"name_th"`
	NameEN string `json:"name_en"`
}

// SETData contains the complete SET exchange data
type SETData struct {
	Companies  map[string]Company  `json:"companies"`
	Industries map[string]Industry `json:"industries"`
	Sectors    map[string]Sector   `json:"sectors"`
	Metadata   SETMetadata         `json:"metadata,omitempty"`
}

// SETMetadata contains metadata about the SET data source
type SETMetadata struct {
	LastUpdate  string `json:"last_update"`
	SourceURLEN string `json:"source_url_en,omitempty"`
	SourceURLTH string `json:"source_url_th,omitempty"`
	RecordCount int    `json:"record_count"`
}

const (
	SETURLEN = "https://www.set.or.th/dat/eod/listedcompany/static/listedCompanies_en_US.xls"
	SETURLTH = "https://www.set.or.th/dat/eod/listedcompany/static/listedCompanies_th_TH.xls"
)

// NewClient creates a new SET client with the given data directory
func NewClient(dataDir string) (*Client, error) {
	client := &Client{
		dataDir: dataDir,
	}

	// Load SET data
	if err := client.loadSETData(); err != nil {
		return nil, fmt.Errorf("failed to load SET data: %w", err)
	}

	// Load custom translations
	if err := client.loadCustomTranslations(); err != nil {
		// Log but don't fail - custom translations are optional
		fmt.Printf("Warning: Could not load custom translations: %v\n", err)
	}

	return client, nil
}

// loadSETData loads the official SET exchange data from JSON
func (c *Client) loadSETData() error {
	path := filepath.Join(c.dataDir, "SET_mappings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read SET_mappings.json: %w", err)
	}

	var mappings SETData
	if err := json.Unmarshal(data, &mappings); err != nil {
		return fmt.Errorf("failed to parse SET_mappings.json: %w", err)
	}

	c.setData = &mappings
	return nil
}

// loadCustomTranslations loads user-defined translations from JSON
func (c *Client) loadCustomTranslations() error {
	c.customTranslationsPath = filepath.Join(c.dataDir, "custom_translations.json")

	data, err := os.ReadFile(c.customTranslationsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create empty translations
			c.customTranslations = &CustomTranslations{
				Version:    1,
				Sectors:    make(map[string]Translation),
				Industries: make(map[string]Translation),
				Companies:  make(map[string]Translation),
			}
			return c.saveCustomTranslations()
		}
		return fmt.Errorf("failed to read custom_translations.json: %w", err)
	}

	var ct CustomTranslations
	if err := json.Unmarshal(data, &ct); err != nil {
		return fmt.Errorf("failed to parse custom_translations.json: %w", err)
	}

	// Ensure maps are initialized
	if ct.Sectors == nil {
		ct.Sectors = make(map[string]Translation)
	}
	if ct.Industries == nil {
		ct.Industries = make(map[string]Translation)
	}
	if ct.Companies == nil {
		ct.Companies = make(map[string]Translation)
	}

	c.customTranslations = &ct
	return nil
}

// saveCustomTranslations saves custom translations to disk
func (c *Client) saveCustomTranslations() error {
	if c.customTranslations == nil {
		return nil
	}

	c.mu.RLock()
	data, err := json.MarshalIndent(c.customTranslations, "", "  ")
	c.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("failed to marshal custom translations: %w", err)
	}

	return os.WriteFile(c.customTranslationsPath, data, 0644)
}

// GetCustomTranslation looks up a custom translation
func (c *Client) GetCustomTranslation(thaiText, transType string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.customTranslations == nil {
		return "", false
	}

	var trans Translation
	var found bool

	switch transType {
	case "sector":
		trans, found = c.customTranslations.Sectors[thaiText]
	case "industry":
		trans, found = c.customTranslations.Industries[thaiText]
	case "company":
		trans, found = c.customTranslations.Companies[thaiText]
	}

	if found {
		return trans.EN, true
	}
	return "", false
}

// SetCustomTranslation adds or updates a custom translation
func (c *Client) SetCustomTranslation(thaiText, englishText, transType string, verified bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.customTranslations == nil {
		return
	}

	trans := Translation{
		EN:       englishText,
		Verified: verified,
		Date:     time.Now().Format("2006-01-02"),
	}

	switch transType {
	case "sector":
		c.customTranslations.Sectors[thaiText] = trans
	case "industry":
		c.customTranslations.Industries[thaiText] = trans
	case "company":
		c.customTranslations.Companies[thaiText] = trans
	}

	// Save asynchronously
	go c.saveCustomTranslations()
}

// LogUntranslated records an item that couldn't be translated
func (c *Client) LogUntranslated(thaiText, transType, fundID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.customTranslations == nil {
		return
	}

	entry := UntranslatedEntry{
		Text:   thaiText,
		Type:   transType,
		Date:   time.Now().Format("2006-01-02"),
		FundID: fundID,
	}

	c.customTranslations.Untranslated = append(c.customTranslations.Untranslated, entry)

	// Keep only last 100 entries
	if len(c.customTranslations.Untranslated) > 100 {
		c.customTranslations.Untranslated = c.customTranslations.Untranslated[len(c.customTranslations.Untranslated)-100:]
	}
}

// GetBySymbol looks up a company by its ticker symbol
func (c *Client) GetBySymbol(symbol string) (*Company, error) {
	if c.setData == nil {
		return nil, fmt.Errorf("SET data not loaded")
	}
	symbol = strings.ToUpper(symbol)
	if company, ok := c.setData.Companies[symbol]; ok {
		return &company, nil
	}
	return nil, fmt.Errorf("symbol not found: %s", symbol)
}

// GetByName searches for a company by name (exact or fuzzy match)
func (c *Client) GetByName(name string) (*Company, error) {
	if c.setData == nil {
		return nil, fmt.Errorf("SET data not loaded")
	}

	nameUpper := strings.ToUpper(strings.TrimSpace(name))

	// Try exact match first
	for _, company := range c.setData.Companies {
		if strings.ToUpper(company.NameEN) == nameUpper {
			return &company, nil
		}
		if strings.ToUpper(company.NameTH) == nameUpper {
			return &company, nil
		}
	}

	// Try substring match
	for _, company := range c.setData.Companies {
		if strings.Contains(strings.ToUpper(company.NameEN), nameUpper) {
			return &company, nil
		}
		if strings.Contains(strings.ToUpper(company.NameTH), nameUpper) {
			return &company, nil
		}
	}

	// Progressive word removal for Thai names
	words := strings.Fields(name)
	for len(words) > 0 {
		searchName := strings.Join(words, " ")
		searchUpper := strings.ToUpper(searchName)

		for _, company := range c.setData.Companies {
			if strings.ToUpper(company.NameEN) == searchUpper {
				return &company, nil
			}
			if strings.ToUpper(company.NameTH) == searchUpper {
				return &company, nil
			}
		}

		words = words[:len(words)-1]
	}

	return nil, fmt.Errorf("company not found: %s", name)
}

// IsThaiName checks if a string contains Thai characters
func (c *Client) IsThaiName(name string) bool {
	for _, r := range name {
		if unicode.In(r, unicode.Thai) {
			return true
		}
	}
	return false
}

// TranslateName translates a Thai company name to English
func (c *Client) TranslateName(thaiName string) string {
	return c.TranslateWithFallback(thaiName, "company", "")
}

// TranslateIndustry translates a Thai industry name to English
func (c *Client) TranslateIndustry(thaiIndustry string) string {
	return c.TranslateWithFallback(thaiIndustry, "industry", "")
}

// TranslateSector translates a Thai sector name to English
func (c *Client) TranslateSector(thaiSector string) string {
	return c.TranslateWithFallback(thaiSector, "sector", "")
}

// TranslateWithFallback attempts translation using multiple sources:
// 1. Custom translations (user-defined)
// 2. Hardcoded Finnomena mappings
// 3. SET data (official Stock Exchange of Thailand)
// 4. Returns original if no translation found
func (c *Client) TranslateWithFallback(thaiText, transType, fundID string) string {
	if !c.IsThaiName(thaiText) {
		return thaiText
	}

	// 1. Check custom translations
	if trans, found := c.GetCustomTranslation(thaiText, transType); found {
		return trans
	}

	// 2. Check hardcoded Finnomena mappings
	if mapped, found := finnomenaSectorMappings[thaiText]; found {
		return mapped
	}

	// 3. Check SET data
	var setResult string
	switch transType {
	case "sector":
		setResult = c.translateFromSETSector(thaiText)
	case "industry":
		setResult = c.translateFromSETIndustry(thaiText)
	case "company":
		setResult = c.translateFromSETCompany(thaiText)
	}
	if setResult != thaiText {
		return setResult
	}

	// 4. Log untranslated for manual review
	c.LogUntranslated(thaiText, transType, fundID)

	return thaiText
}

func (c *Client) translateFromSETSector(thaiText string) string {
	if c.setData == nil {
		return thaiText
	}
	for _, sector := range c.setData.Sectors {
		if sector.NameTH == thaiText {
			return sector.NameEN
		}
	}
	return thaiText
}

func (c *Client) translateFromSETIndustry(thaiText string) string {
	if c.setData == nil {
		return thaiText
	}
	for _, industry := range c.setData.Industries {
		if industry.NameTH == thaiText {
			return industry.NameEN
		}
	}
	return thaiText
}

func (c *Client) translateFromSETCompany(thaiText string) string {
	cleaned := strings.TrimPrefix(thaiText, "หุ้นสามัญของ")
	cleaned = strings.TrimSpace(cleaned)

	company, err := c.GetByName(cleaned)
	if err == nil {
		return company.NameEN
	}
	return thaiText
}

// GetSETData returns the raw SET data (for advanced use cases)
func (c *Client) GetSETData() *SETData {
	return c.setData
}

func (c *Client) downloadWithRetry(url string) ([]byte, error) {
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
		log.Printf("[SET] Attempt %d failed for %s: %v", attempt, url, err)
		if attempt < 3 {
			sleepDuration := time.Duration(attempt) * 2 * time.Second
			log.Printf("[SET] Retrying in %v...", sleepDuration)
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

func (c *Client) parseHTMLTable(data []byte) (map[string]map[string]string, error) {
	content := string(tis620.ToUTF8(data))

	result := make(map[string]map[string]string)
	rowCount := 0

	for rowStr, ok := extractNextRow(&content); ok; rowStr, ok = extractNextRow(&content) {
		rowCount++
		if rowCount <= 2 {
			continue
		}

		cells := extractCells(rowStr)
		if record, ok := parseRowToRecord(cells); ok {
			result[record["symbol"]] = record
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no table data found in HTML (processed %d rows)", rowCount)
	}

	return result, nil
}

func extractNextRow(content *string) (string, bool) {
	trStart := "<tr"
	trEnd := "</tr>"

	trPos := strings.Index(*content, trStart)
	if trPos == -1 {
		return "", false
	}

	nextTrPos := strings.Index((*content)[trPos:], trEnd)
	if nextTrPos == -1 {
		return "", false
	}
	nextTrPos += trPos + len(trEnd)

	rowStr := (*content)[trPos:nextTrPos]
	*content = (*content)[nextTrPos:]

	return rowStr, true
}

func extractCells(rowStr string) []string {
	tdStart := "<td"
	tdEnd := "</td>"
	thStart := "<th"
	thEnd := "</th>"

	var cells []string
	cellIdx := 0

	for {
		tdPos := strings.Index(rowStr[cellIdx:], tdStart)
		if tdPos == -1 {
			tdPos = strings.Index(rowStr[cellIdx:], thStart)
			if tdPos == -1 {
				break
			}
		}
		tdPos += cellIdx

		nextTdPos := strings.Index(rowStr[tdPos:], tdEnd)
		if nextTdPos == -1 {
			nextTdPos = strings.Index(rowStr[tdPos:], thEnd)
		}
		if nextTdPos == -1 {
			break
		}
		nextTdPos += tdPos

		cell := rowStr[tdPos:nextTdPos]
		cell = cleanHTML(cell)

		if cell != "" {
			cells = append(cells, cell)
		}

		cellIdx = nextTdPos
	}

	return cells
}

func parseRowToRecord(cells []string) (map[string]string, bool) {
	if len(cells) < 2 {
		return nil, false
	}

	symbol := strings.TrimSpace(cells[0])
	if symbol == "" {
		return nil, false
	}

	record := make(map[string]string)
	record["symbol"] = symbol
	record["company"] = strings.TrimSpace(cells[1])

	if len(cells) > 2 {
		record["market"] = strings.TrimSpace(cells[2])
	}
	if len(cells) > 3 {
		record["industry"] = strings.TrimSpace(cells[3])
	}
	if len(cells) > 4 {
		record["sector"] = strings.TrimSpace(cells[4])
	}

	return record, true
}

func cleanHTML(s string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	s = re.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.NewReplacer("\n", " ", "\r", "", "\t", " ", "  ", " ").Replace(s)
	return strings.TrimSpace(s)
}

func (c *Client) mergeAndSave(enData, thData map[string]map[string]string) error {
	mappings := SETMappings{
		Companies:  make(map[string]Company),
		Industries: make(map[string]Industry),
		Sectors:    make(map[string]Sector),
	}

	industryCounter := 0
	sectorCounter := 0
	industryToID := make(map[string]string)
	sectorToID := make(map[string]string)

	for symbol, enRecord := range enData {
		thRecord := thData[symbol]

		industryID := getID("industry", thRecord, enRecord, industryToID, &industryCounter, &mappings)
		sectorID := getID("sector", thRecord, enRecord, sectorToID, &sectorCounter, &mappings)

		mappings.Companies[symbol] = Company{
			NameEN:     enRecord["company"],
			NameTH:     thRecord["company"],
			Market:     enRecord["market"],
			IndustryID: industryID,
			SectorID:   sectorID,
		}
	}

	mappings.Metadata = SETMetadata{
		LastUpdate:  time.Now().Format("2006-01-02"),
		SourceURLEN: SETURLEN,
		SourceURLTH: SETURLTH,
		RecordCount: len(mappings.Companies),
	}

	path := filepath.Join(c.dataDir, "SET_mappings.json")
	data, err := json.MarshalIndent(mappings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	c.setData = &SETData{
		Companies:  mappings.Companies,
		Industries: mappings.Industries,
		Sectors:    mappings.Sectors,
		Metadata:   mappings.Metadata,
	}

	return nil
}

type SETMappings struct {
	Companies  map[string]Company  `json:"companies"`
	Industries map[string]Industry `json:"industries"`
	Sectors    map[string]Sector   `json:"sectors"`
	Metadata   SETMetadata         `json:"metadata"`
}

func getID(prefix string, thRecord map[string]string, enRecord map[string]string, counterMap map[string]string, counter *int, mappings *SETMappings) string {
	prefixTH := thRecord[prefix]
	prefixEN := enRecord[prefix]
	prefixID := ""
	if prefixTH != "" || prefixEN != "" {
		key := fmt.Sprintf("%s|%s", prefixTH, prefixEN)
		if id, exists := counterMap[key]; exists {
			prefixID = id
		} else {
			*counter++
			prefixID = fmt.Sprintf("%s_%03d", prefix, *counter)
			counterMap[key] = prefixID

			if prefix == "industry" {
				mappings.Industries[prefixID] = Industry{NameTH: prefixTH, NameEN: prefixEN}
			} else {
				mappings.Sectors[prefixID] = Sector{NameTH: prefixTH, NameEN: prefixEN}
			}
		}
	}
	return prefixID
}

// FetchAndUpdate downloads latest SET data from official SET website
// Returns error on failure but existing data remains usable
func (c *Client) FetchAndUpdate() error {
	log.Println("[SET] Starting data update from SET website...")

	log.Printf("[SET] Downloading English data from %s...", SETURLEN)
	enDataRaw, err := c.downloadWithRetry(SETURLEN)
	if err != nil {
		return fmt.Errorf("failed to download English data: %w", err)
	}

	log.Printf("[SET] Downloading Thai data from %s...", SETURLTH)
	thDataRaw, err := c.downloadWithRetry(SETURLTH)
	if err != nil {
		return fmt.Errorf("failed to download Thai data: %w", err)
	}

	log.Println("[SET] Parsing English data...")
	enData, err := c.parseHTMLTable(enDataRaw)
	if err != nil {
		return fmt.Errorf("failed to parse English HTML: %w", err)
	}
	log.Printf("[SET] Found %d companies in English data", len(enData))

	log.Println("[SET] Parsing Thai data...")
	thData, err := c.parseHTMLTable(thDataRaw)
	if err != nil {
		return fmt.Errorf("failed to parse Thai HTML: %w", err)
	}
	log.Printf("[SET] Found %d companies in Thai data", len(thData))

	log.Println("[SET] Merging and saving data...")
	if err := c.mergeAndSave(enData, thData); err != nil {
		return fmt.Errorf("failed to save data: %w", err)
	}

	log.Printf("[SET] Update complete! %d companies saved", len(enData))
	return nil
}

// NeedsUpdate checks if SET data needs to be updated based on last update timestamp
// Returns true if data is older than 30 days or has never been updated
func (c *Client) NeedsUpdate() bool {
	if c.setData == nil || c.setData.Metadata.LastUpdate == "" {
		return true
	}

	lastUpdate, err := time.Parse("2006-01-02", c.setData.Metadata.LastUpdate)
	if err != nil {
		return true
	}

	daysSinceUpdate := time.Since(lastUpdate).Hours() / 24
	return daysSinceUpdate > 30
}

// FetchAndSaveNew fetches SET data from the website and saves to the given data directory
// This is a standalone function to handle the case where no existing data exists
func FetchAndSaveNew(dataDir string) error {
	client := &Client{dataDir: dataDir}

	log.Printf("[SET] Downloading English data from %s...", SETURLEN)
	enDataRaw, err := client.downloadWithRetry(SETURLEN)
	if err != nil {
		return fmt.Errorf("failed to download English data: %w", err)
	}

	log.Printf("[SET] Downloading Thai data from %s...", SETURLTH)
	thDataRaw, err := client.downloadWithRetry(SETURLTH)
	if err != nil {
		return fmt.Errorf("failed to download Thai data: %w", err)
	}

	log.Println("[SET] Parsing English data...")
	enData, err := client.parseHTMLTable(enDataRaw)
	if err != nil {
		return fmt.Errorf("failed to parse English HTML: %w", err)
	}
	log.Printf("[SET] Found %d companies in English data", len(enData))

	log.Println("[SET] Parsing Thai data...")
	thData, err := client.parseHTMLTable(thDataRaw)
	if err != nil {
		return fmt.Errorf("failed to parse Thai HTML: %w", err)
	}
	log.Printf("[SET] Found %d companies in Thai data", len(thData))

	log.Println("[SET] Merging and saving data...")
	if err := client.mergeAndSave(enData, thData); err != nil {
		return fmt.Errorf("failed to save data: %w", err)
	}

	log.Printf("[SET] Update complete! %d companies saved", len(enData))
	return nil
}

// Hardcoded mappings for common Finnomena sectors not in SET data
// These are asset allocation and fund-specific terms that differ from SET exchange sectors
var finnomenaSectorMappings = map[string]string{
	// Asset allocation categories
	"หุ้น":                        "Equities",
	"เงินฝาก":                     "Cash & Deposits",
	"เงินฝากธนาคาร":               "Bank Deposits",
	"เงินฝากธนาคาร P/N และ B/E":   "Bank Deposits P/N & B/E",
	"ตราสารหนี้":                  "Bonds",
	"ตราสารอนุพันธ์":              "Derivatives",
	"สินทรัพย์อื่นๆ":              "Other Assets",
	"สินทรัพย์อื่นๆ/หนี้สินอื่นๆ": "Other Assets/Liabilities",
	"หน่วยลงทุน":                  "Fund Units",
	"หนี้สิน":                     "Liabilities",

	// GICS/Standard sector classifications (different from SET)
	"เทคโนโลยี":                  "Technology",
	"อุตสาหกรรม":                 "Industrials",
	"อรรถประโยชน์":               "Utilities",
	"บริการด้านการสื่อสาร":       "Communication Services",
	"บริการด้านการเงิน":          "Financial Services",
	"พลังงาน":                    "Energy",
	"สินค้าฟุ่มเฟือย/ตามวัฏจักร": "Consumer Discretionary",
	"สินค้าอุปโภคบริโภค":         "Consumer Staples",
	"อสังหาริมทรัพย์":            "Real Estate",
	"วัสดุทั่วไป":                "Materials",
	"สาธารณูปโภคพื้นฐาน":         "Utilities",
	"การดูแลสุขภาพ":              "Health Care",

	// Generic/fund terms
	"เงินสด": "Cash",
	"รวม":    "Total",
	"สุทธิ":  "Net",
}
