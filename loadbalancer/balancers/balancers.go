package balancers

import (
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type RoundRobinBalancer struct {
	current        uint64
	targets        []string
	healthyTargets []string
	rwMutex        sync.RWMutex
}

func NewRoundRobinBalancer(targets []string) *RoundRobinBalancer {
	return &RoundRobinBalancer{
		targets:        targets,
		healthyTargets: append([]string(nil), targets...)}
}

func (rb *RoundRobinBalancer) NextTarget() string {
	// Atomic operation: Increments current and returns the new value
	next := atomic.AddUint64(&rb.current, 1) - 1

	// Locks the targets slice to prevent misread during modification
	rb.rwMutex.RLock()
	defer rb.rwMutex.RUnlock()

	// If there are no targets, return an empty string
	if len(rb.healthyTargets) == 0 {
		return ""
	}

	// Gets the next target using the modulus of the current value and the number of targets
	return rb.healthyTargets[next%uint64(len(rb.healthyTargets))]
}

func (rb *RoundRobinBalancer) RemoveTarget(target string) {
	rb.rwMutex.Lock()
	defer rb.rwMutex.Unlock()
	// loop through all targets looking for target to remove
	for i, t := range rb.healthyTargets {
		if t == target {
			// Remove the element at index i
			copy(rb.healthyTargets[i:], rb.healthyTargets[i+1:])
			rb.healthyTargets[len(rb.healthyTargets)-1] = "" // Clear last element (optional)
			rb.healthyTargets = rb.healthyTargets[:len(rb.healthyTargets)-1]
			break
		}
	}
}

func (rb *RoundRobinBalancer) AddTarget(target string) {
	rb.rwMutex.Lock()
	defer rb.rwMutex.Unlock()
	rb.healthyTargets = append(rb.healthyTargets, target)
}

func healthCheck(target string) bool {
	// Ping target server and return true if successful
	_, err := http.Get(target)
	// if error, return false and show error
	if err != nil {
		// print error
		log.Printf("Server %s is not healthy: %v\n", target, err)
		return false
	}
	return true

}

// Checks the health of servers at a specified interval (milliseconds)
//
//	lb (*balancers.RoundRobinBalancer): load balancer to health check
//	intervalDuration (int): time in milliseconds between health checks
func (rb *RoundRobinBalancer) RunHealthChecks(intervalDuration int) {
	for {
		log.Printf("Running health checks...")
		// create new health targets slice
		// loop through all targets looking for target to remove or add
		newHealthyTargets := []string{}
		for _, target := range rb.targets {
			// if target is healthy track it
			if healthCheck(target) {
				newHealthyTargets = append(newHealthyTargets, target)
			}
		}

		// update healthy targets
		rb.rwMutex.Lock()
		rb.healthyTargets = newHealthyTargets
		rb.rwMutex.Unlock()

		// Sleep for interval duration (milliseconds)
		time.Sleep(time.Duration(intervalDuration) * time.Millisecond)
	}
}
