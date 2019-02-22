package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/iulianclita/flysms/sms"
)

const port = 3500

func main() {
	fmt.Printf("Listening on port %d\n", port)

	opts := sms.Options{
		AccessKey: os.Getenv("MESSAGE_BIRD_ACCESSKEY"),
		Timeout:   10 * time.Second,
	}

	cfg := sms.Config{
		Buffer:        10,
		ReqTimeout:    5 * time.Second,
		ThrottleRate:  time.Second,
		MessageClient: sms.NewClient(opts),
	}

	srv := sms.NewServer(cfg)
	srv.Run()

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), srv); err != nil {
		log.Fatal("Failed to start server")
	}
}
