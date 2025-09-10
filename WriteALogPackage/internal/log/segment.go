package log

import (
	"fmt"
	"os"
	"path"

	api "github.com/GergesHany/Event-Streaming-System/StructureDataWithProtobuf/api/v1"
	"google.golang.org/protobuf/proto"
)

type segment struct {
	store                  *store
	index                  *index
	baseOffset, nextOffset uint64
	config                 Config
}

// newSegment creates a new segment with the given base offset and configuration.
func newSegment(dir string, baseOffset uint64, c Config) (*segment, error) {
	s := &segment{
		baseOffset: baseOffset,
		config:     c,
	}

	// ---------- Initialize the store ----------

	// Open the store file in read-write and append mode, creating it if it doesn't exist
	storeFile, err := os.OpenFile(
		path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".store")),
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)

	if err != nil {
		return nil, err
	}

	if s.store, err = newStore(storeFile); err != nil {
		return nil, err
	}

	// ---------- Initialize the index ----------

	// Open the index file in read-write and append mode, creating it if it doesn't exist
	indexFile, err := os.OpenFile(
		path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".index")),
		os.O_RDWR|os.O_CREATE,
		0644,
	)

	if err != nil {
		return nil, err
	}

	if s.index, err = newIndex(indexFile, c); err != nil {
		return nil, err
	}

	// ---------- Set the next offset ----------
	// Read the last entry in the index to determine the next offset
	// If the index is empty, start at the base offset
	// Otherwise, set the next offset to one more than the last offset in the index
	if off, _, err := s.index.Read(-1); err != nil {
		s.nextOffset = baseOffset
	} else {
		s.nextOffset = baseOffset + uint64(off) + 1
	}

	return s, nil
}

// Append adds a new record to the segment and returns its offset.
func (s *segment) Append(record *api.Record) (offset uint64, err error) {
	cur := s.nextOffset
	record.Offset = cur

	p, err := proto.Marshal(record)
	if err != nil {
		return 0, nil
	}

	_, pos, err := s.store.Append(p)
	if err != nil {
		return 0, nil
	}

	if err = s.index.Write(
		// index offsets are relative to base offset: <baseOffset + indexOffset> = absoluteOffset
		// So, to get the indexOffset, we subtract baseOffset from nextOffset
		/*
			           - baseOffset: the starting offset of the segment
					   - nextOffset: the next available offset in the segment
					   - (nextOffset - baseOffset): gives the relative offset within the segment
		*/
		uint32(s.nextOffset-uint64(s.baseOffset)),
		pos,
	); err != nil {
		return 0, err
	}

	s.nextOffset++
	return cur, nil
}

// Read retrieves a record by its offset from the segment.
func (s *segment) Read(off uint64) (*api.Record, error) {
	_, pos, err := s.index.Read(int64(off - s.baseOffset))
	if err != nil {
		return nil, err
	}

	p, err := s.store.Read(pos)
	if err != nil {
		return nil, err
	}

	record := &api.Record{}
	//  Deserializes the raw bytes (p) back into the structured record:
	err = proto.Unmarshal(p, record)
	return record, err
}

// IsMaxed checks if the segment has reached its maximum size for either the store or the index.
func (s *segment) IsMaxed() bool {
	return s.store.size >= s.config.Segment.MaxStoreBytes || s.index.size >= s.config.Segment.MaxIndexBytes
}

func (s *segment) Remove() error {
	if err := s.Close(); err != nil {
		return err
	}
	if err := os.Remove(s.index.Name()); err != nil {
		return err
	}
	if err := os.Remove(s.store.Name()); err != nil {
		return err
	}
	return nil
}

func (s *segment) Close() error {
	if err := s.index.Close(); err != nil {
		return err
	}
	if err := s.store.Close(); err != nil {
		return err
	}
	return nil
}
