package spur

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAPI_LatestFeedInfo(t *testing.T) {
	token := os.Getenv("SPUR_REDIS_API_TOKEN")
	if token == "" {
		t.Fatal("SPUR_REDIS_API_TOKEN is not set")
	}

	type args struct {
		ctx      context.Context
		feedType FeedType
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test Anonymous Feed",
			args: args{
				ctx:      context.Background(),
				feedType: AnonymousFeed,
			},
			wantErr: false,
		},
		{
			name: "Test Anonymous Feed IPV6",
			args: args{
				ctx:      context.Background(),
				feedType: AnonymousFeedIPV6,
			},
			wantErr: false,
		},
		{
			name: "Test Anonymous Residential",
			args: args{
				ctx:      context.Background(),
				feedType: AnonymousResidential,
			},
			wantErr: false,
		},
		{
			name: "Test Anonymous Residential IPV6",
			args: args{
				ctx:      context.Background(),
				feedType: AnonymousResidentialIPv6,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &API{
				BaseURL: "https://feeds.spur.us",
				Version: "v2",
				Token:   token,
			}

			// Get the latest feed info
			info, err := api.LatestFeedInfo(tt.args.ctx, tt.args.feedType)
			if (err != nil) != tt.wantErr {
				t.Errorf("LatestFeedInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check that the function returns the expected values
			if err != nil {
				t.Errorf("GetFeedInfo returned an error: %v", err)
			}

			date := time.Now().UTC().Format("20060102")
			location := date + "/feed.json.gz"
			if info.JSON.Location != location {
				t.Fatalf("GetFeedInfo returned incorrect location %s, expected %s", info.JSON.Location, location)
			}
		})
	}
}

func TestLatestRealtimeFeedInfo(t *testing.T) {
	token := os.Getenv("SPUR_REDIS_API_TOKEN")
	if token == "" {
		t.Fatal("SPUR_REDIS_API_TOKEN is not set")
	}

	// Set up the API client with the test server URL and a mock token
	api := API{
		BaseURL: "https://feeds.spur.us",
		Version: "v2",
		Token:   token,
	}

	// Call the function being tested
	info, err := api.LatestRealtimeFeedInfo(context.Background(), AnonymousResidential)

	// Check that the function returns the expected values
	if err != nil {
		t.Errorf("GetRealtimeFeedInfo returned an error: %v", err)
	}

	date := time.Now().UTC().Format("20060102")
	location := "realtime/" + date
	if !strings.HasPrefix(info.JSON.Location, location) {
		t.Errorf("GetFeedInfo returned incorrect location %s, expected %s", info.JSON.Location, location)
	}
}

func TestAPI_LatestFeed(t *testing.T) {
	token := os.Getenv("SPUR_REDIS_API_TOKEN")
	if token == "" {
		t.Fatal("SPUR_REDIS_API_TOKEN is not set")
	}

	type args struct {
		ctx      context.Context
		feedType FeedType
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test Anonymous Feed",
			args: args{
				ctx:      context.Background(),
				feedType: AnonymousFeed,
			},
			wantErr: false,
		},
		{
			name: "Test Anonymous Feed IPV6",
			args: args{
				ctx:      context.Background(),
				feedType: AnonymousFeedIPV6,
			},
			wantErr: false,
		},
		{
			name: "Test Anonymous Residential",
			args: args{
				ctx:      context.Background(),
				feedType: AnonymousResidential,
			},
			wantErr: false,
		},
		{
			name: "Test Anonymous Residential IPV6",
			args: args{
				ctx:      context.Background(),
				feedType: AnonymousResidentialIPv6,
			},
			wantErr: false,
		},
		// this is big and takes too long so skipping it
		//{
		//	name: "Test IP Summary Feed",
		//	args: args{
		//		ctx:      context.Background(),
		//		feedType: IPSummaryFeed,
		//	},
		//	wantErr: false,
		//},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &API{
				BaseURL: "https://feeds.spur.us",
				Version: "v2",
				Token:   token,
			}

			body, err := api.LatestFeed(tt.args.ctx, tt.args.feedType)
			if (err != nil) != tt.wantErr {
				t.Errorf("LatestFeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Discard the response body
			size, err := io.Copy(io.Discard, body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			if size == 0 {
				t.Fatalf("Response body was empty")
			}
		})
	}
}

func TestLatestRealtimeFeed(t *testing.T) {
	token := os.Getenv("SPUR_REDIS_API_TOKEN")
	if token == "" {
		t.Fatal("SPUR_REDIS_API_TOKEN is not set")
	}

	// Set up the API client with the test server URL and a mock token
	api := API{
		BaseURL: "https://feeds.spur.us",
		Version: "v2",
		Token:   token,
	}

	// Call the function being tested
	body, err := api.LatestRealtimeFeed(context.Background(), AnonymousResidential)
	if err != nil {
		t.Fatalf("LatestFeed returned an error: %v", err)
	}

	// Discard the response body
	size, err := io.Copy(io.Discard, body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if size == 0 {
		t.Fatalf("Response body was empty")
	}
}

func TestRealtimeFeed(t *testing.T) {
	token := os.Getenv("SPUR_REDIS_API_TOKEN")
	if token == "" {
		t.Fatal("SPUR_REDIS_API_TOKEN is not set")
	}

	// Set up the API client with the test server URL and a mock token
	api := API{
		BaseURL: "https://feeds.spur.us",
		Version: "v2",
		Token:   token,
	}

	// Call the function being tested
	now := time.Now().UTC()
	lastFiveMin := now.Add(-5 * time.Minute)
	roundedFiveMin := lastFiveMin.Round(5 * time.Minute)
	body, err := api.RealtimeFeed(context.Background(), AnonymousResidential, roundedFiveMin)
	if err != nil {
		t.Fatalf("LatestFeed returned an error: %v", err)
	}

	// Discard the response body
	size, err := io.Copy(io.Discard, body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if size == 0 {
		t.Fatalf("Response body was empty")
	}
}

func TestAPI_LatestGeoFeedMMDB(t *testing.T) {
	token := os.Getenv("SPUR_REDIS_API_TOKEN")
	if token == "" {
		t.Fatal("SPUR_REDIS_API_TOKEN is not set")
	}

	api := &API{
		BaseURL: "https://feeds.spur.us",
		Version: "v2",
		Token:   token,
	}

	mmdb, err := api.LatestGeoFeedMMDB(context.Background(), IPGeo)
	if err != nil {
		t.Fatalf("LatestGeoFeedMMDB returned error: %v", err)
	}

	var record map[string]interface{}
	err = mmdb.Lookup(net.ParseIP("1.1.1.1"), &record)
	if err != nil {
		t.Fatalf("Lookup for 1.1.1.1 failed: %v", err)
	}

	t.Logf("1.1.1.1 record: %+v", record)
}

func TestIPGeoMMDBResponse_ToIPContext(t *testing.T) {
	// Test data based on the example JSON provided
	response := IPGeoMMDBResponse{
		City: struct {
			Names map[string]string `maxminddb:"names"`
		}{
			Names: map[string]string{
				"en": "Anycast",
				"de": "Anycast",
			},
		},
		Country: struct {
			ISOCode string            `maxminddb:"iso_code"`
			Names   map[string]string `maxminddb:"names"`
		}{
			ISOCode: "ZZ",
			Names: map[string]string{
				"en": "Anycast",
				"de": "Anycast",
			},
		},
		Spur: struct {
			AS struct {
				Number       int    `maxminddb:"number"`
				Organization string `maxminddb:"organization"`
			} `maxminddb:"as"`
			Infrastructure string `maxminddb:"infrastructure"`
		}{
			AS: struct {
				Number       int    `maxminddb:"number"`
				Organization string `maxminddb:"organization"`
			}{
				Number:       13335,
				Organization: "Cloudflare, Inc.",
			},
			Infrastructure: "DATACENTER",
		},
		Subdivisions: []struct {
			Names map[string]string `maxminddb:"names"`
		}{
			{
				Names: map[string]string{
					"en": "Anycast",
					"de": "Anycast",
				},
			},
		},
	}

	testIP := "1.1.1.1"
	ipContext := response.ToIPContext(testIP)

	// Verify the conversion
	if ipContext.IP != testIP {
		t.Errorf("Expected IP %s, got %s", testIP, ipContext.IP)
	}

	if ipContext.Location.City != "Anycast" {
		t.Errorf("Expected city 'Anycast', got %s", ipContext.Location.City)
	}

	if ipContext.Location.Country != "ZZ" {
		t.Errorf("Expected country 'ZZ', got %s", ipContext.Location.Country)
	}

	if ipContext.Location.State != "Anycast" {
		t.Errorf("Expected state 'Anycast', got %s", ipContext.Location.State)
	}

	if ipContext.AS.Number != 13335 {
		t.Errorf("Expected AS number 13335, got %d", ipContext.AS.Number)
	}

	if ipContext.AS.Organization != "Cloudflare, Inc." {
		t.Errorf("Expected AS organization 'Cloudflare, Inc.', got %s", ipContext.AS.Organization)
	}

	if ipContext.Infrastructure != "DATACENTER" {
		t.Errorf("Expected infrastructure 'DATACENTER', got %s", ipContext.Infrastructure)
	}

	// Test that Client field is properly nil
	if ipContext.Client != nil {
		t.Errorf("Expected Client to be nil, but got: %+v", ipContext.Client)
	}

	// Test JSON serialization to ensure client is omitted
	jsonData, err := json.Marshal(ipContext)
	if err != nil {
		t.Fatalf("Failed to marshal IPContext: %v", err)
	}

	jsonStr := string(jsonData)
	t.Logf("JSON output: %s", jsonStr)

	// Check if client field is present in JSON
	if strings.Contains(jsonStr, "client") {
		t.Errorf("Client field should be omitted from JSON, but found: %s", jsonStr)
	}

	t.Logf("Converted IPContext: %+v", ipContext)
}
