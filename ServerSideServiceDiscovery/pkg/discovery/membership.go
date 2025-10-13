package discovery

import (
	"net"

	"github.com/hashicorp/serf/serf"
	"go.uber.org/zap"
)

type Config struct {
	NodeName       string            // Unique name for the node
	BindAddr       string            // Address to bind for Serf communication
	Tags           map[string]string // Metadata tags for the node (How can others reach this node)
	StartJoinAddrs []string          // Addresses of existing members to join
}

type Handler interface {
	Join(name, addr string) error
	Leave(name string) error
}

type Membership struct {
	Config
	handler Handler
	serf    *serf.Serf      // Serf instance (manages cluster membership)
	events  chan serf.Event // Channel to receive Serf events
	Logger  *zap.Logger
}

// New creates a new Membership instance and starts the Serf agent.
func New(handler Handler, config Config) (*Membership, error) {
	c := &Membership{
		Config:  config,
		handler: handler,
		Logger:  zap.L().Named("membership"),
	}

	if err := c.setupSerf(); err != nil {
		return nil, err
	}

	return c, nil
}

// setupSerf initializes and starts the Serf agent.
func (m *Membership) setupSerf() (err error) {
	// Resolve the bind address (returns an address of TCP end point.)
	addr, err := net.ResolveTCPAddr("tcp", m.BindAddr)
	if err != nil {
		return err
	}

	// Initialize Serf configuration
	config := serf.DefaultConfig()
	config.Init()

	config.MemberlistConfig.BindAddr = addr.IP.String()
	config.MemberlistConfig.BindPort = addr.Port

	m.events = make(chan serf.Event)
	config.EventCh = m.events

	config.Tags = m.Tags
	config.NodeName = m.Config.NodeName

	m.serf, err = serf.Create(config)

	if err != nil {
		return err
	}

	go m.eventHandler() // Start the event handler to process Serf events

	// Join existing members if any
	if m.StartJoinAddrs != nil {
		_, err := m.serf.Join(m.StartJoinAddrs, true)
		if err != nil {
			return err
		}
	}
	return nil
}

// eventHandler processes incoming Serf events.
func (m *Membership) eventHandler() {
	for e := range m.events {
		switch e.EventType() {
		case serf.EventMemberJoin:
			for _, member := range e.(serf.MemberEvent).Members {
				if m.isLocal(member) {
					continue
				}
				m.handleJoin(member)
			}
		case serf.EventMemberLeave, serf.EventMemberFailed:
			for _, member := range e.(serf.MemberEvent).Members {
				if m.isLocal(member) {
					continue
				}
				m.handleLeave(member)
			}
		}
	}
}

// handleJoin processes a member join event.
func (m *Membership) handleJoin(member serf.Member) {
	if err := m.handler.Join(member.Name, member.Tags["rpc_addr"]); err != nil {
		m.logError(err, "failed to join", member)
	}
}

// handleLeave processes a member leave event.
func (m *Membership) handleLeave(member serf.Member) {
	if err := m.handler.Leave(member.Name); err != nil {
		m.logError(err, "failed to leave", member)
	}
}

// isLocal checks if the given member is the local node
func (m *Membership) isLocal(member serf.Member) bool {
	// LocalMember returns the Member information for the local node
	return m.serf.LocalMember().Name == member.Name
}

func (m *Membership) Members() []serf.Member {
	return m.serf.Members()
}

func (m *Membership) Leave() error {
	return m.serf.Leave()
}

func (m *Membership) logError(err error, msg string, member serf.Member) {
	m.Logger.Error(msg, zap.Error(err), zap.String("name", member.Name), zap.String("rpc_addr", member.Tags["rpc_addr"]))
}
