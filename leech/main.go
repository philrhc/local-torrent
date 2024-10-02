package main

import (
	"flag"
	"log"
	"log/slog"
	"net"
	"os"
	"path/filepath"

	"bsc.es/colmena/local-torrent/common"
	"github.com/anacrolix/envpprof"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
)


func newClientConfig() *torrent.ClientConfig {
	cfg := torrent.NewDefaultClientConfig()
	cfg.ListenPort = 0
	cfg.NoDHT = true
	cfg.NoDefaultPortForwarding = true
	cfg.Seed = true
	cfg.Debug = false
	cfg.AlwaysWantConns = true

	return cfg
}

var host = flag.String("host", "127.0.0.1", "seed host")
var hostport = flag.Int("port", 36191, "seed port")
var magnet = flag.String("magnet", "magnet:?xt=urn:btih:8b16054886998b3cb98a30e9240b8d62dc3362e7&dn=file", "magnet link")

func main() {
	flag.Parse()
	peer := torrent.PeerInfo{
		Addr:   common.IpPortAddr{IP: net.ParseIP(*host), Port: *hostport},
	}

	tmpDir := common.SetupTmpFolder()
	defer envpprof.Stop()
	defer os.RemoveAll(tmpDir)
	sourceDir := filepath.Join(tmpDir, "source")
	clientConfig := newClientConfig()
	clientConfig.DefaultStorage = storage.NewMMap(sourceDir)
	c, _ := torrent.NewClient(clientConfig)
	defer c.Close()
	
	slog.Info("Starting magnet download", slog.String("magnetLink", *magnet))
	t, _ := c.AddMagnet(*magnet)
	slog.Info("Adding peer", slog.Any("add", peer.Addr.String()))
	t.AddPeers([]torrent.PeerInfo{peer})
	<-t.GotInfo()
	t.DownloadAll()
	c.WaitAll()
	log.Print("torrent downloaded")
}

func assertNil(x any) {
	if x != nil {
		panic(x)
	}
}