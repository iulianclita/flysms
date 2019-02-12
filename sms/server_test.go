package sms_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iulianclita/messagebird/sms"
)

func TestServer_createMessage(t *testing.T) {

	type wantType struct {
		statusCode int
		response   sms.Response
	}

	tests := map[string]struct {
		httpMethod    string
		path          string
		payload       io.Reader
		serverOptions sms.Config
		want          wantType
	}{
		"HTTP Method not allowed": {
			httpMethod: http.MethodGet,
			path:       "/messages",
			payload:    nil,
			serverOptions: sms.Config{
				Buffer:       10,
				ReqTimeout:   5 * time.Second,
				ThrottleRate: time.Second,
			},
			want: wantType{
				statusCode: http.StatusMethodNotAllowed,
				response: sms.Response{
					Success: false,
					Error:   "Request not allowed (invalid HTTP method)",
				},
			},
		},

		"Invalid JSON": {
			httpMethod: http.MethodPost,
			path:       "/messages",
			payload:    strings.NewReader(`{"invalid_json"}`),
			serverOptions: sms.Config{
				Buffer:       10,
				ReqTimeout:   5 * time.Second,
				ThrottleRate: time.Second,
			},
			want: wantType{
				statusCode: http.StatusBadRequest,
				response: sms.Response{
					Success: false,
					Error:   "Bad request (invalid payload json structure)",
				},
			},
		},

		"Invalid recipient value": {
			httpMethod: http.MethodPost,
			path:       "/messages",
			payload:    strings.NewReader(`{"recipient":123456, "originator": "MesssageBird", "message": "This is a test message"}`),
			serverOptions: sms.Config{
				Buffer:       10,
				ReqTimeout:   5 * time.Second,
				ThrottleRate: time.Second,
			},
			want: wantType{
				statusCode: http.StatusUnprocessableEntity,
				response: sms.Response{
					Success: false,
					Error:   "Invalid parameter (recipient value is out of bounds)",
				},
			},
		},

		"Missing originator value": {
			httpMethod: http.MethodPost,
			path:       "/messages",
			payload:    strings.NewReader(`{"recipient":1234567890, "originator": "", "message": "This is a test message"}`),
			serverOptions: sms.Config{
				Buffer:       10,
				ReqTimeout:   5 * time.Second,
				ThrottleRate: time.Second,
			},
			want: wantType{
				statusCode: http.StatusUnprocessableEntity,
				response: sms.Response{
					Success: false,
					Error:   "Missing parameter (originator value is not present)",
				},
			},
		},

		"Invalid originator value": {
			httpMethod: http.MethodPost,
			path:       "/messages",
			payload:    strings.NewReader(`{"recipient":1234567890, "originator": "VeryLongNameForThisOriginator", "message": "This is a test message"}`),
			serverOptions: sms.Config{
				Buffer:       10,
				ReqTimeout:   5 * time.Second,
				ThrottleRate: time.Second,
			},
			want: wantType{
				statusCode: http.StatusUnprocessableEntity,
				response: sms.Response{
					Success: false,
					Error:   "Invalid parameter (originator value is to long)",
				},
			},
		},

		"Missing message value": {
			httpMethod: http.MethodPost,
			path:       "/messages",
			payload:    strings.NewReader(`{"recipient":1234567890, "originator": "MessageBird", "message": ""}`),
			serverOptions: sms.Config{
				Buffer:       10,
				ReqTimeout:   5 * time.Second,
				ThrottleRate: time.Second,
			},
			want: wantType{
				statusCode: http.StatusUnprocessableEntity,
				response: sms.Response{
					Success: false,
					Error:   "Missing parameter (message value is not present)",
				},
			},
		},

		"Invalid message value": {
			httpMethod: http.MethodPost,
			path:       "/messages",
			payload:    strings.NewReader(fmt.Sprintf(`{"recipient":1234567890, "originator": "MessageBird", "message": "%s"}`, strings.Repeat("X", 161))),
			serverOptions: sms.Config{
				Buffer:       10,
				ReqTimeout:   5 * time.Second,
				ThrottleRate: time.Second,
			},
			want: wantType{
				statusCode: http.StatusUnprocessableEntity,
				response: sms.Response{
					Success: false,
					Error:   "Invalid parameter (message value is to long)",
				},
			},
		},

		"API Client not set": {
			httpMethod: http.MethodPost,
			path:       "/messages",
			payload:    strings.NewReader(`{"recipient":1234567890, "originator": "MessageBird", "message": "This is a test message"}`),
			serverOptions: sms.Config{
				Buffer:       10,
				ReqTimeout:   5 * time.Second,
				ThrottleRate: time.Second,
			},
			want: wantType{
				statusCode: http.StatusInternalServerError,
				response: sms.Response{
					Success: false,
					Error:   "Internal error (API client not set)",
				},
			},
		},

		"API request timeout": {
			httpMethod: http.MethodPost,
			path:       "/messages",
			payload:    strings.NewReader(`{"recipient":1234567890, "originator": "MessageBird", "message": "This is a test message"}`),
			serverOptions: sms.Config{
				Buffer:       10,
				ReqTimeout:   500 * time.Millisecond,
				ThrottleRate: time.Second,
				APIClient:    &sms.Client{},
			},
			want: wantType{
				statusCode: http.StatusRequestTimeout,
				response: sms.Response{
					Success: false,
					Error:   "Request timeout (process took to long to finish)",
				},
			},
		},

		"Created SMS": {
			httpMethod: http.MethodPost,
			path:       "/messages",
			payload:    strings.NewReader(`{"recipient":1234567890, "originator": "MessageBird", "message": "This is a test message"}`),
			serverOptions: sms.Config{
				Buffer:       10,
				ReqTimeout:   5 * time.Second,
				ThrottleRate: time.Second,
				APIClient:    &sms.Client{},
			},
			want: wantType{
				statusCode: http.StatusCreated,
				response: sms.Response{
					Success: true,
					Data: sms.Content{
						ID:         123,
						Recipient:  1234567890,
						Originator: "MessageBird",
						Message:    "This is a test message",
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r := httptest.NewRequest(tc.httpMethod, tc.path, tc.payload)
			w := httptest.NewRecorder()

			srv := sms.NewServer(tc.serverOptions)
			srv.Run()

			srv.ServeHTTP(w, r)

			res := w.Result()

			if res.StatusCode != tc.want.statusCode {
				t.Errorf("Status code was %d; want %d", res.StatusCode, tc.want.statusCode)
			}

			body, err := ioutil.ReadAll(res.Body)
			defer res.Body.Close()

			if err != nil {
				t.Fatal("Failed to read response body:", err)
			}

			var smsRes sms.Response

			if err := json.Unmarshal(body, &smsRes); err != nil {
				t.Fatal("Failed to unmarshal json response body:", err)
			}

			if smsRes != tc.want.response {
				t.Errorf("HTTP json response was %#v; want %#v", smsRes, tc.want.response)
			}
		})
	}
}
