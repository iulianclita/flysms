package sms

import (
	"context"
	"encoding/json"
	"fmt"
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
			sendErrorResponse(w, http.StatusMethodNotAllowed, "Request not allowed (invalid HTTP method)")
			return
		}

		// Validate JSON structure
		var req Request

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendErrorResponse(w, http.StatusBadRequest, "Bad request (invalid payload json structure)")
			return
		}

		// Validate recipient property value in json input
		// Make sure its length is between 7 and 15
		recp := fmt.Sprintf("%d", req.Recipient)

		if len(recp) < 7 || len(recp) > 15 {
			sendErrorResponse(w, http.StatusUnprocessableEntity, "Invalid parameter (recipient value is out of bounds)")
			return
		}

		// Validate originator property value in json input
		// Make sure it is present
		if len(req.Originator) == 0 {
			sendErrorResponse(w, http.StatusUnprocessableEntity, "Missing parameter (originator value is not present)")
			return
		}

		// Validate originator property value in json input
		// Make sure it's length does not go beyond 11 characters
		if len(req.Originator) > 11 {
			sendErrorResponse(w, http.StatusUnprocessableEntity, "Invalid parameter (originator value is to long)")
			return
		}

		// Validate message property value in json input
		// Make sure it is present
		if len(req.Message) == 0 {
			sendErrorResponse(w, http.StatusUnprocessableEntity, "Missing parameter (message value is not present)")
			return
		}

		// Validate message property value in json input
		// Make sure it's length does not go beyond 160 characters
		if len(req.Message) > 160 {
			sendErrorResponse(w, http.StatusUnprocessableEntity, "Invalid parameter (message value is to long)")
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

func sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	res := Response{
		Success: false,
		Error:   message,
	}
	w.WriteHeader(statusCode)
	w.Header().Set("Cotent-Type", "application/json")
	w.Header().Set("Accept", "application/json")
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		log.Fatalf("Cannot encode value %#v; Error: %v", res, err)
	}
}
