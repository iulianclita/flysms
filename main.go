package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/iulianclita/messagebird/sms"
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

	// TEST
	opts = sms.Options{
		AccessKey: "a2joQy9zlSaJg2UdrMV7HnsZt", // a2joQy9zlSaJg2UdrMV7HnsZt fake_key
		Timeout:   10 * time.Second,
	}

	c := sms.NewClient(opts)

	r := &sms.Request{
		Recipient:  40724423589,
		Originator: "+40724423589",
		Message:    "Hi! This is your first message.",
	}

	data, statusCode, err := sms.SendMessasge(c, r)

	fmt.Printf("DATA: %#v\n", data)
	fmt.Printf("STATUS CODE: %#v\n", statusCode)
	fmt.Printf("ERROR: %#v\n", err)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), srv); err != nil {
		log.Fatal("Failed to start server")
	}
}
