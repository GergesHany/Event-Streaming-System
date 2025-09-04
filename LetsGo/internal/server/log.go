package server

import (
	"fmt"
	"sync"
)

// Record represents a single log entry with its value and offset position
type Record struct {
	Value  []byte `json:"value"`
	Offset uint64 `json:"offset"`
}

var (
	ErrOffsetNotFound = fmt.Errorf("offset not found")
)

// Log is a thread-safe in-memory log that stores records sequentially
type Log struct {
	mu      sync.Mutex
	records []Record
}

func NewLog() *Log {
	return &Log{}
}

// Append adds a new record to the end of the log and returns its offset
func (c *Log) Append(record Record) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Set the offset to the current length (next available position)
	record.Offset = uint64(len(c.records))
	// Add the record to the end of the slice
	c.records = append(c.records, record)
	return record.Offset, nil
}

// Read retrieves a record at the specified offset
func (c *Log) Read(offset uint64) (Record, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if the requested offset exists
	if offset >= uint64(len(c.records)) {
		return Record{}, ErrOffsetNotFound
	}
	return c.records[offset], nil
}
