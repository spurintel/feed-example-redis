package storage

import (
	"encoding/json"
	"feedexampleredis/internal/spur"
	"strings"
	"testing"
)

func TestBackwardCompatibilityJSONUnmarshal(t *testing.T) {
	// Test data simulating old JSON format with empty client object
	oldFormatJSON := `{
		"ip": "1.2.3.4",
		"location": {"country": "US", "city": "New York"},
		"as": {"number": 123, "organization": "Test Org"},
		"client": {}
	}`

	// Unmarshal the old format JSON
	var ipContext spur.IPContext
	err := json.Unmarshal([]byte(oldFormatJSON), &ipContext)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Before fix: Client would be non-nil empty struct
	t.Logf("Client field after unmarshal: %+v", ipContext.Client)
	t.Logf("Client == nil? %v", ipContext.Client == nil)

	// Apply the backward compatibility fix
	if ipContext.Client != nil && isEmptyClient(ipContext.Client) {
		ipContext.Client = nil
	}

	// After fix: Client should be nil
	if ipContext.Client != nil {
		t.Errorf("Expected Client to be nil after fix, but got: %+v", ipContext.Client)
	}

	// Test JSON serialization to ensure client is omitted
	jsonData, err := json.Marshal(ipContext)
	if err != nil {
		t.Fatalf("Failed to marshal IPContext: %v", err)
	}

	jsonStr := string(jsonData)
	t.Logf("JSON output after fix: %s", jsonStr)

	// Check if client field is present in JSON
	if strings.Contains(jsonStr, "client") {
		t.Errorf("Client field should be omitted from JSON, but found: %s", jsonStr)
	}

	// Verify other fields are preserved
	if ipContext.IP != "1.2.3.4" {
		t.Errorf("Expected IP '1.2.3.4', got %s", ipContext.IP)
	}
	if ipContext.Location.Country != "US" {
		t.Errorf("Expected country 'US', got %s", ipContext.Location.Country)
	}
	if ipContext.AS.Number != 123 {
		t.Errorf("Expected AS number 123, got %d", ipContext.AS.Number)
	}
}

func TestIsEmptyClient(t *testing.T) {
	tests := []struct {
		name     string
		client   *spur.Client
		expected bool
	}{
		{
			name:     "nil client",
			client:   nil,
			expected: false, // Function shouldn't be called with nil
		},
		{
			name:     "completely empty client",
			client:   &spur.Client{},
			expected: true,
		},
		{
			name: "client with behaviors",
			client: &spur.Client{
				Behaviors: []string{"AUTOMATED"},
			},
			expected: false,
		},
		{
			name: "client with count",
			client: &spur.Client{
				Count: 5,
			},
			expected: false,
		},
		{
			name: "client with empty concentration",
			client: &spur.Client{
				Concentration: &spur.Concentration{},
			},
			expected: true,
		},
		{
			name: "client with non-empty concentration",
			client: &spur.Client{
				Concentration: &spur.Concentration{
					Country: "US",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.client == nil {
				// Skip nil test as the function assumes non-nil input
				return
			}
			result := isEmptyClient(tt.client)
			if result != tt.expected {
				t.Errorf("isEmptyClient() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
