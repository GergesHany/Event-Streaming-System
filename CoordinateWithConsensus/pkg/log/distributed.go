package log

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"io"

	api "github.com/GergesHany/Event-Streaming-System/ServeRequestsWithgRPC/api/v1"
	SDWPApi "github.com/GergesHany/Event-Streaming-System/StructureDataWithProtobuf/api/v1"

	. "github.com/GergesHany/Event-Streaming-System/WriteALogPackage/log"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb" // implementation of both a LogStore and StableStore.
	"google.golang.org/protobuf/proto"
)

var (
	// Encode data to persist it to a disk
	enc = binary.BigEndian
)

type DistributedLog struct {
	config Config
	log    *Log
	raft   *raft.Raft
}

type fsm struct {
	log *Log
}

type logStore struct {
	*Log
}

type RequestType uint8

type snapshot struct {
	reader io.Reader
}

const (
	AppendRequestType RequestType = 0
)

func NewDistributedLog(dataDir string, config Config) (*DistributedLog, error) {
	l := &DistributedLog{
		config: config,
	}

	if err := l.setupLog(dataDir); err != nil {
		return nil, err
	}

	if err := l.setupRaft(dataDir); err != nil {
		return nil, err
	}

	return l, nil
}

func (l *DistributedLog) setupLog(dataDir string) error {
	logDir := filepath.Join(dataDir, "log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}
	var err error
	l.log, err = NewLog(logDir, l.config)
	return err
}

/*

  A Raft instance comprises:
	1- A finite-state machine that applies the commands you give Raft;
	2- A log store where Raft stores those commands;
	3- A stable store where Raft stores the cluster’s configuration—the servers in the cluster, their addresses, and so on;
	4- A snapshot store where Raft stores compact snapshots of its data; and
	5- A transport that Raft uses to connect with the server’s peers.
*/

func (l *DistributedLog) setupRaft(dataDir string) error {
	// 1- A finite-state machine that applies the commands you give Raft

	fsm := &fsm{log: l.log}
	logDir := filepath.Join(dataDir, "raft", "log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// 2- A log store where Raft stores those commands;

	logConfig := l.config
	logConfig.Segment.InitialOffset = 1
	logStore, err := newLogStore(logDir, logConfig)
	if err != nil {
		return err
	}

	// 3- A stable store where Raft stores the cluster’s configuration—the servers in the cluster, their addresses, and so on;

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft", "stable"))
	if err != nil {
		return err
	}

	// 4- A snapshot store where Raft stores compact snapshots of its data;

	retain := 1
	snapshotStore, err := raft.NewFileSnapshotStore(
		filepath.Join(dataDir, "raft"),
		retain,
		os.Stderr,
	)

	if err != nil {
		return err
	}

	// 5- A transport that Raft uses to connect with the server’s peers.

	/*
	* NewNetworkTransport creates a new network transport with the given dialer and listener.
	* The maxPool controls how many connections we will pool.
	* The timeout is used to apply I/O deadlines. For InstallSnapshot, we multiply the timeout by (SnapshotSize / TimeoutScale).
	 */

	maxPool := 5
	timeout := 10 * time.Second
	transport := raft.NewNetworkTransport(
		l.config.Raft.StreamLayer,
		maxPool,
		timeout,
		os.Stderr,
	)

	config := raft.DefaultConfig()
	config.LocalID = l.config.Raft.LocalID

	if l.config.Raft.HeartbeatTimeout > 0 {
		config.HeartbeatTimeout = l.config.Raft.HeartbeatTimeout
	}

	if l.config.Raft.ElectionTimeout > 0 {
		config.ElectionTimeout = l.config.Raft.ElectionTimeout
	}

	if l.config.Raft.LeaderLeaseTimeout != 0 {
		config.LeaderLeaseTimeout = l.config.Raft.LeaderLeaseTimeout
	}

	// CommitTimeout is a Raft configuration parameter that controls how long Raft will wait for a log entry to be committed before timing out.
	if l.config.Raft.CommitTimeout > 0 {
		config.CommitTimeout = l.config.Raft.CommitTimeout
	}

	l.raft, err = raft.NewRaft(
		config,
		fsm,
		logStore,
		stableStore,
		snapshotStore,
		transport,
	)

	if err != nil {
		return err
	}

	// If there is no existing state, and we are supposed to bootstrap, then we need to bootstrap the cluster.
	if l.config.Raft.Bootstrap {
		config := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		err = l.raft.BootstrapCluster(config).Error()
	}

	return err
}

func newLogStore(dir string, c Config) (*logStore, error) {
	log, err := NewLog(dir, c)
	if err != nil {
		return nil, err
	}
	return &logStore{log}, nil
}

func (l *DistributedLog) Append(record *api.Record) (uint64, error) {
	res, err := l.apply(AppendRequestType, &api.ProduceRequest{Record: record})

	if err != nil {
		return 0, err
	}

	return res.(*api.ProduceResponse).Offset, nil
}

/*
* Apply the command to the Raft log and wait for it to be committed
* The command is contained in buf.Bytes()
* The timeout specifies how long to wait for the command to be committed
* The Apply method returns a Future, which contains the result of the command
* or an error if the command failed to be committed within the timeout
 */

func (l *DistributedLog) apply(reqType RequestType, req proto.Message) (interface{}, error) {
	var buf bytes.Buffer
	_, err := buf.Write([]byte{byte(reqType)})
	if err != nil {
		return nil, err
	}

	b, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(b)
	if err != nil {
		return nil, err
	}

	timeout := 10 * time.Second
	future := l.raft.Apply(buf.Bytes(), timeout)

	if future.Error() != nil {
		return nil, future.Error()
	}

	res := future.Response()
	if err, ok := res.(error); ok {
		return nil, err
	}

	return res, nil
}

func (l *DistributedLog) Read(offset uint64) (*api.Record, error) {
	// Convert the returned record to the correct type
	record, err := l.log.Read(offset)
	if err != nil {
		// Check if it's an offset out of range error and convert to the proper type
		if strings.Contains(err.Error(), "offset out of range") {
			return nil, api.ErrOffsetOutOfRange{Offset: offset}
		}
		return nil, err
	}
	converted := &api.Record{
		Value:  record.Value,
		Offset: record.Offset,
	}
	return converted, nil
}

// Compile-time check!
// If the fsm struct does not implement the raft.FSM interface, the code will not compile.
var _ raft.FSM = (*fsm)(nil) // Finite-State Machine

func (l *fsm) Apply(record *raft.Log) interface{} {
	buf := record.Data
	reqType := RequestType(buf[0])

	switch reqType {
	case AppendRequestType:
		return l.applyAppend(buf[1:])
	}

	return nil
}

func (l *fsm) applyAppend(b []byte) interface{} {
	var req api.ProduceRequest
	err := proto.Unmarshal(b, &req)

	if err != nil {
		return err
	}

	structureRecord := &SDWPApi.Record{
		Value:  req.Record.Value,
		Offset: req.Record.Offset,
	}

	offset, err := l.log.Append(structureRecord)
	if err != nil {
		return err
	}

	return &api.ProduceResponse{Offset: offset}
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	r := f.log.Reader()
	return &snapshot{reader: r}, nil
}

func (f *fsm) Restore(r io.ReadCloser) error {
	b := make([]byte, LenWidth)
	var buf bytes.Buffer

	for i := 0; ; i++ {
		_, err := io.ReadFull(r, b) // Read the length prefix
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		size := int64(enc.Uint64(b))
		if _, err = io.CopyN(&buf, r, size); err != nil {
			return err
		}

		record := &SDWPApi.Record{}
		if err = proto.Unmarshal(buf.Bytes(), record); err != nil {
			return err
		}

		if i == 0 {
			f.log.Config.Segment.InitialOffset = record.Offset
			if err := f.log.Reset(); err != nil {
				return err
			}
		}

		if _, err = f.log.Append(record); err != nil {
			return err
		}

		buf.Reset()
	}
	return nil
}

var _ raft.FSMSnapshot = (*snapshot)(nil)

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := io.Copy(sink, s.reader); err != nil {
		_ = sink.Cancel()
		return err
	}
	return sink.Close()
}

// Release is required to implement raft.FSMSnapshot interface.
// It is called when we are done with the snapshot.
func (s *snapshot) Release() {}

var _ raft.LogStore = (*logStore)(nil)

func (l *logStore) FirstIndex() (uint64, error) {
	return l.LowestOffset()
}

func (l *logStore) LastIndex() (uint64, error) {
	off, err := l.HighestOffset()
	return off, err
}

func (l *logStore) GetLog(index uint64, out *raft.Log) error {
	in, err := l.Read(index)
	if err != nil {
		return err
	}

	out.Data = in.Value
	out.Index = in.Offset
	out.Type = raft.LogType(in.Type)
	out.Term = in.Term

	return nil
}

func (l *logStore) StoreLog(record *raft.Log) error {
	return l.StoreLogs([]*raft.Log{record})
}

func (l *logStore) StoreLogs(records []*raft.Log) error {
	for _, record := range records {
		if _, err := l.Append(&SDWPApi.Record{
			Value: record.Data,
			Term:  record.Term,
			Type:  uint32(record.Type),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (l *logStore) DeleteRange(min, max uint64) error {
	return l.Truncate(max)
}

// The StreamLayer type is responsible for managing the network connections between Raft nodes.
var _ raft.StreamLayer = (*StreamLayer)(nil)

func (l *DistributedLog) Join(id, addr string) error {
	configFuture := l.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return err
	}

	serverID := raft.ServerID(id)
	serverAddr := raft.ServerAddress(addr)

	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == serverID || srv.Address == serverAddr {
			// server has same ID or address, remove it first
			if srv.ID == serverID && srv.Address == serverAddr {
				// server already in the cluster with same ID and address, ignore
				return nil
			}
			// remove any existing server with same ID or address
			removeFuture := l.raft.RemoveServer(srv.ID, 0, 0)
			if err := removeFuture.Error(); err != nil {
				return err
			}
		}
	}

	addFuture := l.raft.AddVoter(serverID, serverAddr, 0, 0)

	return addFuture.Error()
}

func (l *DistributedLog) Leave(id string) error {
	removeFuture := l.raft.RemoveServer(raft.ServerID(id), 0, 0)
	return removeFuture.Error()
}

func (l *DistributedLog) WaitForLeader(timeout time.Duration) error {
	timeoutc := time.After(timeout)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutc:
			return fmt.Errorf("timed out waiting for leader")
		case <-ticker.C:
			if l.raft.Leader() != "" {
				return nil
			}
		}
	}
}

func (l *DistributedLog) Close() error {
	future := l.raft.Shutdown()
	if err := future.Error(); err != nil {
		return err
	}
	return l.log.Close()
}
