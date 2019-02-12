package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/iulianclita/messagebird/sms"
)

const port = 3500

func main() {
	fmt.Printf("Listening on port %d\n", port)

	cfg := sms.Config{
		Buffer:       10,
		ReqTimeout:   5 * time.Second,
		ThrottleRate: time.Second,
		APIClient:    &sms.Client{},
	}

	srv := sms.NewServer(cfg)
	srv.Run()

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), srv); err != nil {
		log.Fatal("Failed to start server")
	}
}
