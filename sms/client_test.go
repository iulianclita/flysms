package sms_test

import (
	"testing"

	"github.com/iulianclita/messagebird/sms"
)

func TestClient_URL(t *testing.T) {
	tests := map[string]struct {
		opts sms.Options
		path string
		want string
	}{
		"No base URL without before slash": {
			opts: sms.Options{},
			path: "messages",
			want: "https://rest.messagebird.com/messages",
		},

		"No base URL with before slash": {
			opts: sms.Options{},
			path: "/messages",
			want: "https://rest.messagebird.com/messages",
		},

		"Base URL without before slash": {
			opts: sms.Options{
				BaseURL: "https://example.com",
			},
			path: "messages",
			want: "https://example.com/messages",
		},

		"Base URL with before slash": {
			opts: sms.Options{
				BaseURL: "https://example.com",
			},
			path: "/messages",
			want: "https://example.com/messages",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := sms.NewClient(tc.opts)
			got := client.URL(tc.path)
			if got != tc.want {
				t.Errorf("URL(%q) = %q; want %q", tc.path, got, tc.want)
			}
		})
	}
}
