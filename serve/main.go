package main

import (
	"context"
	"crypto/rand"
	"flag"
	"io"
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
	"github.com/philrhc/zyre"
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

func parsePort(c *torrent.Client) string {
	listenAddrs := c.ListenAddrs()
	first := listenAddrs[0].String()
	return strings.Split(first, ":")[1]
}

var interfc = flag.String("interface", "", "interface used by Zyre")


func main() {
	flag.Parse()
	tmpDir := common.SetupTmpFolder()
	defer os.RemoveAll(tmpDir)
	defer envpprof.Stop()

	sourceDir := filepath.Join(tmpDir, "source")
	mi := createTorrent(sourceDir)

	clientConfig := newClientConfig()
	clientConfig.DefaultStorage = storage.NewMMap(sourceDir)
	c, err := torrent.NewClient(clientConfig)
	slog.Info("created bt client", slog.Any("listenAddr", c.ListenAddrs()))
	slog.Info("listening", slog.String("port", parsePort(c)))
	assertNil(err)
	defer c.Close()

	_, err = c.AddTorrent(&mi)
	magnet, err := mi.MagnetV2()
	assertNil(err)
	slog.Info("torrent magnet link", slog.Any("magnet", magnet.String()))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node := zyre.NewZyre(ctx)
	node.SetName(parsePort(c))
	node.SetInterface(*interfc)
	defer node.Stop()
	err = node.Start()
	assertNil(err)
	node.Join("hello")
	slog.Info("joining group", slog.String("groupId", "hello"), slog.String("nodeId", node.Name()))

	go func() {
		for {
			msg := <-node.Events()
			slog.Info("received", slog.Any("message", msg))
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

func createTorrent(sourceDir string) metainfo.MetaInfo {
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
	return mi
}

func assertNil(x any) {
	if x != nil {
		panic(x)
	}
}
