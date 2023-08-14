package spur

import (
	"time"
)

// FeedType - enum for feed types: anonymous, anonymous-residential, ipsummary
type FeedType string

const (
	AnonymousFeed        FeedType = "anonymous"
	AnonymousResidential FeedType = "anonymous-residential"
	IPSummaryFeed        FeedType = "ipsummary"
)

// API - struct for spur api configuration
type API struct {
	BaseURL string
	Version string
	Token   string
}

// FeedInfo - struct for latest feed info
type FeedInfo struct {
	JSON struct {
		Location    string    `json:"location"`
		Date        string    `json:"date"`
		GeneratedAt time.Time `json:"generated_at"`
		AvailableAt time.Time `json:"available_at"`
	} `json:"json"`
}

// RealtimeFeedInfo - struct for latest realtime feed info
type RealtimeFeedInfo struct {
	JSON struct {
		Date     time.Time `json:"date"`
		Location string    `json:"location"`
	} `json:"json"`
}

type FeedError struct {
	Err string `json:"error"`
}

func (e *FeedError) Error() string {
	return e.Err
}

type IPContext struct {
	Location       Location `json:"location"`
	IP             string   `json:"ip"`
	Organization   string   `json:"organization"`
	Infrastructure string   `json:"infrastructure"`
	Tunnels        []Tunnel `json:"tunnels"`
	Services       []string `json:"services"`
	Risks          []string `json:"risks"`
	AS             AS       `json:"as"`
	Client         Client   `json:"client"`
}

type AS struct {
	Organization string `json:"organization"`
	Number       int    `json:"number"`
}

type Client struct {
	Behaviors     []string `json:"behaviors"`
	Types         []string `json:"types"`
	Proxies       []string `json:"proxies"`
	Concentration struct {
		Country string  `json:"country"`
		State   string  `json:"state"`
		City    string  `json:"city"`
		Geohash string  `json:"geohash"`
		Density float64 `json:"density"`
		Skew    int     `json:"skew"`
	} `json:"concentration"`
	Countries int `json:"countries"`
	Spread    int `json:"spread"`
	Count     int `json:"count"`
}

type Location struct {
	Country string `json:"country"`
	State   string `json:"state"`
	City    string `json:"city"`
}

type Tunnel struct {
	Operator  string   `json:"operator"`
	Type      string   `json:"type"`
	Entries   []string `json:"entries"`
	Exits     []string `json:"exits"`
	Anonymous bool     `json:"anonymous"`
}

// Deep merging for each struct
func (as *AS) merge(other *AS) {
	if other.Number != 0 {
		as.Number = other.Number
	}
	if other.Organization != "" {
		as.Organization = other.Organization
	}
}

func (client *Client) merge(other *Client) {
	client.Behaviors = mergeUniqueSlices(client.Behaviors, other.Behaviors)
	client.Concentration.Country = takeNewerIfNotEmpty(client.Concentration.Country, other.Concentration.Country)
	client.Concentration.State = takeNewerIfNotEmpty(client.Concentration.State, other.Concentration.State)
	client.Concentration.City = takeNewerIfNotEmpty(client.Concentration.City, other.Concentration.City)
	client.Concentration.Geohash = takeNewerIfNotEmpty(client.Concentration.Geohash, other.Concentration.Geohash)
	client.Concentration.Density = takeNewerIfNotEmpty(client.Concentration.Density, other.Concentration.Density)
	client.Concentration.Skew = takeNewerIfNotEmpty(client.Concentration.Skew, other.Concentration.Skew)
	client.Countries = takeNewerIfNotEmpty(client.Countries, other.Countries)
	client.Spread = takeNewerIfNotEmpty(client.Spread, other.Spread)
	client.Proxies = mergeUniqueSlices(client.Proxies, other.Proxies)
	client.Count = takeNewerIfNotEmpty(client.Count, other.Count)
	client.Types = mergeUniqueSlices(client.Types, other.Types)
}

func (location *Location) merge(other *Location) {
	location.Country = takeNewerIfNotEmpty(location.Country, other.Country)
	location.State = takeNewerIfNotEmpty(location.State, other.State)
	location.City = takeNewerIfNotEmpty(location.City, other.City)
}

func (ipContext *IPContext) Merge(other *IPContext) {
	ipContext.IP = takeNewerIfNotEmpty(ipContext.IP, other.IP)
	ipContext.AS.merge(&other.AS)
	ipContext.Organization = takeNewerIfNotEmpty(ipContext.Organization, other.Organization)
	ipContext.Infrastructure = takeNewerIfNotEmpty(ipContext.Infrastructure, other.Infrastructure)
	ipContext.Client.merge(&other.Client)
	ipContext.Location.merge(&other.Location)
	ipContext.Services = mergeUniqueSlices(ipContext.Services, other.Services)
	ipContext.Risks = mergeUniqueSlices(ipContext.Risks, other.Risks)
	ipContext.Tunnels = mergeTunnels(ipContext.Tunnels, other.Tunnels)
}

func takeNewerIfNotEmpty[K comparable](k1, k2 K) K {
	var zero K
	if k2 != zero {
		return k2
	}

	return k1
}

func mergeUniqueSlices[K comparable](s1, s2 []K) []K {
	merged := s1
	for _, k := range s2 {
		exists := false
		for _, m := range s1 {
			if k == m {
				exists = true
				break
			}
		}
		if !exists {
			merged = append(merged, k)
		}
	}
	return merged
}

func mergeTunnels(t1, t2 []Tunnel) []Tunnel {
	// Set to original list
	merged := t1
	for _, tn := range t2 {
		exists := false
		for _, tm := range t1 {
			if tn.Operator == tm.Operator && tm.Operator != "" && tn.Operator != "" {
				exists = true
				break
			}
		}
		if !exists {
			merged = append(merged, tn)
		}
	}
	return merged
}
