package main

import (
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/GergesHany/Event-Streaming-System/SecurityAndObservability/pkg/config"
	"github.com/GergesHany/Event-Streaming-System/ServerSideServiceDiscovery/pkg/agent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type cfg struct {
	agent.Config
	ServerTLSConfig config.TLSConfig
	PeerTLSConfig   config.TLSConfig
}

type cli struct {
	cfg cfg
}

const (
	systemName = "StreamingSystem"
)

func main() {
	cli := &cli{}

	// Define the root command
	cmd := &cobra.Command{
		Use:     systemName,
		PreRunE: cli.setupConfig,
		RunE:    cli.run,
	}

	if err := setupFlags(cmd); err != nil {
		log.Fatal(err)
	}

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// setupFlags defines command-line flags and binds them to viper.
func setupFlags(cmd *cobra.Command) error {
	hostname, err := os.Hostname() // The host name reported by the kernel.
	if err != nil {
		log.Fatal(err)
	}

	// Configuration file
	cmd.Flags().String("config-file", "", "Path to config file.")

	// Node configuration
	dataDir := path.Join(os.TempDir(), systemName)
	cmd.Flags().String("node-name", hostname, "Unique server ID.")
	cmd.Flags().String("data-dir", dataDir, "Directory to store log and Raft data.")

	// Cluster configuration
	cmd.Flags().Bool("bootstrap", false, "Bootstrap the cluster.")
	cmd.Flags().StringSlice("start-join-addrs", nil, "Serf addresses to join.")
	cmd.Flags().String("bind-addr", "127.0.0.1:8401", "Address to bind Serf on.")
	cmd.Flags().Int("rpc-port", 8400, "Port for RPC clients (and Raft) connections.")

	// Access Control List (ACL) configuration
	cmd.Flags().String("acl-model-file", "", "Path to ACL model.")
	cmd.Flags().String("acl-policy-file", "", "Path to ACL policy.")

	// Server TLS configuration
	cmd.Flags().String("server-tls-cert-file", "", "Path to server tls cert.")
	cmd.Flags().String("server-tls-key-file", "", "Path to server tls key.")
	cmd.Flags().String("server-tls-ca-file", "", "Path to server certificate authority.")

	// Peer TLS configuration
	cmd.Flags().String("peer-tls-cert-file", "", "Path to peer tls cert.")
	cmd.Flags().String("peer-tls-key-file", "", "Path to peer tls key.")
	cmd.Flags().String("peer-tls-ca-file", "", "Path to peer certificate authority.")

	return viper.BindPFlags(cmd.Flags())
}

// setupConfig reads in config file and ENV variables if set.
func (c *cli) setupConfig(cmd *cobra.Command, args []string) error {
	var err error

	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}
	viper.SetConfigFile(configFile)

	if err = viper.ReadInConfig(); err != nil {
		// it's ok if config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// Node configuration
	c.cfg.DataDir = viper.GetString("data-dir")
	c.cfg.NodeName = viper.GetString("node-name")

	// Cluster configuration
	c.cfg.BindAddr = viper.GetString("bind-addr")
	c.cfg.RPCPort = viper.GetInt("rpc-port")
	c.cfg.StartJoinAddrs = viper.GetStringSlice("start-join-addrs")
	c.cfg.Bootstrap = viper.GetBool("bootstrap")

	// ACL configuration
	c.cfg.ACLModelFile = viper.GetString("acl-mode-file")
	c.cfg.ACLPolicyFile = viper.GetString("acl-policy-file")

	// Server TLS configuration
	c.cfg.ServerTLSConfig.CertFile = viper.GetString("server-tls-cert-file")
	c.cfg.ServerTLSConfig.KeyFile = viper.GetString("server-tls-key-file")
	c.cfg.ServerTLSConfig.CAFile = viper.GetString("server-tls-ca-file")

	// Peer TLS configuration
	c.cfg.PeerTLSConfig.CertFile = viper.GetString("peer-tls-cert-file")
	c.cfg.PeerTLSConfig.KeyFile = viper.GetString("peer-tls-key-file")
	c.cfg.PeerTLSConfig.CAFile = viper.GetString("peer-tls-ca-file")

	if c.cfg.ServerTLSConfig.CertFile != "" && c.cfg.ServerTLSConfig.KeyFile != "" {
		c.cfg.ServerTLSConfig.Server = true
		c.cfg.Config.ServerTLSConfig, err = config.SetupTLSConfig(c.cfg.ServerTLSConfig)
		if err != nil {
			return err
		}
	}

	if c.cfg.PeerTLSConfig.CertFile != "" && c.cfg.PeerTLSConfig.KeyFile != "" {
		c.cfg.Config.PeerTLSConfig, err = config.SetupTLSConfig(c.cfg.PeerTLSConfig)
		if err != nil {
			return err
		}
	}

	return nil
}

// run starts the agent and waits for termination signals.
func (c *cli) run(cmd *cobra.Command, args []string) error {
	var err error
	agent, err := agent.New(c.cfg.Config)
	if err != nil {
		return err
	}
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
	return agent.Shutdown()
}
