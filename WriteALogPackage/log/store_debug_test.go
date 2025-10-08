package log

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoreAppendReadDebug(t *testing.T) {
	fmt.Println("=== Starting TestStoreAppendReadDebug ===")

	// Create temporary file
	f, err := ioutil.TempFile("", "StoreAppendReadTest")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	fmt.Printf("Created temp file: %s\n", f.Name())

	// Create new store
	s, err := newStore(f)
	require.NoError(t, err)
	fmt.Printf("Created store with initial size: %d\n", s.size)

	// Test append operations
	fmt.Println("\n--- Testing Append Operations ---")
	testAppendDebug(t, s)

	// Test read operations
	fmt.Println("\n--- Testing Read Operations ---")
	testReadDebug(t, s)

	// Test ReadAt operations
	fmt.Println("\n--- Testing ReadAt Operations ---")
	testReadAtDebug(t, s)

	// Create new store from same file
	fmt.Println("\n--- Creating New Store from Same File ---")
	s, err = newStore(f)
	require.NoError(t, err)
	fmt.Printf("New store size: %d\n", s.size)

	// Test read operations again
	fmt.Println("\n--- Testing Read Operations with New Store ---")
	testReadDebug(t, s)

	fmt.Println("=== Test Completed Successfully ===")
}

func testAppendDebug(t *testing.T, s *store) {
	t.Helper()
	fmt.Printf("Data to append: %q (length: %d)\n", write, len(write))
	fmt.Printf("Width (data + length prefix): %d\n", width)

	for i := uint64(1); i < 4; i++ {
		fmt.Printf("\nAppend iteration %d:\n", i)
		n, pos, err := s.Append(write)
		require.NoError(t, err)

		fmt.Printf("  - Position: %d\n", pos)
		fmt.Printf("  - Bytes written: %d\n", n)
		fmt.Printf("  - Store size after: %d\n", s.size)
		fmt.Printf("  - Expected pos+n: %d, Actual: %d\n", width*i, pos+n)

		require.Equal(t, pos+n, width*i)
	}
}

func testReadDebug(t *testing.T, s *store) {
	t.Helper()
	var pos uint64

	for i := 1; i < 4; i++ {
		fmt.Printf("\nRead iteration %d:\n", i)
		fmt.Printf("  - Reading at position: %d\n", pos)

		read, err := s.Read(pos)
		require.NoError(t, err)

		fmt.Printf("  - Data read: %q\n", read)
		fmt.Printf("  - Expected: %q\n", write)

		require.Equal(t, write, read)
		pos += width
	}
}

func testReadAtDebug(t *testing.T, s *store) {
	t.Helper()

	for i, off := uint64(1), int64(0); i < 4; i++ {
		fmt.Printf("\nReadAt iteration %d:\n", i)
		fmt.Printf("  - Starting offset: %d\n", off)

		// Read length prefix
		b := make([]byte, LenWidth)
		n, err := s.ReadAt(b, off)
		require.NoError(t, err)
		require.Equal(t, LenWidth, n)

		size := enc.Uint64(b)
		fmt.Printf("  - Length prefix read: %d bytes\n", size)
		off += int64(n)

		// Read actual data
		b = make([]byte, size)
		n, err = s.ReadAt(b, off)
		require.NoError(t, err)

		fmt.Printf("  - Data read: %q\n", b)
		fmt.Printf("  - Expected: %q\n", write)

		require.Equal(t, write, b)
		require.Equal(t, int(size), n)
		off += int64(n)

		fmt.Printf("  - Next offset: %d\n", off)
	}
}

func TestStoreCloseDebug(t *testing.T) {
	fmt.Println("=== Starting TestStoreCloseDebug ===")

	// Create temporary file
	f, err := ioutil.TempFile("", "StoreCloseTest")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	fmt.Printf("Created temp file: %s\n", f.Name())

	// Create new store
	s, err := newStore(f)
	require.NoError(t, err)
	fmt.Printf("Created store with initial size: %d\n", s.size)

	// Append data (this goes to buffer, not immediately to disk)
	fmt.Println("\n--- Appending Data ---")
	n, pos, err := s.Append(write)
	require.NoError(t, err)
	fmt.Printf("Appended %q at position %d, wrote %d bytes\n", write, pos, n)
	fmt.Printf("Store size in memory: %d\n", s.size)

	// Check file size before close (should be smaller because data is in buffer)
	fmt.Println("\n--- Checking File Size Before Close ---")
	f2, beforeSize, err := openFile(f.Name())
	require.NoError(t, err)
	f2.Close()
	fmt.Printf("File size on disk before close: %d bytes\n", beforeSize)
	fmt.Printf("Store size in memory: %d bytes\n", s.size)
	fmt.Printf("Difference (data in buffer): %d bytes\n", int64(s.size)-beforeSize)

	// Close the store (this flushes buffer to disk)
	fmt.Println("\n--- Closing Store (Flushing Buffer) ---")
	err = s.Close()
	require.NoError(t, err)
	fmt.Println("Store closed successfully")

	// Check file size after close (should be larger because buffer was flushed)
	fmt.Println("\n--- Checking File Size After Close ---")
	f3, afterSize, err := openFile(f.Name())
	require.NoError(t, err)
	f3.Close()
	fmt.Printf("File size on disk after close: %d bytes\n", afterSize)
	fmt.Printf("Size increase: %d bytes\n", afterSize-beforeSize)

	// Verify the assertion
	require.True(t, afterSize > beforeSize)
	fmt.Printf("✅ Assertion passed: afterSize (%d) > beforeSize (%d)\n", afterSize, beforeSize)

	fmt.Println("=== Test Completed Successfully ===")
}
