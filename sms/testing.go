package sms

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

const keyHeaderName = "AccessKey"

// NewTestServer starts a new development server
// The purpose of this server is to mimic the send SMS messagebird API behaviour
// It uses only a subset of the JSON response data coming from messagebird
// The server would normally need to treat also the error cases when the payload
// contains invalid input. This test server is oversimplified also because of the fact
// that the application does input validation before hiting the API.
func NewTestServer(t *testing.T, accessKey string) *httptest.Server {
	t.Helper()

	fn := func(w http.ResponseWriter, r *http.Request) {
		errCodes := make(map[int]MessageError)
		var errRes MessageErrors

		authHeader := r.Header.Get("Authorization")

		if !strings.HasPrefix(authHeader, keyHeaderName) {
			errCodes[2] = MessageError{
				Code:        2,
				Description: "Request not allowed (incorrect access_key)",
				Parameter:   "access_key",
			}
		}

		if len(authHeader) <= len(keyHeaderName) {
			errCodes[2] = MessageError{
				Code:        2,
				Description: "Request not allowed (incorrect access_key)",
				Parameter:   "access_key",
			}
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 {
			errCodes[2] = MessageError{
				Code:        2,
				Description: "Request not allowed (incorrect access_key)",
				Parameter:   "access_key",
			}
		}

		if len(headerParts) == 2 && headerParts[1] != accessKey {
			errCodes[2] = MessageError{
				Code:        2,
				Description: "Request not allowed (incorrect access_key)",
				Parameter:   "access_key",
			}
		}

		if len(errCodes) > 0 {
			for _, ec := range errCodes {
				errRes.Errors = append(errRes.Errors, ec)
			}

			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Accept", "application/json")
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(&errRes); err != nil {
				log.Fatalf("Could not encode value %#v; Error: %v", errRes, err)
			}

			return
		}

		if err := r.ParseForm(); err != nil {
			log.Fatalf("Could not parse incoming form request %#v; Error: %v", r, err)
		}

		recp, err := strconv.ParseInt(r.FormValue("recipients"), 10, 64)
		if err != nil {
			log.Fatalf("Could not convert recipients to int64 %s; Error: %v", r.FormValue("recipients"), err)
		}

		okRes := MessageCreated{
			ID:              fmt.Sprintf("%d", time.Now().UnixNano()),
			Originator:      r.FormValue("originator"),
			Body:            r.FormValue("body"),
			CreatedDateTime: time.Now(),
			Recipients: MessageRecipients{
				TotalSentCount:           1,
				TotalDeliveredCount:      0,
				TotalDeliveryFailedCount: 0,
				Items: []MessageItem{
					{
						Recipient:      recp,
						Status:         "sent",
						StatusDateTime: time.Now(),
					},
				},
			},
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Accept", "application/json")
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(&okRes); err != nil {
			log.Fatalf("Could not encode value %#v; Error: %v", okRes, err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/messages", fn)

	return httptest.NewServer(mux)
}
