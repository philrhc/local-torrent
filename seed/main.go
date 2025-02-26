package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"

	"bsc.es/colmena/local-torrent/common"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

func newClientConfig() *torrent.ClientConfig {
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
	return cfg
}

var filename = flag.String("filename", "", "file to serve")

func main() {
	flag.Parse()

	sourceDir := filepath.Join(".")
	mi := createTorrent(*filename)

	clientConfig := newClientConfig()
	clientConfig.DefaultStorage = storage.NewMMap(sourceDir)
	c, err := torrent.NewClient(clientConfig)
	common.AssertNil(err)
	defer c.Close()

	c.AddTorrent(&mi)
	magnet, err := mi.MagnetV2()
	common.AssertNil(err)
	slog.Info("torrent magnet link", slog.Any("magnet", magnet.String()))

	//wait for SIGINT
	sigint_channel := make(chan os.Signal, 1)
	signal.Notify(sigint_channel, os.Interrupt)
	for i := range sigint_channel {
		slog.Info("captured sigint", i)
		break
	}

	slog.Info("Server stopped")
}

func createTorrent(filename string) metainfo.MetaInfo {
	var info metainfo.Info
	err := info.BuildFromFilePath(filename)
	common.AssertNil(err)
	var mi metainfo.MetaInfo
	mi.InfoBytes, err = bencode.Marshal(info)
	common.AssertNil(err)
	return mi
}
