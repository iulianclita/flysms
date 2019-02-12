package sms

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://rest.messagebird.com"

// Client sends requests to the SMS API
type Client struct {
	accessKey  string
	baseURL    string
	httpClient *http.Client
}

// Options is a collection of client options
type Options struct {
	AccessKey string
	BaseURL   string
	Timeout   time.Duration
}

// NewClient creates a new client from the given options
func NewClient(opts Options) *Client {
	return &Client{
		accessKey: opts.AccessKey,
		baseURL:   opts.BaseURL,
		httpClient: &http.Client{
			Timeout: opts.Timeout,
		},
	}
}

// URL computes the full path using the base URL
func (c *Client) URL(path string) string {
	if c.baseURL == "" {
		c.baseURL = defaultBaseURL
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return fmt.Sprintf("%s%s", c.baseURL, path)
}
