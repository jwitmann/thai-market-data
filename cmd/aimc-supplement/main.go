package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jwitmann/thai-market-data/aimc"
)

// FinnomenaFund represents a fund from the Finnomena API
type FinnomenaFund struct {
	FundID         string `json:"fund_id"`
	ShortCode      string `json:"short_code"`
	NameTH         string `json:"name_th"`
	NameEN         string `json:"name_en"`
	AIMCCategoryID string `json:"aimc_category_id"`
	SECIsActive    bool   `json:"sec_is_active"`
}

// PatternMatch represents a detected pattern for a fund
type PatternMatch struct {
	FundCode    string
	FundName    string
	DetectedCo  string
	Confidence  string // "high", "medium", "low"
	SuggestedCo string
	CategoryID  string
}

// Known company prefix patterns
type CompanyPattern struct {
	Prefix      string
	CompanyName string
}

var companyPatterns = []CompanyPattern{
	{"SCB", "SCB Asset Management"},
	{"SCBAM", "SCB Asset Management"},
	{"KTAM", "Krung Thai Asset Management"},
	{"KT-A", "Krung Thai Asset Management"},
	{"TMBAM", "TMB Asset Management"},
	{"KASIKORN", "Kasikorn Asset Management"},
	{"KASIKAM", "Kasikorn Asset Management"},
	{"K-G", "Kasikorn Asset Management"},
	{"BAY", "Bank of Ayudhya Asset Management"},
	{"BBL", "Bangkok Bank Asset Management"},
	{"B-", "Bangkok Bank Asset Management"},
	{"Krungsri", "Krungsri Asset Management"},
	{"UOBAM", "UOB Asset Management"},
	{"UOB", "UOB Asset Management"},
	{"TISCO", "TISCO Asset Management"},
	{"TISCA", "TISCO Asset Management"},
	{"KFS", "Krungsri Asset Management"},
	{"PHILLIP", "Phillip Asset Management"},
	{"PRINCIPAL", "Principal Asset Management"},
	{"PRINC", "Principal Asset Management"},
	{"ONEAM", "One Asset Management"},
	{"LH", "Land and Houses Asset Management"},
	{"LHAM", "Land and Houses Asset Management"},
	{"CIMB", "CIMB-Principal Asset Management"},
	{"MFC", "MFC Asset Management"},
	{"MFCAM", "MFC Asset Management"},
	{"DAOLIO", "Daol Securities"},
	{"DAOL", "Daol Securities"},
	{"SENA", "Sena Asset Management"},
	{"PAM", "Phillip Asset Management"},
	{"FEF", "FEF Asset Management"},
	{",J", "Jayne Asset Management"},
}

func main() {
	var (
		dataDir     = flag.String("data-dir", "", "Data directory (default: ~/.thaifa/data)")
		outputFile  = flag.String("output", "", "Output file (default: {data-dir}/company_supplement.json)")
		interactive = flag.Bool("i", false, "Interactive mode - prompt for confirmation")
		showAll     = flag.Bool("all", false, "Show all unmapped funds (not just pattern matches)")
	)
	flag.Parse()

	// Determine data directory
	if *dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		*dataDir = filepath.Join(home, ".thaifa", "data")
	}

	// Determine output file
	if *outputFile == "" {
		*outputFile = filepath.Join(*dataDir, "company_supplement.json")
	}

	fmt.Printf("Loading AIMC data from: %s\n", *dataDir)

	// Load AIMC client
	client, err := aimc.NewClient(*dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading AIMC data: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Fetching funds from Finnomena API...")

	// Fetch all funds from Finnomena API
	funds, err := fetchFinnomenaFunds()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching funds: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d funds from API\n", len(funds))

	// Find unmapped funds (not in AIMC, not in supplement)
	var unmapped []FinnomenaFund
	for _, fund := range funds {
		if !fund.SECIsActive {
			continue
		}

		// Check if in AIMC mappings
		_, _, firmName, _ := client.GetFundInfo(fund.ShortCode)
		if firmName == "" {
			unmapped = append(unmapped, fund)
		}
	}

	fmt.Printf("Found %d unmapped active funds\n", len(unmapped))

	// Detect patterns
	matches := detectPatterns(unmapped, client)

	if *showAll {
		fmt.Printf("\n=== All Unmapped Funds (%d) ===\n", len(unmapped))
		for _, fund := range unmapped {
			fmt.Printf("  %s: %s\n", fund.ShortCode, fund.NameEN)
		}
	}

	// Display pattern matches
	if len(matches) == 0 {
		fmt.Println("\nNo pattern matches found for unmapped funds.")
		return
	}

	fmt.Printf("\n=== Pattern Matches (%d) ===\n", len(matches))

	// Group by confidence
	highConf := filterByConfidence(matches, "high")
	medConf := filterByConfidence(matches, "medium")
	lowConf := filterByConfidence(matches, "low")

	if len(highConf) > 0 {
		fmt.Printf("\nHigh Confidence (%d):\n", len(highConf))
		for _, m := range highConf {
			fmt.Printf("  %s -> %s\n", m.FundCode, m.SuggestedCo)
		}
	}

	if len(medConf) > 0 {
		fmt.Printf("\nMedium Confidence (%d):\n", len(medConf))
		for _, m := range medConf {
			fmt.Printf("  %s -> %s\n", m.FundCode, m.SuggestedCo)
		}
	}

	if len(lowConf) > 0 {
		fmt.Printf("\nLow Confidence (%d):\n", len(lowConf))
		for _, m := range lowConf {
			fmt.Printf("  %s -> %s (detected: %s)\n", m.FundCode, m.SuggestedCo, m.DetectedCo)
		}
	}

	// Generate supplement entries
	supplement := generateSupplement(matches, client)

	// Interactive mode
	if *interactive {
		supplement = interactiveMode(matches, supplement, client)
	}

	// Save supplement
	if len(supplement.Funds) > 0 {
		if err := saveSupplement(supplement, *outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving supplement: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\nSaved %d entries to %s\n", len(supplement.Funds), *outputFile)
	}
}

func fetchFinnomenaFunds() ([]FinnomenaFund, error) {
	resp, err := http.Get("https://www.finnomena.com/fn3/api/fund/v2/public/funds")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var result struct {
		Data []FinnomenaFund `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func detectPatterns(funds []FinnomenaFund, client *aimc.Client) []PatternMatch {
	var matches []PatternMatch

	for _, fund := range funds {
		code := fund.ShortCode
		name := fund.NameEN

		// Try to match by code prefix
		detected, confidence, suggested := detectCompanyFromCode(code, name)

		if detected != "" {
			matches = append(matches, PatternMatch{
				FundCode:    code,
				FundName:    name,
				DetectedCo:  detected,
				Confidence:  confidence,
				SuggestedCo: suggested,
				CategoryID:  fund.AIMCCategoryID,
			})
		}
	}

	return matches
}

func detectCompanyFromCode(code, name string) (detected, confidence, suggested string) {
	codeUpper := strings.ToUpper(code)

	// Sort patterns by length (longest first) to match most specific first
	patterns := make([]CompanyPattern, len(companyPatterns))
	copy(patterns, companyPatterns)
	sort.Slice(patterns, func(i, j int) bool {
		return len(patterns[i].Prefix) > len(patterns[j].Prefix)
	})

	for _, pattern := range patterns {
		prefix := strings.ToUpper(pattern.Prefix)

		// Exact match at start
		if strings.HasPrefix(codeUpper, prefix) {
			// Check if next char is digit, -, or end of string (avoid partial matches)
			if len(codeUpper) == len(prefix) ||
				codeUpper[len(prefix)] == '-' ||
				(codeUpper[len(prefix)] >= '0' && codeUpper[len(prefix)] <= '9') {
				return pattern.Prefix, "high", pattern.CompanyName
			}
		}

		// Check if contained within first part (before first -)
		if idx := strings.Index(codeUpper, "-"); idx > 0 {
			firstPart := codeUpper[:idx]
			if firstPart == prefix {
				return pattern.Prefix, "high", pattern.CompanyName
			}
		}
	}

	// No pattern match
	return "", "", ""
}

func filterByConfidence(matches []PatternMatch, confidence string) []PatternMatch {
	var filtered []PatternMatch
	for _, m := range matches {
		if m.Confidence == confidence {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func generateSupplement(matches []PatternMatch, client *aimc.Client) *aimc.Supplement {
	supplement := &aimc.Supplement{
		Categories: make(map[string]string),
		Funds:      make(map[string]aimc.SupplementFundInfo),
	}

	// Build category map from AIMC
	if mappings := client.GetMappings(); mappings != nil {
		for id, name := range mappings.Categories {
			supplement.Categories[id] = name
		}
	}

	for _, m := range matches {
		// Only include high confidence matches
		if m.Confidence != "high" {
			continue
		}

		supplement.Funds[m.FundCode] = aimc.SupplementFundInfo{
			FirmName:       m.SuggestedCo,
			AIMCCategoryID: m.CategoryID,
		}
	}

	return supplement
}

func interactiveMode(matches []PatternMatch, supplement *aimc.Supplement, client *aimc.Client) *aimc.Supplement {
	fmt.Println("\n=== Interactive Mode ===")
	fmt.Println("Review each fund and confirm (y/n) or edit company name (e)")

	confirmed := &aimc.Supplement{
		Categories: make(map[string]string),
		Funds:      make(map[string]aimc.SupplementFundInfo),
	}

	// Copy categories
	for k, v := range supplement.Categories {
		confirmed.Categories[k] = v
	}

	for _, m := range matches {
		fmt.Printf("\n%s: %s\n", m.FundCode, m.FundName)
		fmt.Printf("  Suggested company: %s (%s confidence)\n", m.SuggestedCo, m.Confidence)
		fmt.Print("  Confirm? (y/n/e[edit]/s[skip rest]): ")

		var input string
		fmt.Scanln(&input)
		input = strings.ToLower(strings.TrimSpace(input))

		switch input {
		case "y", "yes":
			confirmed.Funds[m.FundCode] = aimc.SupplementFundInfo{
				FirmName:       m.SuggestedCo,
				AIMCCategoryID: m.CategoryID,
			}
		case "e", "edit":
			fmt.Print("  Enter company name: ")
			var company string
			fmt.Scanln(&company)
			if company != "" {
				confirmed.Funds[m.FundCode] = aimc.SupplementFundInfo{
					FirmName:       company,
					AIMCCategoryID: m.CategoryID,
				}
			}
		case "s", "skip":
			fmt.Println("  Skipping remaining...")
			break
		default:
			fmt.Println("  Skipped")
		}
	}

	return confirmed
}

func saveSupplement(supplement *aimc.Supplement, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(supplement, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
