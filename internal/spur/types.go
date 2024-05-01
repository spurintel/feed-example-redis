package spur

import (
	"errors"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"time"
)

// FeedType - enum for feed types: anonymous, anonymous-residential, ipsummary
type FeedType string

var ErrorNoV6Feed = errors.New("no v6 feed for this feed type")

const (
	AnonymousFeed            FeedType = "anonymous"
	AnonymousFeedIPV6        FeedType = "anonymous-ipv6"
	AnonymousResidential     FeedType = "anonymous-residential"
	AnonymousResidentialIPv6 FeedType = "anonymous-residential-ipv6"
	IPSummaryFeed            FeedType = "ipsummary"
	FeedTypeUnknown          FeedType = "unknown"
)

func (ft FeedType) V6FeedType() (FeedType, error) {
	switch ft {
	case AnonymousFeed:
		return AnonymousFeedIPV6, nil
	case AnonymousResidential:
		return AnonymousResidentialIPv6, nil
	default:
		return FeedTypeUnknown, ErrorNoV6Feed
	}
}

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
	Location       Location `json:"location,omitempty"`
	IP             string   `json:"ip,omitempty"`
	Organization   string   `json:"organization,omitempty"`
	Infrastructure string   `json:"infrastructure,omitempty"`
	Tunnels        []Tunnel `json:"tunnels,omitempty"`
	Services       []string `json:"services,omitempty"`
	Risks          []string `json:"risks,omitempty"`
	AS             AS       `json:"as,omitempty"`
	Client         Client   `json:"client,omitempty"`
}

type IPContextV6 struct {
	Location       Location `json:"location,omitempty" maxminddb:"location"`
	Network        string   `json:"network,omitempty" maxminddb:"network"`
	Organization   string   `json:"organization,omitempty" maxminddb:"organization"`
	Infrastructure string   `json:"infrastructure,omitempty" maxminddb:"infrastructure"`
	Tunnels        []Tunnel `json:"tunnels,omitempty" maxminddb:"tunnels"`
	Services       []string `json:"services,omitempty" maxminddb:"services"`
	Risks          []string `json:"risks,omitempty" maxminddb:"risks"`
	AS             AS       `json:"as,omitempty" maxminddb:"as"`
	Client         Client   `json:"client,omitempty" maxminddb:"client"`
}

type AS struct {
	Organization string `json:"organization,omitempty" maxminddb:"organization"`
	Number       int    `json:"number,omitempty" maxminddb:"number"`
}

type Client struct {
	Behaviors     []string `json:"behaviors,omitempty" maxminddb:"behaviors"`
	Types         []string `json:"types,omitempty" maxminddb:"types"`
	Proxies       []string `json:"proxies,omitempty" maxminddb:"proxies"`
	Concentration struct {
		Country string  `json:"country,omitempty" maxminddb:"country"`
		State   string  `json:"state,omitempty" maxminddb:"state"`
		City    string  `json:"city,omitempty" maxminddb:"city"`
		Geohash string  `json:"geohash,omitempty" maxminddb:"geohash"`
		Density float64 `json:"density,omitempty" maxminddb:"density"`
		Skew    int     `json:"skew,omitempty" maxminddb:"skew"`
	} `json:"concentration,omitempty" maxminddb:"concentration"`
	Countries int `json:"countries,omitempty" maxminddb:"countries"`
	Spread    int `json:"spread,omitempty" maxminddb:"spread"`
	Count     int `json:"count,omitempty" maxminddb:"count"`
}

type Location struct {
	Country string `json:"country,omitempty" maxminddb:"country"`
	State   string `json:"state,omitempty" maxminddb:"state"`
	City    string `json:"city,omitempty" maxminddb:"city"`
}

type Tunnel struct {
	Operator  string   `json:"operator,omitempty" maxminddb:"operator"`
	Type      string   `json:"type,omitempty" maxminddb:"type"`
	Entries   []string `json:"entries,omitempty" maxminddb:"entries"`
	Exits     []string `json:"exits,omitempty" maxminddb:"exits"`
	Anonymous bool     `json:"anonymous,omitempty" maxminddb:"anonymous"`
}

func (ipCtx IPContextV6) ToMMDB() mmdbtype.Map {
	record := mmdbtype.Map{}
	record["location"] = mmdbtype.Map{
		"country": mmdbtype.String(ipCtx.Location.Country),
		"city":    mmdbtype.String(ipCtx.Location.City),
		"state":   mmdbtype.String(ipCtx.Location.State),
	}
	record["network"] = mmdbtype.String(ipCtx.Network)
	record["organization"] = mmdbtype.String(ipCtx.Organization)
	record["infrastructure"] = mmdbtype.String(ipCtx.Infrastructure)

	tunnels := mmdbtype.Slice{}
	for _, t := range ipCtx.Tunnels {
		entries := mmdbtype.Slice{}
		exits := mmdbtype.Slice{}
		for _, e := range t.Entries {
			entries = append(entries, mmdbtype.String(e))
		}
		for _, e := range t.Exits {
			exits = append(exits, mmdbtype.String(e))
		}
		tunnels = append(tunnels, mmdbtype.Map{
			"operator":  mmdbtype.String(t.Operator),
			"type":      mmdbtype.String(t.Type),
			"entries":   entries,
			"exits":     exits,
			"anonymous": mmdbtype.Bool(t.Anonymous),
		})
	}
	record["tunnels"] = tunnels

	services := mmdbtype.Slice{}
	for _, s := range ipCtx.Services {
		services = append(services, mmdbtype.String(s))
	}
	record["services"] = services

	risks := mmdbtype.Slice{}
	for _, r := range ipCtx.Risks {
		risks = append(risks, mmdbtype.String(r))
	}
	record["risks"] = risks

	record["as"] = mmdbtype.Map{
		"organization": mmdbtype.String(ipCtx.AS.Organization),
		"number":       mmdbtype.Uint64(uint64(ipCtx.AS.Number)),
	}

	clientBehaviors := mmdbtype.Slice{}
	for _, b := range ipCtx.Client.Behaviors {
		clientBehaviors = append(clientBehaviors, mmdbtype.String(b))
	}

	clientTypes := mmdbtype.Slice{}
	for _, t := range ipCtx.Client.Types {
		clientTypes = append(clientTypes, mmdbtype.String(t))
	}

	clientProxies := mmdbtype.Slice{}
	for _, p := range ipCtx.Client.Proxies {
		clientProxies = append(clientProxies, mmdbtype.String(p))
	}

	record["client"] = mmdbtype.Map{
		"behaviors": clientBehaviors,
		"types":     clientTypes,
		"proxies":   clientProxies,
		"concentration": mmdbtype.Map{
			"country": mmdbtype.String(ipCtx.Client.Concentration.Country),
			"state":   mmdbtype.String(ipCtx.Client.Concentration.State),
			"city":    mmdbtype.String(ipCtx.Client.Concentration.City),
			"geohash": mmdbtype.String(ipCtx.Client.Concentration.Geohash),
			"density": mmdbtype.Float64(ipCtx.Client.Concentration.Density),
			"skew":    mmdbtype.Int32(ipCtx.Client.Concentration.Skew),
		},
		"countries": mmdbtype.Int32(ipCtx.Client.Countries),
		"spread":    mmdbtype.Int32(ipCtx.Client.Spread),
		"count":     mmdbtype.Int32(ipCtx.Client.Count),
	}

	return record
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
