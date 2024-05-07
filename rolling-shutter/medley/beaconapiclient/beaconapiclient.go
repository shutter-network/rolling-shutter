package beaconapiclient

import (
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
