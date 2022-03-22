package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDSequence_Basic(t *testing.T) {
	assert := assert.New(t)

	uint16ID := NewIDSequence(5, 7)

	id, overflow := uint16ID.Next()
	assert.Equal(uint16(5), id)
	assert.Equal(false, overflow)

	id, overflow = uint16ID.Next()
	assert.Equal(uint16(6), id)
	assert.Equal(false, overflow)

	id, overflow = uint16ID.Next()
	assert.Equal(uint16(7), id)
	assert.Equal(false, overflow)

	id, overflow = uint16ID.Next()
	assert.Equal(uint16(5), id)
	assert.Equal(true, overflow)

	id, overflow = uint16ID.Next()
	assert.Equal(uint16(6), id)
	assert.Equal(false, overflow)
}

func TestIDSequence_Single(t *testing.T) {
	assert := assert.New(t)

	uint16ID := NewIDSequence(5, 5)

	id, overflow := uint16ID.Next()
	assert.Equal(uint16(5), id)
	assert.Equal(false, overflow)

	id, overflow = uint16ID.Next()
	assert.Equal(uint16(5), id)
	assert.Equal(true, overflow)

	id, overflow = uint16ID.Next()
	assert.Equal(uint16(5), id)
	assert.Equal(true, overflow)

	id, overflow = uint16ID.Next()
	assert.Equal(uint16(5), id)
	assert.Equal(true, overflow)
}

func ExampleIDSequence_Next() {
	s := NewIDSequence(1, 3)

	for i := 0; i < 8; i++ {
		id, overflow := s.Next()
		fmt.Printf("id=%d, overflow=%t\n", id, overflow)
	}

	// Output:
	// id=1, overflow=false
	// id=2, overflow=false
	// id=3, overflow=false
	// id=1, overflow=true
	// id=2, overflow=false
	// id=3, overflow=false
	// id=1, overflow=true
	// id=2, overflow=false
}
