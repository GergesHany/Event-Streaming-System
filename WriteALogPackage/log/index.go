package log

import (
	"io"
	"os"

	"github.com/tysonmote/gommap"
)

/*
	* Index Entries: Each entry contains:
		Offset (4 bytes): The logical offset of the record in the log
		Position (8 bytes): The physical byte position where the record starts in the store file

	* Fixed-Width Design: Since each entry is exactly 12 bytes, you can jump directly to any entry using the formula: entry_position = offset * 12

	* This index is typically used in conjunction with a store file:
		Store file: Contains the actual log records (variable length)
		Index file: Contains fixed-width entries that point to records in the store file

	* This design provides O(1) lookup time for any record by its offset, making it very efficient for distributed log systems where fast random access is crucial.
*/

var (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
)

type index struct {
	file *os.File
	mmap gommap.MMap // Uses memory-mapped I/O for efficient random access to the file
	size uint64
}

// newIndex creates and initializes a new index for efficient log record lookup.
func newIndex(f *os.File, c Config) (*index, error) {
	idx := &index{
		file: f,
	}

	// Get current file information to determine existing size
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}

	idx.size = uint64(fi.Size())

	// Pre-allocate file to maximum size for performance optimization
	// This avoids frequent file system calls during writes
	if err = f.Truncate(int64(c.Segment.MaxIndexBytes)); err != nil {
		return nil, err
	}

	// Memory-map the file for efficient random access
	// PROT_READ|PROT_WRITE allows both reading and writing
	// MAP_SHARED ensures changes are visible to other processes
	if idx.mmap, err = gommap.Map(idx.file.Fd(), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED); err != nil {
		return nil, err
	}

	return idx, err
}

func (i *index) Close() error {
	// Flushes any changes from the memory-mapped region back to the underlying file
	if err := i.mmap.Sync(gommap.MS_ASYNC); err != nil {
		return err
	}

	//  Forces the operating system to flush any buffered data to the physical storage device
	if err := i.file.Sync(); err != nil {
		return err
	}

	// Shrinks the file back to its actual used size
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}

	return i.file.Close()
}

// Read returns the associated record's position in the store given the offset
func (i *index) Read(in int64) (out uint32, pos uint64, err error) {
	// Return EOF if index is empty
	if i.size == 0 {
		return 0, 0, io.EOF
	}

	// If in is -1, return the last entry
	if in == -1 {
		out = uint32((i.size / entWidth) - 1)
	} else {
		out = uint32(in)
	}

	// Calculate position in memory-mapped file
	pos = uint64(out) * entWidth
	if i.size < (pos + entWidth) {
		return 0, 0, io.EOF
	}

	// Read offset and position from memory-mapped file
	out = enc.Uint32(i.mmap[pos : pos+offWidth])
	pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])

	return out, pos, nil
}

// Write persists an offset and position pair to the index
func (i *index) Write(off uint32, pos uint64) error {
	// Check if there's space for another entry
	if uint64(len(i.mmap)) < i.size+entWidth {
		return io.EOF
	}

	// Write offset and position to memory-mapped file
	enc.PutUint32(i.mmap[i.size:i.size+offWidth], off)
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)

	// Update the current size
	i.size += uint64(entWidth)
	return nil
}

func (i *index) Name() string {
	return i.file.Name()
}
