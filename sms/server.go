package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Request is the representation of an SMS request
// and is extracted from the HTTP request body
type Request struct {
	Recipient  int64  `json:"recipient"`
	Originator string `json:"originator"`
	Message    string `json:"message"`
	ctx        context.Context
	resCh      chan Response
}

// Response is the representation of an HTTP response
// after succesfully handling a HTTP SMS request
type Response struct {
	Success bool `json:"success"`
	Data    struct {
		ID int64 `json:"id"`
	} `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// Server is the frontend server that communicates to our SMS API
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
		var res Response
		// Validate HTTP method
		if r.Method != http.MethodPost {
			res = Response{
				Error: "Request not allowed (invalid HTTP method)",
			}
			sendResponse(w, http.StatusMethodNotAllowed, res)
			return
		}

		// Validate JSON structure
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			res = Response{
				Error: "Bad request (invalid payload json structure)",
			}
			sendResponse(w, http.StatusBadRequest, res)
			return
		}

		// Validate recipient property value in json input
		// Make sure its length is between 7 and 15
		recp := fmt.Sprintf("%d", req.Recipient)

		if len(recp) < 7 || len(recp) > 15 {
			res = Response{
				Error: "Invalid parameter (recipient value is out of bounds)",
			}
			sendResponse(w, http.StatusUnprocessableEntity, res)
			return
		}

		// Validate originator property value in json input
		// Make sure it is present
		if len(req.Originator) == 0 {
			res = Response{
				Error: "Missing parameter (originator value is not present)",
			}
			sendResponse(w, http.StatusUnprocessableEntity, res)
			return
		}

		// Validate originator property value in json input
		// Make sure it's length does not go beyond 11 characters
		if len(req.Originator) > 11 {
			res = Response{
				Error: "Invalid parameter (originator value is to long)",
			}
			sendResponse(w, http.StatusUnprocessableEntity, res)
			return
		}

		// Validate message property value in json input
		// Make sure it is present
		if len(req.Message) == 0 {
			res = Response{
				Error: "Missing parameter (message value is not present)",
			}
			sendResponse(w, http.StatusUnprocessableEntity, res)
			return
		}

		// Validate message property value in json input
		// Make sure it's length does not go beyond 160 characters
		if len(req.Message) > 160 {
			res = Response{
				Error: "Invalid parameter (message value is to long)",
			}
			sendResponse(w, http.StatusUnprocessableEntity, res)
			return
		}

		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()

		req.ctx = ctx
		req.resCh = make(chan Response)

		select {
		case s.reqCh <- &req:
			log.Printf("Accepted incoming request: %#v\n", req)
		default:
			log.Printf("Dropped incoming request: %#v\n", req)
			res = Response{
				Error: "Request limit exceeded (request has been dropped)",
			}
			sendResponse(w, http.StatusTooManyRequests, res)
			return
		}

		select {
		case res := <-req.resCh:
			sendResponse(w, http.StatusCreated, res)
		case <-ctx.Done():
			res = Response{
				Error: "Request timeout (process took to long to finish)",
			}
			sendResponse(w, http.StatusRequestTimeout, res)
		}
	}
}

// Start starts the server
func (s *Server) Start() {
	s.HandleFunc("/messages", s.createMessage())
	go s.handleRequests()
}

func (s *Server) handleRequests() {
	ticker := time.Tick(s.rate)
	for req := range s.reqCh {
		<-ticker
		go s.processRequest(req)
	}
}

func (s *Server) processRequest(req *Request) {
	done := make(chan struct{})
	var res Response

	go func() {
		// fake the API call
		time.Sleep(1 * time.Second)
		res = Response{
			Success: true,
			Data: struct {
				ID int64 `json:"id"`
			}{
				ID: 123,
			},
		}
		close(done)
	}()

	select {
	case <-done:
		select {
		case req.resCh <- res:
			log.Println("Succesfully sent response")
		default:
			// In theory, this should never happen
			log.Printf("Failed to send response %#v for request %#v\n", res, req)
		}
	case <-req.ctx.Done():
		log.Println("Request timeout (process took to long to finish)")
	}
}

func sendResponse(w http.ResponseWriter, statusCode int, res Response) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Accept", "application/json")
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		log.Fatalf("Cannot encode value %#v; Error: %v", res, err)
	}
}
