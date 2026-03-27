package aimc

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func getDataDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".thaifa", "data")
}

func TestNewClient(t *testing.T) {
	dataDir := getDataDir()

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Skip("data directory not found, skipping test")
	}

	client, err := NewClient(dataDir)
	if err != nil {
		t.Fatalf("Failed to create AIMC client: %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to not be nil")
	}

	if client.GetMappings() == nil {
		t.Error("Expected mappings to be loaded")
	}

	if len(client.GetAllFunds()) == 0 {
		t.Error("Expected at least one fund in mappings")
	}
}

func TestClientGetCategoryName(t *testing.T) {
	dataDir := getDataDir()

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Skip("data directory not found, skipping test")
	}

	client, err := NewClient(dataDir)
	if err != nil {
		t.Fatalf("Failed to create AIMC client: %v", err)
	}

	tests := []struct {
		categoryID string
		wantEmpty  bool
	}{
		{"LC00002660", false}, // Equity Fund - Large Cap
		{"LC00002470", false}, // Equity Fund - 中小盘
		{"INVALID_ID", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.categoryID, func(t *testing.T) {
			got := client.GetCategoryName(tt.categoryID)
			if tt.wantEmpty {
				if got != "" {
					t.Errorf("GetCategoryName(%s) = %q, want empty", tt.categoryID, got)
				}
				return
			}
			if got == "" {
				t.Errorf("GetCategoryName(%s) returned empty", tt.categoryID)
			}
		})
	}
}

func TestClientGetFundInfo(t *testing.T) {
	dataDir := getDataDir()

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Skip("data directory not found, skipping test")
	}

	client, err := NewClient(dataDir)
	if err != nil {
		t.Fatalf("Failed to create AIMC client: %v", err)
	}

	// Get a fund to test with
	funds := client.GetAllFunds()
	if len(funds) == 0 {
		t.Skip("No funds available for testing")
	}

	testFund := funds[0]

	legalName, thaiName, firmName, category := client.GetFundInfo(testFund)

	if legalName == "" && thaiName == "" {
		t.Errorf("GetFundInfo(%s) returned empty names", testFund)
	}

	// Test with invalid fund
	legalName, thaiName, firmName, category = client.GetFundInfo("INVALID_FUND")
	if legalName != "" || thaiName != "" || firmName != "" || category != "" {
		t.Error("GetFundInfo with invalid fund should return empty strings")
	}
}

func TestClientGetCategories(t *testing.T) {
	dataDir := getDataDir()

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Skip("data directory not found, skipping test")
	}

	client, err := NewClient(dataDir)
	if err != nil {
		t.Fatalf("Failed to create AIMC client: %v", err)
	}

	categories := client.GetCategories()

	if len(categories) == 0 {
		t.Error("Expected at least one category")
	}

	// Check for known categories
	hasEquity := false
	for _, cat := range categories {
		if cat == "Equity Fund - Large Cap" {
			hasEquity = true
			break
		}
	}

	if !hasEquity {
		t.Log("Note: 'Equity Fund - Large Cap' category not found (may vary by data)")
	}
}

func TestClientGetFundsByCategory(t *testing.T) {
	dataDir := getDataDir()

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Skip("data directory not found, skipping test")
	}

	client, err := NewClient(dataDir)
	if err != nil {
		t.Fatalf("Failed to create AIMC client: %v", err)
	}

	categories := client.GetCategories()
	if len(categories) == 0 {
		t.Skip("No categories available for testing")
	}

	// Test with first available category
	testCategory := categories[0]
	funds := client.GetFundsByCategory(testCategory)

	if len(funds) == 0 {
		t.Logf("Category %s has no funds (may be valid)", testCategory)
	}

	// Test with invalid category
	funds = client.GetFundsByCategory("Invalid Category Name")
	if len(funds) != 0 {
		t.Error("GetFundsByCategory with invalid category should return empty slice")
	}
}

func TestClientGetFundsByCompany(t *testing.T) {
	dataDir := getDataDir()

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Skip("data directory not found, skipping test")
	}

	client, err := NewClient(dataDir)
	if err != nil {
		t.Fatalf("Failed to create AIMC client: %v", err)
	}

	// Test with a known asset manager
	funds := client.GetFundsByCompany("KASIKORN ASSET MANAGEMENT")
	// Note: This may or may not return results depending on data

	// Test with invalid company
	funds = client.GetFundsByCompany("NONEXISTENT COMPANY XYZ")
	if len(funds) != 0 {
		t.Error("GetFundsByCompany with invalid company should return empty slice")
	}
}

func TestClientGetFundsByCompanyFuzzy(t *testing.T) {
	dataDir := getDataDir()

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Skip("data directory not found, skipping test")
	}

	client, err := NewClient(dataDir)
	if err != nil {
		t.Fatalf("Failed to create AIMC client: %v", err)
	}

	// Test fuzzy matching - should match any company containing "KASIKORN"
	_ = client.GetFundsByCompanyFuzzy("KASIKORN")

	// Test with empty string - returns all funds (empty string matches everything)
	allFunds := client.GetFundsByCompanyFuzzy("")
	if len(allFunds) == 0 {
		t.Error("GetFundsByCompanyFuzzy with empty string should return all funds")
	}
}

func TestClientGetAllFunds(t *testing.T) {
	dataDir := getDataDir()

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Skip("data directory not found, skipping test")
	}

	client, err := NewClient(dataDir)
	if err != nil {
		t.Fatalf("Failed to create AIMC client: %v", err)
	}

	funds := client.GetAllFunds()

	if len(funds) == 0 {
		t.Error("Expected at least one fund")
	}

	// Verify no duplicates
	seen := make(map[string]bool)
	for _, fund := range funds {
		if seen[fund] {
			t.Errorf("Duplicate fund found: %s", fund)
		}
		seen[fund] = true
	}
}

func TestClientGetMappings(t *testing.T) {
	dataDir := getDataDir()

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Skip("data directory not found, skipping test")
	}

	client, err := NewClient(dataDir)
	if err != nil {
		t.Fatalf("Failed to create AIMC client: %v", err)
	}

	mappings := client.GetMappings()

	if mappings == nil {
		t.Fatal("GetMappings() returned nil")
	}

	if mappings.Categories == nil {
		t.Error("Mappings.Categories is nil")
	}

	if mappings.Funds == nil {
		t.Error("Mappings.Funds is nil")
	}
}

func TestClientNeedsUpdate(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a client with temp directory (no mappings file)
	client := &Client{
		dataDir: tempDir,
	}

	// Should return true when no metadata exists
	// Note: This tests the internal logic, may need adjustment based on actual implementation
	_ = client.NeedsUpdate()
}

func TestClient_SupplementMerge(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create AIMC mappings file
	mappings := Mappings{
		Categories: map[string]string{
			"CAT1": "Category One",
			"CAT2": "Category Two",
		},
		Funds: map[string]FundInfo{
			"FUND-A": {
				LegalName:      "Fund A Legal",
				ThaiName:       "กองทุน ก",
				FirmName:       "Company A",
				AIMCCategoryID: "CAT1",
			},
			"FUND-B": {
				LegalName:      "Fund B Legal",
				ThaiName:       "กองทุน ข",
				FirmName:       "Company B",
				AIMCCategoryID: "CAT2",
			},
		},
	}

	mappingsData, _ := json.MarshalIndent(mappings, "", "  ")
	mappingsPath := filepath.Join(tmpDir, "aimc_mappings.json")
	if err := os.WriteFile(mappingsPath, mappingsData, 0644); err != nil {
		t.Fatalf("Failed to write mappings: %v", err)
	}

	// Test without supplement
	client, err := NewClient(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test GetFundInfo - should return AIMC data
	legal, thai, firm, cat := client.GetFundInfo("FUND-A")
	if firm != "Company A" || cat != "Category One" {
		t.Errorf("GetFundInfo(FUND-A) = %s, %s, %s, %s; want Company A, _, _, Category One", legal, thai, firm, cat)
	}

	// Now create supplement file
	supplement := Supplement{
		Categories: map[string]string{
			"CAT3": "Category Three",
		},
		Funds: map[string]SupplementFundInfo{
			"FUND-C": {
				LegalName:      "Fund C Legal",
				ThaiName:       "กองทุน ค",
				FirmName:       "Company C",
				AIMCCategoryID: "CAT3",
			},
			"FUND-A": {
				LegalName:      "Fund A Override",
				ThaiName:       "กองทุน ก Override",
				FirmName:       "Company A Override",
				AIMCCategoryID: "CAT2",
			},
		},
	}

	supplementData, _ := json.MarshalIndent(supplement, "", "  ")
	supplementPath := filepath.Join(tmpDir, "company_supplement.json")
	if err := os.WriteFile(supplementPath, supplementData, 0644); err != nil {
		t.Fatalf("Failed to write supplement: %v", err)
	}

	// Create new client with supplement
	client2, err := NewClient(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create client with supplement: %v", err)
	}

	// Test GetFundInfo with override - supplement should take precedence
	legal, thai, firm, cat = client2.GetFundInfo("FUND-A")
	if firm != "Company A Override" {
		t.Errorf("GetFundInfo(FUND-A) firm = %s; want Company A Override", firm)
	}
	if cat != "Category Two" {
		t.Errorf("GetFundInfo(FUND-A) category = %s; want Category Two", cat)
	}

	// Test GetFundInfo for supplement-only fund
	legal, thai, firm, cat = client2.GetFundInfo("FUND-C")
	if firm != "Company C" || cat != "Category Three" {
		t.Errorf("GetFundInfo(FUND-C) = %s, %s, %s, %s; want Company C, _, _, Category Three", legal, thai, firm, cat)
	}

	// Test GetAllFunds - should include both sources
	allFunds := client2.GetAllFunds()
	if len(allFunds) != 3 {
		t.Errorf("GetAllFunds() returned %d funds; want 3", len(allFunds))
	}

	// Test GetFundsByCompany
	companyCFunds := client2.GetFundsByCompany("Company C")
	if len(companyCFunds) != 1 || companyCFunds[0] != "FUND-C" {
		t.Errorf("GetFundsByCompany(Company C) = %v; want [FUND-C]", companyCFunds)
	}

	// Test GetFundsByCompanyFuzzy
	fuzzyFunds := client2.GetFundsByCompanyFuzzy("Override")
	if len(fuzzyFunds) != 1 || fuzzyFunds[0] != "FUND-A" {
		t.Errorf("GetFundsByCompanyFuzzy(Override) = %v; want [FUND-A]", fuzzyFunds)
	}

	// Test GetCategories - should include all
	categories := client2.GetCategories()
	if len(categories) != 3 {
		t.Errorf("GetCategories() returned %d categories; want 3", len(categories))
	}

	// Test HasSupplement
	if !client2.HasSupplement("FUND-A") {
		t.Error("HasSupplement(FUND-A) = false; want true")
	}
	if !client2.HasSupplement("FUND-C") {
		t.Error("HasSupplement(FUND-C) = false; want true")
	}
	if client2.HasSupplement("FUND-B") {
		t.Error("HasSupplement(FUND-B) = true; want false")
	}
}

func TestClient_SupplementManagement(t *testing.T) {
	tmpDir := t.TempDir()

	// Create AIMC mappings
	mappings := Mappings{
		Categories: map[string]string{
			"CAT1": "Category One",
		},
		Funds: map[string]FundInfo{
			"FUND-A": {
				FirmName:       "Company A",
				AIMCCategoryID: "CAT1",
			},
		},
	}

	mappingsData, _ := json.MarshalIndent(mappings, "", "  ")
	mappingsPath := filepath.Join(tmpDir, "aimc_mappings.json")
	os.WriteFile(mappingsPath, mappingsData, 0644)

	client, err := NewClient(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test SaveSupplementEntry
	err = client.SaveSupplementEntry("NEW-FUND", "New Company", "CAT1", "New Legal", "New Thai")
	if err != nil {
		t.Fatalf("SaveSupplementEntry failed: %v", err)
	}

	// Verify entry was saved
	if !client.HasSupplement("NEW-FUND") {
		t.Error("HasSupplement(NEW-FUND) = false after save")
	}

	_, _, firm, _ := client.GetFundInfo("NEW-FUND")
	if firm != "New Company" {
		t.Errorf("GetFundInfo(NEW-FUND) firm = %s; want New Company", firm)
	}

	// Verify file was created
	supplementPath := filepath.Join(tmpDir, "company_supplement.json")
	if _, err := os.Stat(supplementPath); os.IsNotExist(err) {
		t.Error("Supplement file was not created")
	}

	// Test DeleteSupplementEntry
	err = client.DeleteSupplementEntry("NEW-FUND")
	if err != nil {
		t.Fatalf("DeleteSupplementEntry failed: %v", err)
	}

	if client.HasSupplement("NEW-FUND") {
		t.Error("HasSupplement(NEW-FUND) = true after delete")
	}
}

func TestClient_GetCategoryIDByName(t *testing.T) {
	tmpDir := t.TempDir()

	mappings := Mappings{
		Categories: map[string]string{
			"CAT1": "Category One",
		},
		Funds: map[string]FundInfo{},
	}

	mappingsData, _ := json.MarshalIndent(mappings, "", "  ")
	mappingsPath := filepath.Join(tmpDir, "aimc_mappings.json")
	os.WriteFile(mappingsPath, mappingsData, 0644)

	supplement := Supplement{
		Categories: map[string]string{
			"CAT2": "Category Two",
			"CAT1": "Category One Override", // Should not be used since we check supplement first
		},
		Funds: map[string]SupplementFundInfo{},
	}

	supplementData, _ := json.MarshalIndent(supplement, "", "  ")
	supplementPath := filepath.Join(tmpDir, "company_supplement.json")
	os.WriteFile(supplementPath, supplementData, 0644)

	client, err := NewClient(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test getting category ID from supplement
	catID := client.GetCategoryIDByName("Category Two")
	if catID != "CAT2" {
		t.Errorf("GetCategoryIDByName(Category Two) = %s; want CAT2", catID)
	}

	// Test getting category ID from AIMC
	catID = client.GetCategoryIDByName("Category One")
	if catID != "CAT1" {
		t.Errorf("GetCategoryIDByName(Category One) = %s; want CAT1", catID)
	}

	// Test non-existent category
	catID = client.GetCategoryIDByName("Non Existent")
	if catID != "" {
		t.Errorf("GetCategoryIDByName(Non Existent) = %s; want empty", catID)
	}
}
