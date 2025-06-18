package beaconapiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	c   *http.Client
	url *url.URL
}

func New(rawURL string) (*Client, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		c:   &http.Client{},
		url: parsedURL,
	}, nil
}

func (c *Client) GetBeaconNodeVersion(ctx context.Context) (string, error) {
	path := c.url.JoinPath("/eth/v1/node/version")
	req, err := http.NewRequestWithContext(ctx, "GET", path.String(), http.NoBody)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get beacon node version from consensus node: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var result struct {
		Data struct {
			Version string `json:"version"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}
	return result.Data.Version, nil
}
