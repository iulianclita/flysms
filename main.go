package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/iulianclita/messagebird/sms"
)

const port = 3500

func main() {
	fmt.Printf("Listening on port %d\n", port)

	srv := sms.NewServer(100)
	srv.Start()

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), srv); err != nil {
		log.Fatal("Failed to start server")
	}
}
