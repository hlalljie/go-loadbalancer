package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
)

// Handles request and receives the post as an arguement
func handler(listeningPort string) http.HandlerFunc {
	// closure to handle request and allow for parameter to be received
	return func(w http.ResponseWriter, r *http.Request) {
		// Capture request info
		reqHost, reqPort, reqErr := net.SplitHostPort(r.RemoteAddr)
		if reqErr != nil {
			// print error info
			fmt.Printf("Error: %v\n", reqErr)
			return
		}
		// Print request info to server out
		// Print request address
		fmt.Printf("Received request from %s:%s\n", reqHost, reqPort)
		// Print request path
		fmt.Printf("%s %s %s\n", r.Method, r.URL.Path, r.Proto)
		// Print headers
		for name, values := range r.Header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", name, value)
			}
		}

		// Send a simple response
		fmt.Fprintf(w, "Hello from port %s\n", listeningPort)

		// Print confirmation
		fmt.Printf("\nReplied with hello message\n")
	}
}

// TODO: add error handling when server can't start on specified port
func startServer(port string) {
	// Set handler function
	http.HandleFunc("/", handler(port))
	fmt.Println("Server is listening on port " + port + "...")
	// Start server
	http.ListenAndServe(":"+port, nil)
}

func main() {
	port := flag.String("p", "8001", "Port to listen on")
	flag.Parse()
	startServer(*port)
}
