package main

import (
	"fmt"
	"loadbalancer/balancers"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func proxyHandler(lb *balancers.RoundRobinBalancer, w http.ResponseWriter, r *http.Request) {
	log.Println("Retrying request")
	logRequest(r)
	// find next target based on load balancing function
	target := lb.NextTarget()
	// TODO: add error handling when server can't start on specified port
	if target == "" {
		log.Printf("Error: No target available for request\n")
		http.Error(w, "No server available", http.StatusServiceUnavailable)
		return
	}

	// forward request
	forwardRequest(lb, target, w, r)
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

func forwardRequest(lb *balancers.RoundRobinBalancer, target string, w http.ResponseWriter, r *http.Request) {
	// Parse the target server URL
	targetURL, err := url.Parse(target)
	if err != nil {
		http.Error(w, "Error parsing target URL", http.StatusInternalServerError)
		return
	}

	// Create a new reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		lb.RemoveTarget(target)
		log.Printf("Error forwarding request to %s: %v, retrying...", target, e)
		http.Error(w, "Error forwarding request, retrying...", http.StatusInternalServerError)
		proxyHandler(lb, w, r)
	}

	// Forward the request to the target server
	proxy.ServeHTTP(w, r)

	// Forwarding information
	fmt.Printf("\nForwarded request to %s\n", target)
}

// Handles request and receives the post as an arguement
func handler(lb *balancers.RoundRobinBalancer) http.HandlerFunc {
	// closure to handle request allow for parameter to be received

	return func(w http.ResponseWriter, r *http.Request) {

		logRequest(r)
		// find next target based on load balancing function
		target := lb.NextTarget()
		// TODO: add error handling when server can't start on specified port
		if target == "" {
			log.Printf("Error: No target available for request\n")
			http.Error(w, "No server available", http.StatusServiceUnavailable)
			return
		}

		forwardRequest(lb, target, w, r)

	}
}

func startServer(lbPort string, targets []string) {

	// Create new round-robin load balancer
	lb := balancers.NewRoundRobinBalancer(targets)
	// Set handler function
	http.HandleFunc("/", handler(lb))
	log.Printf("Server is listening on port %s...\n", lbPort)
	// Run health checks on target list
	go lb.RunHealthChecks(5000)
	// Start server
	log.Fatalf("Error starting server: %v", http.ListenAndServe(":"+lbPort, nil))

}

func main() {
	lbPort := "8000"
	targets := []string{"http://localhost:8001", "http://localhost:8002", "http://localhost:8003"}
	startServer(lbPort, targets)
}
