package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

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
	cfg.DisablePEX = true
	cfg.NoDefaultPortForwarding = true
	cfg.Seed = true
	cfg.Debug = false
	cfg.AcceptPeerConnections = true
	cfg.AlwaysWantConns = true
	cfg.DisableTrackers = true
	return cfg
}

func parsePort(c *torrent.Client) string {
	listenAddrs := c.ListenAddrs()
	first := listenAddrs[0].String()
	return strings.Split(first, ":")[1]
}

var interfc = flag.String("interface", "", "interface used by Zyre")
var filename = flag.String("filename", "randomfile", "filename")

func main() {
	flag.Parse()
	// tmpDir := common.SetupTmpFolder()
	// defer os.RemoveAll(tmpDir)
	defer envpprof.Stop()

	sourceDir := filepath.Join("./", "store")
	mi := createTorrent(*filename, sourceDir)

	clientConfig := newClientConfig()
	clientConfig.DefaultStorage = storage.NewMMap(sourceDir)
	c, err := torrent.NewClient(clientConfig)
	slog.Info("created bt client", slog.Any("listenAddr", c.ListenAddrs()))
	slog.Info("listening", slog.String("port", parsePort(c)))
	common.AssertNil(err)
	defer c.Close()

	t, err := c.AddTorrent(&mi)
	magnet, err := mi.MagnetV2()
	common.AssertNil(err)
	slog.Info("torrent magnet link", slog.Any("magnet", magnet.String()))

	torrents := make(chan string)
	peerFound := make(chan common.FoundPeer)
	common.FindPeers(parsePort(c), *interfc, peerFound, torrents)
	magnetToTorrent := make(map [string]*torrent.Torrent)

	magnetToTorrent[common.ParseMagnetLink(magnet.String())] = t
	torrents <- common.ParseMagnetLink(magnet.String())

	go func () {
		for {
			found := <- peerFound
			u := magnetToTorrent[found.Magnet]
			if u == nil {
				slog.Info("found peer but not interested in torrent", 
					slog.String("magnet", found.Magnet))
					continue
			}
			slog.Info("add peer", slog.Any("ip", found.Peer.Addr.String()), 
				slog.Any("foundMagnet", (found.Magnet)), 
				slog.Any("torrentMagnet", u.InfoHash().HexString()))
			u.AddPeers([]torrent.PeerInfo{found.Peer})
		}
	}()

	//wait for SIGINT
	sigint_channel := make(chan os.Signal, 1)
	signal.Notify(sigint_channel, os.Interrupt)
	for i := range sigint_channel {
		slog.Info("captured sigint", i)
		break
	}

	slog.Info("Server stopped")
}

func createTorrent(filename string, sourceDir string) metainfo.MetaInfo {
	f, err := os.Open(filepath.Join(sourceDir, filename))
	var info metainfo.Info
	err = info.BuildFromFilePath(f.Name())
	common.AssertNil(err)
	var mi metainfo.MetaInfo
	mi.InfoBytes, err = bencode.Marshal(info)
	common.AssertNil(err)
	return mi
}
