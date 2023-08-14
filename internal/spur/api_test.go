package spur

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLatestFeedInfo(t *testing.T) {
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
	info, err := api.LatestFeedInfo(context.Background(), AnonymousResidential)

	// Check that the function returns the expected values
	if err != nil {
		t.Errorf("GetFeedInfo returned an error: %v", err)
	}

	date := time.Now().UTC().Format("20060102")
	location := date + "/feed.json.gz"
	if info.JSON.Location != location {
		t.Fatalf("GetFeedInfo returned incorrect location %s, expected %s", info.JSON.Location, location)
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

func TestLatestFeed(t *testing.T) {
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
	body, err := api.LatestFeed(context.Background(), AnonymousResidential)
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
