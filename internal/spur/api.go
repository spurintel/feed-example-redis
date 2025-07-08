package spur

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/oschwald/maxminddb-golang"
)

// NewAPI - create new API struct
func NewAPI(baseURL, version, token string) *API {
	return &API{
		BaseURL: baseURL,
		Version: version,
		Token:   token,
	}
}

func (api *API) LatestFeedInfo(ctx context.Context, feedType FeedType) (*FeedInfo, error) {
	url := latestFeedInfoUrl(api.BaseURL, api.Version, string(feedType))
	slog.Info("getting latest feed info", slog.String("url", url))
	req, err := api.constructSpurHttpRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		var feedError FeedError
		err = json.NewDecoder(r.Body).Decode(&feedError)
		if err != nil {
			return nil, err
		}

		return nil, &feedError
	}

	var feedInfo FeedInfo
	err = json.NewDecoder(r.Body).Decode(&feedInfo)
	if err != nil {
		return nil, err
	}

	slog.Info(
		"latest feed info",
		slog.String("feed_type", string(feedType)),
		slog.String("date", feedInfo.JSON.Date),
		slog.String("available_at", feedInfo.JSON.AvailableAt.Format(time.RFC3339)),
		slog.String("expires_at", feedInfo.JSON.Date),
		slog.String("location", feedInfo.JSON.Location),
	)

	return &feedInfo, nil
}

func (api *API) LatestFeed(ctx context.Context, feedType FeedType) (io.ReadCloser, error) {
	url := latestFeedUrl(api.BaseURL, api.Version, string(feedType))
	slog.Info("getting latest feed", slog.String("url", url))
	req, err := api.constructSpurHttpRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		var feedError FeedError
		err = json.NewDecoder(r.Body).Decode(&feedError)
		if err != nil {
			return nil, err
		}

		return nil, &feedError
	}

	return r.Body, nil
}

func (api *API) LatestGeoFeedMMDB(ctx context.Context, feedType FeedType) (*maxminddb.Reader, error) {
	url := latestGeoFeedUrl(api.BaseURL, api.Version, string(feedType))

	// Construct the request manually to avoid setting Accept header
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Token", api.Token)

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		var feedError FeedError
		err = json.NewDecoder(r.Body).Decode(&feedError)
		if err != nil {
			return nil, err
		}
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}

	// Write the body to a temporary file for inspection
	tmpFile, tmpErr := os.CreateTemp("", "spur-mmdb-*.mmdb")
	if tmpErr == nil {
		_, _ = tmpFile.Write(bodyBytes)
		tmpFile.Close()
		slog.Info("MMDB written to temp file", slog.String("path", tmpFile.Name()))
	} else {
		slog.Warn("Failed to write MMDB to temp file", slog.String("error", tmpErr.Error()))
	}

	mmdbReader, err := maxminddb.FromBytes(bodyBytes)
	if err != nil {
		return nil, err
	}

	return mmdbReader, nil
}

func (api *API) LatestRealtimeFeedInfo(ctx context.Context, feedType FeedType) (*RealtimeFeedInfo, error) {
	url := latestRealtimeFeedInfoUrl(api.BaseURL, api.Version, string(feedType))
	slog.Info("getting latest realtime feed info", slog.String("url", url))
	req, err := api.constructSpurHttpRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		var feedError FeedError
		err = json.NewDecoder(r.Body).Decode(&feedError)
		if err != nil {
			return nil, err
		}

		return nil, &feedError
	}

	var feedInfo RealtimeFeedInfo
	err = json.NewDecoder(r.Body).Decode(&feedInfo)
	if err != nil {
		return nil, err
	}

	slog.Info(
		"latest feed info",
		slog.String("feed_type", string(feedType)),
		slog.String("date", feedInfo.JSON.Date.Format(time.RFC3339)),
		slog.String("location", feedInfo.JSON.Location),
	)

	return &feedInfo, nil
}

func (api *API) LatestRealtimeFeed(ctx context.Context, feedType FeedType) (io.ReadCloser, error) {
	url := latestRealtimeFeedUrl(api.BaseURL, api.Version, string(feedType))
	slog.Info("getting latest realtime feed", slog.String("url", url))
	req, err := api.constructSpurHttpRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		var feedError FeedError
		err = json.NewDecoder(r.Body).Decode(&feedError)
		if err != nil {
			return nil, err
		}

		return nil, &feedError
	}

	return r.Body, nil
}

func (api *API) RealtimeFeed(ctx context.Context, feedType FeedType, t time.Time) (io.ReadCloser, error) {
	url := realtimeFeedUrl(api.BaseURL, api.Version, string(feedType), t)
	slog.Info("getting realtime feed", slog.String("url", url))
	req, err := api.constructSpurHttpRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		var feedError FeedError
		err = json.NewDecoder(r.Body).Decode(&feedError)
		if err != nil {
			return nil, err
		}

		return nil, &feedError
	}

	return r.Body, nil
}

func (api *API) constructSpurHttpRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Token", api.Token)
	req.Header.Add("Accept", "application/json")

	return req, nil
}

func latestFeedInfoUrl(baseURL, version, feed string) string {
	return constructFeedBaseURL(baseURL, version, feed) + "/latest"
}

func latestFeedUrl(baseURL, version, feed string) string {
	return constructFeedBaseURL(baseURL, version, feed) + "/latest.json.gz"
}

func latestGeoFeedUrl(baseURL, version, feed string) string {
	return constructFeedBaseURL(baseURL, version, feed) + "/latest.mmdb"
}

func latestRealtimeFeedUrl(baseURL, version, feed string) string {
	return constructFeedBaseURL(baseURL, version, feed) + "/realtime" + "/latest.json.gz"
}

func latestRealtimeFeedInfoUrl(baseURL, version, feed string) string {
	return constructFeedBaseURL(baseURL, version, feed) + "/realtime" + "/latest"
}

func realtimeFeedUrl(baseURL, version, feed string, t time.Time) string {
	date := t.Format("20060102")
	time := t.Format("1504")
	return constructFeedBaseURL(baseURL, version, feed) + "/realtime" + "/" + date + "/" + time + ".json.gz"
}

func constructFeedBaseURL(baseURL, version, feed string) string {
	return baseURL + "/" + version + "/" + feed
}
