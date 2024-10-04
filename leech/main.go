package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"bsc.es/colmena/local-torrent/common"
	"github.com/anacrolix/envpprof"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
	"github.com/philrhc/zyre"
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

var interfc = flag.String("interface", "", "interface used by Zyre")
var magnet = flag.String("magnet", "", "magnet link for download")

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	node := zyre.NewZyre(ctx)
	node.SetInterface(*interfc)
	defer node.Stop()
	err := node.Start()
	assertNil(err)
	slog.Info("joining group", slog.String("groupId", "hello"), slog.String("nodeId", node.Name()))
	node.Join("hello")

	go func() {
		for {
			msg := <-node.Events()
			slog.Info("received", slog.Any("message", msg))
			if msg.Type == "JOIN" && msg.PeerName != "" {
				protocolRemoved := strings.TrimPrefix(msg.PeerAddr, "tcp://")
				split := strings.Split(protocolRemoved, ":")
				host := split[0]
				hostport, err := strconv.Atoi(msg.PeerName)
				assertNil(err)
				peer := torrent.PeerInfo{
					Addr: common.IpPortAddr{IP: net.ParseIP(host), Port: hostport},
				}
				slog.Info("Adding peer", slog.Any("ip", peer.Addr.String()))
				t.AddPeers([]torrent.PeerInfo{peer})
			}
		}
	}()

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
