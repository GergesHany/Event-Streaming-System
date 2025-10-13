package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	// Encode data to persist it to a disk
	enc = binary.BigEndian
)

const (
	// Each record is written to the segment file with an 8-byte length prefix
	LenWidth = 8
)

type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}

	size := uint64(fi.Size())
	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f),
	}, nil
}

/*
  - [8-byte length][actual data][8-byte length][actual data]...
  - The 8-byte length stores the size of the actual data that follows it.
  - File content: [0,0,0,0,0,0,0,11] [H,e,l,l,o, ,W,o,r,l,d]
                  ^^^^^^^^^^^^^^^^^^ ^^^^^^^^^^^^^^^^^^^^^^^
              8-byte length     actual data (11 bytes)
              (value = 11)
*/

func (s *store) Append(p []byte) (n, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pos = s.size
	// Write the length of the data first (8 bytes)
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}

	// Write the actual data
	w, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}

	w += LenWidth       // Add the length prefix size
	s.size += uint64(w) // Update total size
	return uint64(w), pos, nil
}

func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// To flush the buffer to the file.
	if err := s.buf.Flush(); err != nil {
		return nil, err
	}

	size := make([]byte, LenWidth)
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}

	b := make([]byte, enc.Uint64(size))
	if _, err := s.File.ReadAt(b, int64(pos+LenWidth)); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// To flush the buffer to the file.v
	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, off)
}

func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.buf.Flush(); err != nil {
		return err
	}
	return s.File.Close()
}
