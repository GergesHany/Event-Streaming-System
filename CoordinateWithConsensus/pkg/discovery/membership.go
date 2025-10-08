package discovery

import (
	discovery "github.com/GergesHany/Event-Streaming-System/ServerSideServiceDiscovery/pkg/discovery"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
	"go.uber.org/zap"
)

// Local type alias for Membership to allow method definitions
type Membership discovery.Membership

func (m *Membership) LogError(err error, msg string, member serf.Member) {
	log := m.Logger.Error
	if err == raft.ErrNotLeader {
		log = m.Logger.Debug
	}

	log(
		msg,
		zap.Error(err),
		zap.String("name", member.Name),
		zap.String("rpc_addr", member.Tags["rpc_addr"]),
	)
}
