package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type RoundRobinBalancer struct {
	current uint64
	targets []string
}

func NewRoundRobinBalancer(targets []string) *RoundRobinBalancer {
	return &RoundRobinBalancer{targets: targets}
	// Note: current is automatically initialized to 0 (zero value for uint64)
}

func (rb *RoundRobinBalancer) NextTarget() string {
	// Atomic operation: Increments current and returns the new value
	next := atomic.AddUint64(&rb.current, 1) - 1
	// Print next
	fmt.Printf("Next Request id: %d, target: %s\n", rb.current, rb.targets[next%uint64(len(rb.targets))])
	// Gets the next target using the modulus of the current value and the number of targets
	return rb.targets[next%uint64(len(rb.targets))]
}

// Logs request information to standard out
func logRequest(r *http.Request) {
	// Logs request
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
}

func forwardRequest(target string, w http.ResponseWriter, r *http.Request) {
	// Parse the target server URL
	targetURL, err := url.Parse(target)
	if err != nil {
		http.Error(w, "Error parsing target URL", http.StatusInternalServerError)
		return
	}

	// Create a new reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Forward the request to the target server
	proxy.ServeHTTP(w, r)

	// Forwarding information
	fmt.Printf("\nForwarded request to %s\n", target)
}

// Handles request and receives the post as an arguement
func handler(lb *RoundRobinBalancer) http.HandlerFunc {
	// closure to handle request allow for parameter to be received

	return func(w http.ResponseWriter, r *http.Request) {

		logRequest(r)

		target := lb.NextTarget()

		forwardRequest(target, w, r)

	}
}

func startServer(lbPort string, targets []string) {

	// Create new round-robin load balancer
	lb := NewRoundRobinBalancer(targets)
	// Set handler function
	http.HandleFunc("/", handler(lb))
	log.Printf("Server is listening on port %s...\n", lbPort)
	// Start server
	log.Fatalf("Error starting server: %v", http.ListenAndServe(":"+lbPort, nil))
}

func main() {
	lbPort := "8000"
	targets := []string{"http://localhost:8001", "http://localhost:8002", "http://localhost:8003"}
	startServer(lbPort, targets)
}
