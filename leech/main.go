package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"

	"bsc.es/colmena/local-torrent/common"
	"github.com/anacrolix/envpprof"
	"github.com/anacrolix/torrent"
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
	cfg.AlwaysWantConns = true
	return cfg
}

var interfc = flag.String("interface", "", "interface used by Zyre")
var magnet = flag.String("magnet", "", "magnet link for download")
var thenseed = flag.Bool("thenseed", false, "seed torrent after downloading")

func parsePort(c *torrent.Client) string {
	listenAddrs := c.ListenAddrs()
	first := listenAddrs[0].String()
	return strings.Split(first, ":")[1]
}

func main() {
	slog.Info("started")
	flag.Parse()

	tmpDir := common.SetupTmpFolder()
	defer envpprof.Stop()
	defer os.RemoveAll(tmpDir)
	sourceDir := filepath.Join(tmpDir, "source")
	clientConfig := newClientConfig()
	clientConfig.DefaultStorage = storage.NewMMap(sourceDir)
	c, _ := torrent.NewClient(clientConfig)
	defer c.Close()

	slog.Info("starting magnet download", slog.String("magnetLink", *magnet))
	t, _ := c.AddMagnet(*magnet)

	//indicate interest in a magnet
	torrents := make(chan string)
	peerFound := make(chan common.FoundPeer)
	common.FindPeers(parsePort(c), *interfc, peerFound, torrents)
	magnetToTorrent := make(map [string]*torrent.Torrent)

	magnetToTorrent[common.ParseMagnetLink(*magnet)] = t
	torrents <- common.ParseMagnetLink(*magnet)
	
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

	<-t.GotInfo()
	t.DownloadAll()
	c.WaitAll()
	log.Print("torrent downloaded")

	if *thenseed {
		//wait for SIGINT
		sigint_channel := make(chan os.Signal, 1)
		signal.Notify(sigint_channel, os.Interrupt)
		for i := range sigint_channel {
			slog.Info("captured sigint", i)
			break
		}

		slog.Info("Server stopped")
	}
}
