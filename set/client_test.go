package set

import (
	"os"
	"path/filepath"
	"strings"
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
		t.Fatalf("Failed to create SET client: %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to not be nil")
	}

	if client.GetSETData() == nil {
		t.Error("Expected SET data to be loaded")
	}

	if len(client.GetSETData().Companies) == 0 {
		t.Error("Expected at least one company in Companies map")
	}
}

func TestClientGetBySymbol(t *testing.T) {
	dataDir := getDataDir()

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Skip("data directory not found, skipping test")
	}

	client, err := NewClient(dataDir)
	if err != nil {
		t.Fatalf("Failed to create SET client: %v", err)
	}

	tests := []struct {
		symbol    string
		wantError bool
	}{
		{"PTT", false},
		{"KBANK", false},
		{"SCB", false},
		{"INVALID", true},
	}

	for _, tt := range tests {
		t.Run(tt.symbol, func(t *testing.T) {
			company, err := client.GetBySymbol(tt.symbol)
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error for %s", tt.symbol)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if company.NameEN == "" {
				t.Errorf("GetBySymbol(%s) returned empty NameEN", tt.symbol)
			}
		})
	}
}

func TestClientIsThaiName(t *testing.T) {
	client, _ := NewClient(getDataDir())
	if client == nil {
		t.Skip("Could not create client")
	}

	tests := []struct {
		name string
		want bool
	}{
		{"PTT PUBLIC COMPANY LIMITED", false},
		{"บริษัท ปตท. จำกัด (มหาชน)", true},
		{"หุ้นสามัญของบริษัท ปตท. จำกัด (มหาชน)", true},
		{"KASIKORN BANK", false},
		{"ธนาคารกสิกรไทย", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := client.IsThaiName(tt.name); got != tt.want {
				t.Errorf("IsThaiName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestClientGetByName(t *testing.T) {
	dataDir := getDataDir()

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Skip("data directory not found, skipping test")
	}

	client, err := NewClient(dataDir)
	if err != nil {
		t.Fatalf("Failed to create SET client: %v", err)
	}

	tests := []struct {
		name    string
		wantErr bool
	}{
		{"PTT PUBLIC COMPANY LIMITED", false},
		{"บริษัท ปตท. จำกัด (มหาชน)", false},
		{"KASIKORNBANK", false},
		{"ธนาคารกสิกรไทย", false},
		{"NonExistent Company XYZ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			company, err := client.GetByName(tt.name)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for %s", tt.name)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if company.NameEN == "" {
				t.Errorf("GetByName(%s) returned empty NameEN", tt.name)
			}
		})
	}
}

func TestClientTranslateName(t *testing.T) {
	dataDir := getDataDir()

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Skip("data directory not found, skipping test")
	}

	client, err := NewClient(dataDir)
	if err != nil {
		t.Fatalf("Failed to create SET client: %v", err)
	}

	tests := []struct {
		input    string
		contains string
	}{
		{"หุ้นสามัญของบริษัท ปตท. จำกัด (มหาชน)", "PTT"},
		{"บริษัท ปตท. จำกัด (มหาชน)", "PTT"},
		{"English Name Only", "English Name Only"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := client.TranslateName(tt.input)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("TranslateName(%q) = %q, should contain %q", tt.input, got, tt.contains)
			}
		})
	}
}

func TestClientTranslateSector(t *testing.T) {
	dataDir := getDataDir()

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Skip("data directory not found, skipping test")
	}

	client, err := NewClient(dataDir)
	if err != nil {
		t.Fatalf("Failed to create SET client: %v", err)
	}

	tests := []struct {
		thai   string
		wantEN string
	}{
		{"ขนส่งและโลจิสติกส์", "Transportation & Logistics"},
		{"พลังงานและสาธารณูปโภค", "Energy & Utilities"},
		{"เทคโนโลยีสารสนเทศและการสื่อสาร", "Information & Communication Technology"},
		{"พัฒนาอสังหาริมทรัพย์", "Property Development"},
		{"Unknown Sector", "Unknown Sector"},
	}

	for _, tt := range tests {
		t.Run(tt.thai, func(t *testing.T) {
			got := client.TranslateSector(tt.thai)
			if got != tt.wantEN {
				t.Errorf("TranslateSector(%q) = %q, want %q", tt.thai, got, tt.wantEN)
			}
		})
	}
}

func TestClientTranslateIndustry(t *testing.T) {
	dataDir := getDataDir()

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Skip("data directory not found, skipping test")
	}

	client, err := NewClient(dataDir)
	if err != nil {
		t.Fatalf("Failed to create SET client: %v", err)
	}

	tests := []struct {
		thai   string
		wantEN string
	}{
		{"ธุรกิจการเงิน", "Financials"},
		{"ทรัพยากร", "Resources"},
		{"เทคโนโลยี", "Technology"},
		{"Unknown Industry", "Unknown Industry"},
	}

	for _, tt := range tests {
		t.Run(tt.thai, func(t *testing.T) {
			got := client.TranslateIndustry(tt.thai)
			if got != tt.wantEN {
				t.Errorf("TranslateIndustry(%q) = %q, want %q", tt.thai, got, tt.wantEN)
			}
		})
	}
}
