package common

import "github.com/anacrolix/torrent"

func NewClientConfig() *torrent.ClientConfig {
	cfg := torrent.NewDefaultClientConfig()
	cfg.ListenPort = 0
	cfg.NoDHT = true
	cfg.DisablePEX = true
	cfg.NoDefaultPortForwarding = true
	cfg.Seed = true
	cfg.Debug = false
	cfg.AcceptPeerConnections = true
	cfg.AlwaysWantConns = true
	cfg.DisableTrackers = true
	cfg.LocalServiceDiscovery = torrent.LocalServiceDiscoveryConfig{Enabled: true, Ifi: "enxac91a1741908"}
	return cfg
}