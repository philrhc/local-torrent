package main

import (
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
)

func parsePort(c *torrent.Client) string {
	listenAddrs := c.ListenAddrs()
	first := listenAddrs[0].String()
	return strings.Split(first, ":")[1]
}

var interfc = flag.String("interface", "", "network interface for peer discovery")
var fileSize = flag.Int64("size", 500, "file size to seed")

func main() {
	flag.Parse()
	tmpDir := common.SetupTmpFolder()
	defer os.RemoveAll(tmpDir)
	defer envpprof.Stop()

	sourceDir := filepath.Join(tmpDir, "source")
	mi := createTorrent(sourceDir, *fileSize)

	clientConfig := common.NewClientConfig()
	clientConfig.DefaultStorage = storage.NewMMap(sourceDir)
	clientConfig.LocalServiceDiscovery = torrent.LocalServiceDiscoveryConfig{Enabled: true, Ifi: *interfc}
	c, err := torrent.NewClient(clientConfig)
	slog.Info("created bt client", slog.Any("listenAddr", c.ListenAddrs()))
	slog.Info("listening", slog.String("port", parsePort(c)))
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

func createTorrent(sourceDir string, size int64) metainfo.MetaInfo {
	common.AssertNil(os.Mkdir(sourceDir, 0o700))
	f, err := os.Create(filepath.Join(sourceDir, "file"))
	common.AssertNil(err)
	_, err = io.CopyN(f, rand.Reader, (size << 20))
	common.AssertNil(err)
	common.AssertNil(f.Close())
	var info metainfo.Info
	err = info.BuildFromFilePath(f.Name())
	common.AssertNil(err)
	var mi metainfo.MetaInfo
	mi.InfoBytes, err = bencode.Marshal(info)
	common.AssertNil(err)
	return mi
}
