package storage

import (
	"bytes"
	"compress/gzip"
	"context"
	"feedexampleredis/internal/spur"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"testing"
)

func createReadCloser() io.ReadCloser {
	data := `{"network":"2001:1890:1aec:3000::/56","organization":"HYPESTATUS INC","as":{"number":7018,"organization":"AT\u0026T Services, Inc."},"client":{},"tunnels":[{"operator":"HYPE_PROXY","type":"PROXY","anonymous":true}],"location":{"country":"US"},"risks":["TUNNEL"]}
{"network":"2a02:26f7:d198:e068::/64","organization":"Akamai International B.V.","as":{"number":20940,"organization":"Akamai International B.V."},"client":{},"tunnels":[{"operator":"ICLOUD_RELAY_PROXY","type":"PROXY","anonymous":true}],"infrastructure":"DATACENTER","location":{"country":"FI"},"risks":["TUNNEL"]}
`

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(data)); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}

	return io.NopCloser(&buf)
}

func TestStreamingFeedInsert(t *testing.T) {
	type args struct {
		ctx   context.Context
		input io.ReadCloser
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test feed insert",
			args: args{
				ctx:   context.Background(),
				input: createReadCloser(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mmdb := NewMMDB()
			_, err := mmdb.StreamingFeedInsert(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("StreamingFeedInsert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGet(t *testing.T) {
	mmdb := NewMMDB()

	// Insert the test data
	_, err := mmdb.StreamingFeedInsert(context.Background(), createReadCloser())
	if err != nil {
		t.Fatalf("Failed to insert feed: %v", err)
	}

	tests := []struct {
		name        string
		ip          string
		wantContext spur.IPContextV6
		wantErr     bool
	}{
		{
			name:    "Test random IPv6",
			ip:      "2001:db8::2:1",
			wantErr: true,
		},
		{
			name: "Test network 1",
			ip:   "2001:1890:1aec:3000::1",
			wantContext: spur.IPContextV6{
				Network:      "2001:1890:1aec:3000::/56",
				Organization: "HYPESTATUS INC",
				AS: spur.AS{
					Number:       7018,
					Organization: "AT&T Services, Inc.",
				},
				Client: spur.Client{},
				Tunnels: []spur.Tunnel{
					{
						Operator:  "HYPE_PROXY",
						Type:      "PROXY",
						Anonymous: true,
					},
				},
				Location: spur.Location{
					Country: "US",
				},
				Risks: []string{"TUNNEL"},
			},
		},
		{
			name:    "Test network 2",
			ip:      "2a02:26f7:d198:e068::1",
			wantErr: false,
			wantContext: spur.IPContextV6{
				Network:      "2a02:26f7:d198:e068::/64",
				Organization: "Akamai International B.V.",
				AS: spur.AS{
					Number:       20940,
					Organization: "Akamai International B.V.",
				},
				Client: spur.Client{},
				Tunnels: []spur.Tunnel{
					{
						Operator:  "ICLOUD_RELAY_PROXY",
						Type:      "PROXY",
						Anonymous: true,
					},
				},
				Infrastructure: "DATACENTER",
				Location: spur.Location{
					Country: "FI",
				},
				Risks: []string{"TUNNEL"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mmdb.Get(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetIP(t *testing.T) {
	mmdb := NewMMDB()

	// Insert the test data
	_, err := mmdb.StreamingFeedInsert(context.Background(), createReadCloser())
	if err != nil {
		t.Fatalf("Failed to insert feed: %v", err)
	}

	tests := []struct {
		name        string
		ip          net.IP
		wantContext spur.IPContextV6
		wantErr     bool
	}{
		{
			name:    "Test random IPv6",
			ip:      net.ParseIP("2001:db8::2:1"),
			wantErr: true,
		},
		{
			name:    "Test network 1",
			ip:      net.ParseIP("2001:1890:1aec:3000::1"),
			wantErr: false,
			wantContext: spur.IPContextV6{
				Network:      "2001:1890:1aec:3000::/56",
				Organization: "HYPESTATUS INC",
				AS: spur.AS{
					Number:       7018,
					Organization: "AT&T Services, Inc.",
				},
				Client: spur.Client{},
				Tunnels: []spur.Tunnel{
					{
						Operator:  "HYPE_PROXY",
						Type:      "PROXY",
						Anonymous: true,
					},
				},
				Location: spur.Location{
					Country: "US",
				},
				Risks: []string{"TUNNEL"},
			},
		},
		{
			name:    "Test network 2",
			ip:      net.ParseIP("2a02:26f7:d198:e068::1"),
			wantErr: false,
			wantContext: spur.IPContextV6{
				Network:      "2a02:26f7:d198:e068::/64",
				Organization: "Akamai International B.V.",
				AS: spur.AS{
					Number:       20940,
					Organization: "Akamai International B.V.",
				},
				Client: spur.Client{},
				Tunnels: []spur.Tunnel{
					{
						Operator:  "ICLOUD_RELAY_PROXY",
						Type:      "PROXY",
						Anonymous: true,
					},
				},
				Infrastructure: "DATACENTER",
				Location: spur.Location{
					Country: "FI",
				},
				Risks: []string{"TUNNEL"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mmdb.GetIP(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNetworks(t *testing.T) {
	mmdb := NewMMDB()

	// Insert the test data
	_, err := mmdb.StreamingFeedInsert(context.Background(), createReadCloser())
	if err != nil {
		t.Fatalf("Failed to insert feed: %v", err)
	}

	// Get the internal reader
	reader := mmdb.mmdb.Load()

	expected := []string{
		"2001:1890:1aec:3000::/56",
		"2a02:26f7:d198:e068::/64",
	}

	// Iterate over the networks in the reader
	var innerIPs []string
	for _, n := range expected {
		_, network, err := net.ParseCIDR(n)
		assert.Nil(t, err)
		n := reader.NetworksWithin(network)
		for n.Next() {
			record := struct {
				IP string `maxminddb:"ip"`
			}{}
			network, err := n.Network(&record)
			assert.Nil(t, err)
			innerIPs = append(innerIPs, network.String())
		}

		assert.Nil(t, n.Err())
	}

	t.Log(innerIPs)
	assert.Equal(t, expected, innerIPs)
}
