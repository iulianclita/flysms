package sms_test

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iulianclita/messagebird/sms"
)

func TestServer_createMessage(t *testing.T) {
	srv := sms.NewServer(100)
	srv.Start()

	type wantType struct {
		statusCode int
		response   sms.Response
	}

	tests := map[string]struct {
		httpMethod string
		path       string
		payload    io.Reader
		want       wantType
	}{
		"Method not allowed": {
			httpMethod: http.MethodGet,
			path:       "/messages",
			payload:    nil,
			want: wantType{
				statusCode: http.StatusMethodNotAllowed,
				response: sms.Response{
					Success: false,
					Error:   http.StatusText(http.StatusMethodNotAllowed),
				},
			},
		},
		"Invalid JSON": {
			httpMethod: http.MethodPost,
			path:       "/messages",
			payload:    strings.NewReader(`{"invalid_json"}`),
			want: wantType{
				statusCode: http.StatusUnprocessableEntity,
				response: sms.Response{
					Success: false,
					Error:   http.StatusText(http.StatusUnprocessableEntity),
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r := httptest.NewRequest(tc.httpMethod, tc.path, tc.payload)
			w := httptest.NewRecorder()

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
