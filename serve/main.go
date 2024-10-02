package main

import (
	"crypto/rand"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"

	"bsc.es/colmena/local-torrent/common"
	"github.com/anacrolix/envpprof"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

func newClientConfig() *torrent.ClientConfig {
	cfg := torrent.NewDefaultClientConfig()
	cfg.ListenPort = 0
	cfg.NoDHT = true
	cfg.NoDefaultPortForwarding = true
	cfg.Seed = true
	cfg.Debug = false
	cfg.AcceptPeerConnections = true
	cfg.AlwaysWantConns = true
	cfg.DisableTrackers = true
	return cfg
}

func main() {
	tmpDir := common.SetupTmpFolder()
	defer os.RemoveAll(tmpDir)
	defer envpprof.Stop()
	
	sourceDir := filepath.Join(tmpDir, "source")
	assertNil(os.Mkdir(sourceDir, 0o700))
	f, err := os.Create(filepath.Join(sourceDir, "file"))
	assertNil(err)
	_, err = io.CopyN(f, rand.Reader, 1<<30)
	assertNil(err)
	assertNil(f.Close())
	var info metainfo.Info
	err = info.BuildFromFilePath(f.Name())
	assertNil(err)
	var mi metainfo.MetaInfo
	mi.InfoBytes, err = bencode.Marshal(info)
	assertNil(err)

	clientConfig := newClientConfig()
	clientConfig.DefaultStorage = storage.NewMMap(sourceDir)
	c, err := torrent.NewClient(clientConfig)
	slog.Info("created bt client", slog.Any("listenAddr", c.ListenAddrs()))
	assertNil(err)
	defer c.Close()

	_, err = c.AddTorrent(&mi)
	magnet, err := mi.MagnetV2() 
	slog.Info("torrent magnet link", slog.Any("magnet", magnet))
	
	assertNil(err)
	
	//wait for SIGINT
	sigint_channel := make(chan os.Signal, 1)
	signal.Notify(sigint_channel, os.Interrupt)
	for i := range sigint_channel {
		slog.Info("captured sigint", i)
		break
	}
	
	slog.Info("Server stopped")
}

func assertNil(x any) {
	if x != nil {
		panic(x)
	}
}