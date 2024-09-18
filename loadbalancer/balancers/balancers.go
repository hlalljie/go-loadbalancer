package balancers

import (
	"sync"
	"sync/atomic"
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
