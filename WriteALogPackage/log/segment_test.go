package log

import (
	"io/ioutil"
	"os"
	"testing"

	api "github.com/GergesHany/Event-Streaming-System/StructureDataWithProtobuf/api/v1"
	"github.com/stretchr/testify/require"
	"io"
)

func TestSegment(t *testing.T) {
	dir, _ := ioutil.TempDir("", "segmentTest")
	defer os.RemoveAll(dir)

	want := &api.Record{Value: []byte("hello world")}

	c := Config{}
	c.Segment.MaxStoreBytes = 1024
	c.Segment.MaxIndexBytes = entWidth * 3

	s, err := newSegment(dir, 16, c)
	require.NoError(t, err)

	require.Equal(t, uint64(16), s.nextOffset, s.nextOffset)
	require.False(t, s.IsMaxed())

	// append 3 records to the segment
	// the 4th append should fail because the index is maxed (3 entries)
	// each record is 11 bytes, so the store is not maxed yet (33 < 1024)
	// but the index is maxed (3 * 12 = 36 == 36)
	for i := uint64(0); i < 3; i++ {
		off, err := s.Append(want)
		require.NoError(t, err)
		require.Equal(t, 16+i, off)
		got, err := s.Read(off)
		require.NoError(t, err)
		require.Equal(t, want.Value, got.Value)
	}

	_, err = s.Append(want)
	require.Equal(t, io.EOF, err)

	// maxed index
	require.True(t, s.IsMaxed())

	c.Segment.MaxStoreBytes = uint64(len(want.Value) * 3)
	c.Segment.MaxIndexBytes = 1024

	s, err = newSegment(dir, 16, c)
	require.NoError(t, err)
	// maxed store
	require.True(t, s.IsMaxed())
	err = s.Remove()
	require.NoError(t, err)
	s, err = newSegment(dir, 16, c)
	require.NoError(t, err)
	require.False(t, s.IsMaxed())
}
