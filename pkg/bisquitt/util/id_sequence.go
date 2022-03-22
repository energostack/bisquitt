package util

import (
	"sync"
)

// IDSequence generates successive uint16 IDs from (minID, maxID) range.
// When maxID is reached, the counter wraps to minID on the following Next()
// call and the overflow is signalized.
type IDSequence struct {
	lock     sync.Mutex
	next     uint16
	min      uint16
	max      uint16
	overflow bool
}

// NewIDSequence creates a new IDSequence.
//
// Both minID and maxID are inclusive.
func NewIDSequence(minID, maxID uint16) *IDSequence {
	return &IDSequence{
		next: minID,
		min:  minID,
		max:  maxID,
	}
}

// Next returns a new ID. If (maxID, false) is returned, the following call will
// return (minID, true).
func (c *IDSequence) Next() (id uint16, overflow bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	id = c.next
	overflow = c.overflow
	if c.overflow {
		c.overflow = false
	}

	if c.next == c.max {
		c.next = c.min
		c.overflow = true
	} else {
		c.next++
	}

	return
}
