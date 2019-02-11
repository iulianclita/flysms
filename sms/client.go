package sms

import "net/http"

// Client sends requests to the SMS API
type Client struct {
	AccessKey  string
	BaseURL    string
	HTTPClient *http.Client
}
