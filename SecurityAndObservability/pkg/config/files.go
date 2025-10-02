package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/joho/godotenv"
)

var (
	CAFile               = configFile("ca.pem")
	ServerCertFile       = configFile("server.pem")
	ServerKeyFile        = configFile("server-key.pem")
	RootClientCertFile   = configFile("root-client.pem")
	RootClientKeyFile    = configFile("root-client-key.pem")
	NobodyClientCertFile = configFile("nobody-client.pem")
	NobodyClientKeyFile  = configFile("nobody-client-key.pem")
	ACLModelFile         = configFile("model.conf")
	ACLPolicyFile        = configFile("policy.csv")
)

// Singleton pattern to load workspace root only once

var once sync.Once       // ensure loadWorkspaceRoot is called only once
var workspaceRoot string // holds the workspace root path

func loadWorkspaceRoot() string {
	once.Do(func() {
		godotenv.Load("../../../.env")
		workspaceRoot = os.Getenv("WORKSPACE_ROOT")
	})
	return workspaceRoot
}

func configFile(filename string) string {
	var root string = loadWorkspaceRoot()
	return filepath.Join(root, "SecurityAndObservability", "test", filename)
}
