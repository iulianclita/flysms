package sms

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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

// MessageCreated is the API mapping for a succesfully created message
type MessageCreated struct {
	ID              string            `json:"id"`
	Originator      string            `json:"originator"`
	Body            string            `json:"body"`
	Recipients      MessageRecipients `json:"recipients"`
	CreatedDateTime string            `json:"createdDatetime"`
}

// MessageRecipients contains relevant information about every recipient
type MessageRecipients struct {
	Items []MessageItem `json:"items"`
}

// MessageItem containts relevant information for a given recipient
type MessageItem struct {
	Recipient      int64  `json:"recipient"`
	Status         string `json:"status"`
	StatusDateTime string `json:"statusDatetime"`
}

// MessageErrors is the errors bag API response for a failed create message action
type MessageErrors struct {
	Errors []MessageError `json:"errors"`
}

// MessageError represents every error in the bag
type MessageError struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
	Parameter   string `json:"parameter"`
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

// createMessage sends the API request to messagebird
func (c *Client) createMessage(r *Request) (interface{}, int, error) {
	v := url.Values{}
	v.Set("recipients", fmt.Sprintf("%d", r.Recipient))
	v.Set("originator", r.Originator)
	v.Set("body", r.Message)

	endpoint := c.URL("messages")
	payload := strings.NewReader(v.Encode())

	req, err := http.NewRequest(http.MethodPost, endpoint, payload)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("Cannot create POST request for url %s; Error: %v", endpoint, err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("AccessKey %s", c.accessKey))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("Cannot get response for request %#v; Error: %v", req, err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("Cannot read response body %#v; Error: %v", res, err)
	}
	defer res.Body.Close()

	var data interface{}
	var msgSuccess MessageCreated
	var msgFail MessageErrors

	if err := json.Unmarshal(body, &msgSuccess); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("Failed to unmarshal body into JSON %s; Error: %v", string(body), err)
	}

	if msgSuccess.ID != "" {
		data = msgSuccess
		return data, res.StatusCode, nil
	}

	if err := json.Unmarshal(body, &msgFail); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("Failed to unmarshal body into JSON %s; Error: %v", string(body), err)
	}

	data = msgFail

	return data, res.StatusCode, nil
}
