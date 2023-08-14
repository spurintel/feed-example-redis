package spur

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
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
	req, err := api.constructSpurHttpRequest(ctx, latestFeedInfoUrl(api.BaseURL, api.Version, string(feedType)))
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

	return &feedInfo, nil
}

func (api *API) LatestFeed(ctx context.Context, feedType FeedType) (io.ReadCloser, error) {
	req, err := api.constructSpurHttpRequest(ctx, latestFeedUrl(api.BaseURL, api.Version, string(feedType)))
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

func (api *API) LatestRealtimeFeedInfo(ctx context.Context, feedType FeedType) (*RealtimeFeedInfo, error) {
	url := latestRealtimeFeedInfoUrl(api.BaseURL, api.Version, string(feedType))
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

	return &feedInfo, nil
}

func (api *API) LatestRealtimeFeed(ctx context.Context, feedType FeedType) (io.ReadCloser, error) {
	url := latestRealtimeFeedUrl(api.BaseURL, api.Version, string(feedType))
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
