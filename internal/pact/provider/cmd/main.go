package main

import (
	"log"
	"net"
	"net/http"
	"ralts-cms/internal/pact/provider"
)

func main() {
	mux := provider.GetHTTPHandler()

	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Printf("API starting: port %d (%s)", 8081, ln.Addr())
	log.Printf("API terminating: %v", http.Serve(ln, mux))
}
