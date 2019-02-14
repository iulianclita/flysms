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
	ctx        context.Context
	resCh      chan Response
	Recipient  int64  `json:"recipient"`
	Originator string `json:"originator"`
	Message    string `json:"message"`
}

// Content keeps together all the parameters associated with a SMS
type Content struct {
	ID         string `json:"id"`
	Recipient  int64  `json:"recipient"`
	Originator string `json:"originator"`
	Message    string `json:"message"`
	Status     string `json:"status"`
	Created    string `json:"created"`
}

// Response is the representation of an HTTP response
// after succesfully handling a HTTP SMS request
type Response struct {
	statusCode int
	Success    bool    `json:"success"`
	Data       Content `json:"data,omitempty"`
	Error      string  `json:"error,omitempty"`
}

// Server is the frontend server that communicates to our SMS API
type Server struct {
	*http.ServeMux
	reqCh         chan *Request
	done          chan struct{}
	buf           int
	reqTimeout    time.Duration
	throttleRate  time.Duration
	messageClient *Client
}

// Config is a collection of configuration options for the server
type Config struct {
	Buffer        int
	ReqTimeout    time.Duration
	ThrottleRate  time.Duration
	MessageClient *Client
}

// NewServer creates a new server from the given config
func NewServer(cfg Config) *Server {
	return &Server{
		ServeMux:      http.NewServeMux(),
		reqCh:         make(chan *Request, cfg.Buffer),
		done:          make(chan struct{}),
		reqTimeout:    cfg.ReqTimeout,
		throttleRate:  cfg.ThrottleRate,
		messageClient: cfg.MessageClient,
	}
}

// createMessage is the HTTP handler for message creation
func (s *Server) createMessage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var res Response
		// Validate HTTP method
		if r.Method != http.MethodPost {
			res = Response{
				statusCode: http.StatusMethodNotAllowed,
				Error:      "Request not allowed (invalid HTTP method)",
			}
			sendResponse(w, res)
			return
		}

		// Validate JSON structure
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			res = Response{
				statusCode: http.StatusBadRequest,
				Error:      "Bad request (invalid payload json structure)",
			}
			sendResponse(w, res)
			return
		}

		// Validate recipient property value in json input
		// Make sure its length is between 7 and 15
		recp := fmt.Sprintf("%d", req.Recipient)

		if len(recp) < 7 || len(recp) > 15 {
			res = Response{
				statusCode: http.StatusUnprocessableEntity,
				Error:      "Invalid parameter (recipient value is out of bounds)",
			}
			sendResponse(w, res)
			return
		}

		// Validate originator property value in json input
		// Make sure it is present
		if len(req.Originator) == 0 {
			res = Response{
				statusCode: http.StatusUnprocessableEntity,
				Error:      "Missing parameter (originator value is not present)",
			}
			sendResponse(w, res)
			return
		}

		// Validate originator property value in json input
		// Make sure it's length does not go beyond 11 characters
		if len(req.Originator) > 11 {
			res = Response{
				statusCode: http.StatusUnprocessableEntity,
				Error:      "Invalid parameter (originator value is to long)",
			}
			sendResponse(w, res)
			return
		}

		// Validate message property value in json input
		// Make sure it is present
		if len(req.Message) == 0 {
			res = Response{
				statusCode: http.StatusUnprocessableEntity,
				Error:      "Missing parameter (message value is not present)",
			}
			sendResponse(w, res)
			return
		}

		// Validate message property value in json input
		// Make sure it's length does not go beyond 160 characters
		if len(req.Message) > 160 {
			res = Response{
				statusCode: http.StatusUnprocessableEntity,
				Error:      "Invalid parameter (message value is to long)",
			}
			sendResponse(w, res)
			return
		}

		ctx, cancel := context.WithTimeout(context.TODO(), s.reqTimeout)
		defer cancel()

		req.ctx = ctx
		req.resCh = make(chan Response)

		select {
		case s.reqCh <- &req:
			log.Printf("Accepted incoming request: %#v\n", req)
		default:
			log.Printf("Dropped incoming request: %#v\n", req)
			res = Response{
				statusCode: http.StatusTooManyRequests,
				Error:      "Request limit exceeded (request has been dropped)",
			}
			sendResponse(w, res)
			return
		}

		select {
		case res := <-req.resCh:
			sendResponse(w, res)
		case <-ctx.Done():
			res = Response{
				statusCode: http.StatusRequestTimeout,
				Error:      "Request timeout (process took to long to finish)",
			}
			sendResponse(w, res)
		}
	}
}

// Run the server
func (s *Server) Run() {
	s.HandleFunc("/messages", s.createMessage())
	go s.handleRequests()
}

func (s *Server) handleRequests() {
	ticker := time.Tick(s.throttleRate)

	for req := range s.reqCh {
		<-ticker
		go s.processRequest(req)
	}
}

func (s *Server) processRequest(req *Request) {
	done := make(chan struct{})
	var res Response

	go func() {
		defer close(done)
		if s.messageClient == nil {
			// In theory, this should never happen
			res = Response{
				statusCode: http.StatusInternalServerError,
				Error:      "Internal error (API client not set)",
			}
			return
		}
		// Make the API call
		msgRes, statusCode, err := s.messageClient.createMessage(req)
		if err != nil {
			res = Response{
				statusCode: http.StatusInternalServerError,
				Error:      "Internal error (API request failed)",
			}
			log.Printf("Failed creating SMS message through API for request %#v; Error: %v\n", req, err)
			return
		}

		switch v := msgRes.(type) {
		case MessageCreated:
			res = Response{
				statusCode: statusCode,
				Success:    true,
				Data: Content{
					ID:         v.ID,
					Originator: v.Originator,
					Message:    v.Body,
					Created:    v.CreatedDateTime.Format(time.RFC3339),
					Recipient:  v.Recipients.Items[0].Recipient,
					Status:     v.Recipients.Items[0].Status,
				},
			}
		case MessageErrors:
			res = Response{
				statusCode: statusCode,
				Success:    false,
				Error:      v.Errors[0].Description,
			}
		}
	}()

	select {
	case <-done:
		select {
		case req.resCh <- res:
			log.Println("Succesfully sent the response")
		default:
			// In theory, this should never happen
			log.Printf("Failed to send response %#v for request %#v\n", res, req)
		}
	case <-req.ctx.Done():
		log.Println("The API request timed out")
	}
}

func sendResponse(w http.ResponseWriter, res Response) {
	w.WriteHeader(res.statusCode)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Accept", "application/json")
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		log.Fatalf("Could not encode value %#v; Error: %v", res, err)
	}
}
