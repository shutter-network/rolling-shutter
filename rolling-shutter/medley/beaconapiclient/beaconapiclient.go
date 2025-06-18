package beaconapiclient

import (
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

func (c *Client) GetBeaconNodeVersion() (string, error) {
	endpoint := "/eth/v1/node/version"
	resp, err := c.c.Get(c.url.JoinPath(endpoint).String())
	if err != nil {
		return "", err
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
