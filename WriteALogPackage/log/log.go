package log

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"

	"io"

	api "github.com/GergesHany/Event-Streaming-System/StructureDataWithProtobuf/api/v1"
)

type Log struct {
	mu sync.RWMutex

	Dir    string
	Config Config

	activeSegment *segment
	segments      []*segment
}

func NewLog(dir string, c Config) (*Log, error) {
	if c.Segment.MaxIndexBytes == 0 {
		c.Segment.MaxIndexBytes = 1024
	}

	if c.Segment.MaxStoreBytes == 0 {
		c.Segment.MaxStoreBytes = 1024
	}

	l := &Log{
		Dir:    dir,
		Config: c,
	}
	return l, l.setup()
}

// setup initializes the log by loading existing segments from disk or creating a new initial segment
func (l *Log) setup() error {
	files, err := ioutil.ReadDir(l.Dir)
	if err != nil {
		return err
	}

	var baseOffsets []uint64
	for _, file := range files {
		// TrimSuffix returns s without the provided trailing suffix string. If s doesn't end with suffix, s is returned unchanged.
		offStr := strings.TrimSuffix(file.Name(), path.Ext(file.Name()))
		off, _ := strconv.ParseUint(offStr, 10, 0)
		baseOffsets = append(baseOffsets, off)
	}

	sort.Slice(baseOffsets, func(i, j int) bool {
		return baseOffsets[i] < baseOffsets[j]
	})

	for i := 0; i < len(baseOffsets); i = i + 2 {
		// Create a new segment for each unique base offset
		// Each segment consists of a pair of files: .store and .index
		// So, we increment i by 2 in each iteration to skip the duplicate base offsets
		if err = l.newSegment(baseOffsets[i]); err != nil {
			return err
		}
	}

	// If no segments exist, create the initial segment
	if l.segments == nil {
		if err = l.newSegment(l.Config.Segment.InitialOffset); err != nil {
			return err
		}
	}

	return nil
}

// newSegment creates a new segment and adds it to the log
func (l *Log) newSegment(off uint64) error {
	s, err := newSegment(l.Dir, off, l.Config)
	if err != nil {
		return err
	}
	l.segments = append(l.segments, s)
	l.activeSegment = s
	return nil
}

func (l *Log) Append(record *api.Record) (uint64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	off, err := l.activeSegment.Append(record)
	if err != nil {
		return 0, err
	}

	// If the active segment is full, create a new segment
	if l.activeSegment.IsMaxed() {
		err = l.newSegment(off + 1)
	}
	return off, err
}

func (l *Log) Read(off uint64) (*api.Record, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var s *segment
	for _, segment := range l.segments {
		// Check if the offset falls within the range of this segment
		if segment.baseOffset <= off && off < segment.nextOffset {
			s = segment
			break
		}
	}

	// If no segment found or offset is out of range
	if s == nil || s.nextOffset <= off {
		return nil, fmt.Errorf("offset out of range: %d", off)
	}

	return s.Read(off)
}

// Close closes all the segments in the log
func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, segment := range l.segments {
		if err := segment.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (l *Log) Remove() error {
	if err := l.Close(); err != nil {
		return err
	}
	return os.RemoveAll(l.Dir)
}

// Reset removes all the segments and creates a new initial segment
func (l *Log) Reset() error {
	if err := l.Remove(); err != nil {
		return err
	}
	return l.setup()
}

// LowestOffset returns the lowest offset in the log
func (l *Log) LowestOffset() (uint64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if len(l.segments) == 0 {
		return 0, nil
	}
	return l.segments[0].baseOffset, nil
}

// HighestOffset returns the highest offset in the log
func (l *Log) HighestOffset() (uint64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if len(l.segments) == 0 {
		return 0, nil
	}
	off := l.segments[len(l.segments)-1].nextOffset
	if off == 0 {
		return 0, nil
	}
	return off - 1, nil
}

// Truncate(lowest uint64) removes all segments whose highest offset is lower than lowest.
// Because we don’t have disks with infinite space,
func (l *Log) Truncate(lowest uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var segments []*segment
	for _, s := range l.segments {
		if s.nextOffset <= lowest+1 {
			if err := s.Remove(); err != nil {
				return err
			}
			continue
		}
		segments = append(segments, s)
	}

	l.segments = segments
	return nil
}

// originReader is a wrapper around store to implement io.Reader interface
// It keeps track of the current read offset within the store
type originReader struct {
	*store
	off int64
}

// Reader returns an io.Reader to read the entire log.
// It uses io.MultiReader to concatenate readers for each segment's store.
func (l *Log) Reader() io.Reader {
	l.mu.RLock()
	defer l.mu.RUnlock()
	readers := make([]io.Reader, len(l.segments))
	for i, segment := range l.segments {
		readers[i] = &originReader{segment.store, 0}
	}
	return io.MultiReader(readers...)
}

// ReadAt reads len(p) bytes from the store at the given offset.
// Read implements the io.ReaderAt interface.
func (o *originReader) Read(p []byte) (int, error) {
	n, err := o.ReadAt(p, o.off)
	o.off += int64(n)
	return n, err
}
