package storage

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"feedexampleredis/internal/spur"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync/atomic"

	"github.com/maxmind/mmdbwriter"
	maxminddb "github.com/oschwald/maxminddb-golang"
)

var ErrorIPNotFound = fmt.Errorf("IP not found")

type MMDB struct {
	mmdb         *atomic.Pointer[maxminddb.Reader]
	lastFeedInfo *spur.FeedInfo
}

type ipv6record struct {
	RawJSON string `json:"json" mmdb:"json"`
}

func NewMMDB() *MMDB {
	mmdbPtr := atomic.Pointer[maxminddb.Reader]{}
	return &MMDB{
		mmdb: &mmdbPtr,
	}
}

// GetLastFeedInfo returns the last feed info
func (m *MMDB) GetLastFeedInfo() *spur.FeedInfo {
	return m.lastFeedInfo
}

// SetLastFeedInfo sets the last feed info
func (m *MMDB) SetLastFeedInfo(fi *spur.FeedInfo) {
	m.lastFeedInfo = fi
}

// StreamingFeedInsert inserts a feed into the MMDB
func (m *MMDB) StreamingFeedInsert(ctx context.Context, rc io.ReadCloser) (int64, error) {
	defer rc.Close()

	// Create a new mmdb writer
	writer, err := mmdbwriter.New(
		mmdbwriter.Options{
			DatabaseType: "Spur-IP-Context-V6",
			RecordSize:   32,
		},
	)

	if err != nil {
		return 0, nil
	}

	// Feed donwloads are gzipped, so we need to decompress them
	gzr, err := gzip.NewReader(rc)
	if err != nil {
		return 0, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Read the feed input line by line, each line is a JSON object which can be parsed into a *spur.IPContextV6
	scanner := bufio.NewScanner(gzr)
	scanBuf := make([]byte, 64*1024)   // 64KB buffer
	scanner.Buffer(scanBuf, 1024*1024) // 1MB maximum token size
	var count int64
	for scanner.Scan() {
		var ipCtx spur.IPContextV6
		raw := scanner.Bytes()
		err := json.Unmarshal(raw, &ipCtx)
		if err != nil {
			slog.Warn("error unmarshalling IP context", "error", err.Error())
			continue
		}

		_, network, err := net.ParseCIDR(ipCtx.Network)
		if err != nil {
			slog.Warn("error parsing network", "error", err.Error())
			continue
		}

		// Create a record from the IPContextV6
		record := ipCtx.ToMMDB()

		// Write the record to the mmdb
		err = writer.Insert(network, record)
		if err != nil {
			slog.Warn("error inserting record into mmdb", "error", err.Error())
			continue
		}

		count++
	}

	if err := scanner.Err(); err != nil {
		return count, err
	}

	// Write the mmdb to a byte slice
	buf := bytes.NewBuffer(nil)
	_, err = writer.WriteTo(buf)
	if err != nil {
		return count, err
	}

	// Create a new reader from the byte slice
	reader, err := maxminddb.FromBytes(buf.Bytes())
	if err != nil {
		return count, err
	}

	// Swap the reader into the atomic pointer
	m.mmdb.Store(reader)

	return count, nil
}

// Get looks up an IP in the MMDB
func (m *MMDB) Get(ip string) (*spur.IPContextV6, error) {
	netIP := net.ParseIP(ip)
	if netIP == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}
	return m.GetIP(netIP)
}

// GetIP looks up an IP in the MMDB
func (m *MMDB) GetIP(ip net.IP) (*spur.IPContextV6, error) {
	var record spur.IPContextV6
	db := m.mmdb.Load()
	err := db.Lookup(ip, &record)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup IP: %w", err)
	}

	if record.Network == "" {
		return nil, ErrorIPNotFound
	}

	return &record, err
}

// GetIPGeoByIP looks up an IP in the ipgeo MMDB and returns geolocation data
func (m *MMDB) GetIPGeoByIP(ip net.IP) (*spur.IPGeoMMDBResponse, error) {
	var record spur.IPGeoMMDBResponse
	db := m.mmdb.Load()
	if db == nil {
		return nil, fmt.Errorf("ipgeo MMDB not loaded")
	}

	err := db.Lookup(ip, &record)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup IP in ipgeo MMDB: %w", err)
	}

	// Check if we have valid data
	if record.Spur.AS.Number == 0 && record.Country.ISOCode == "" {
		return nil, ErrorIPNotFound
	}

	return &record, nil
}

// GetIPGeoByIPString looks up an IP string in the ipgeo MMDB and returns geolocation data
func (m *MMDB) GetIPGeoByIPString(ipStr string) (*spur.IPGeoMMDBResponse, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipStr)
	}
	return m.GetIPGeoByIP(ip)
}

// LoadIPGeoFromReader loads ipgeo MMDB data from a reader
func (m *MMDB) LoadIPGeoFromReader(reader *maxminddb.Reader) error {
	if reader == nil {
		return fmt.Errorf("reader cannot be nil")
	}

	// Store the reader
	m.mmdb.Store(reader)

	return nil
}

// GetByIP looks up an IP in the ipgeo MMDB and returns an IPContext (for fallback usage)
func (m *MMDB) GetByIP(ctx context.Context, ipStr string) (*spur.IPContext, error) {
	geoRecord, err := m.GetIPGeoByIPString(ipStr)
	if err != nil {
		return nil, err
	}

	// Convert the ipgeo response to IPContext
	ipContext := geoRecord.ToIPContext(ipStr)
	return ipContext, nil
}
