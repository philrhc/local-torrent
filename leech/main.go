package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
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
var thenseed = flag.String("thenseed", "false", "seed torrent after downloading")

func parsePort(c *torrent.Client) string {
	listenAddrs := c.ListenAddrs()
	first := listenAddrs[0].String()
	return strings.Split(first, ":")[1]
}

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
	node.SetName(parsePort(c))
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
				if err != nil {
					slog.Info("could not parse port number from peer name", slog.String("name", msg.PeerName))
					continue
				}
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

	thenseedParsed, err := strconv.ParseBool(*thenseed)
	assertNil(err)

	if thenseedParsed {
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

func assertNil(x any) {
	if x != nil {
		panic(x)
	}
}
