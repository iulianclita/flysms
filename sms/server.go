package sms

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Request ...
type Request struct {
	Recipient  int64  `json:"recipient"`
	Originator string `json:"originator"`
	Message    string `json:"message"`
	ctx        context.Context
	resCh      chan *Response
}

// Response ...
type Response struct {
	Success bool `json:"success"`
	Data    struct {
		ID int64 `json:"id"`
	} `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// Server ...
type Server struct {
	*http.ServeMux
	reqCh chan *Request
	rate  time.Duration
}

// NewServer creates a new server with a given buffer size
// which will be used to accumulate incoming requests
func NewServer(buf int) *Server {
	return &Server{
		ServeMux: http.NewServeMux(),
		reqCh:    make(chan *Request, buf),
		rate:     time.Second,
	}
}

func (s *Server) createMessage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate HTTP method
		if r.Method != http.MethodPost {
			res := Response{
				Success: false,
				Error:   http.StatusText(http.StatusMethodNotAllowed),
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Header().Set("Cotent-Type", "application/json")
			w.Header().Set("Accept", "application/json")
			if err := json.NewEncoder(w).Encode(&res); err != nil {
				log.Fatalf("Cannot encode value %#v; Error: %v", res, err)
			}
			return
		}

		// Validate JSON structure
		var req Request

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			res := Response{
				Success: false,
				Error:   http.StatusText(http.StatusUnprocessableEntity),
			}
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Header().Set("Cotent-Type", "application/json")
			w.Header().Set("Accept", "application/json")
			if err := json.NewEncoder(w).Encode(&res); err != nil {
				log.Fatalf("Cannot encode value %#v; Error: %v", res, err)
			}
			return
		}
	}
}

// Start starts the server
func (s *Server) Start() {
	s.HandleFunc("/messages", s.createMessage())
	// go s.handleRequests()
}

// func (s *Server) handleRequests() {
// 	ticker := time.Tick(s.rate)
// 	for req := range s.reqCh {
// 		<-ticker
// 		go s.processRequest(req)
// 	}
// }

// func (s *Server) processRequest(req *Request) {
// 	done := make(chan struct{})
// 	go func() {
// 		// make the API call
// 		close(done)
// 	}()

// 	select {
// 	case <-done:

// 	}
// }
