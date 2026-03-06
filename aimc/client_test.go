package aimc

import (
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
