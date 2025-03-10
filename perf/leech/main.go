package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"bsc.es/colmena/local-torrent/common"
	"github.com/anacrolix/envpprof"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
)

var interfc = flag.String("interface", "", "network interface for peer discovery")
var magnet = flag.String("magnet", "", "magnet link for download")
var thenseed = flag.Bool("thenseed", false, "seed torrent after downloading")
var debug = flag.Bool("debug", false, "debug logging")

func main() {
	slog.Info("started")
	flag.Parse()

	tmpDir := common.SetupTmpFolder()
	defer envpprof.Stop()
	defer os.RemoveAll(tmpDir)
	sourceDir := filepath.Join(tmpDir, "source")
	clientConfig := common.NewClientConfig()
	clientConfig.DefaultStorage = storage.NewMMap(sourceDir)
	clientConfig.LocalServiceDiscovery = torrent.LocalServiceDiscoveryConfig{Enabled: true, Ifi: *interfc}
	c, _ := torrent.NewClient(clientConfig)
	defer c.Close()

	slog.Info("starting magnet download", slog.String("magnetLink", *magnet))
	t, _ := c.AddMagnet(*magnet)

	if *debug {
		go func ()  {
			for {
				c.WriteStatus(os.Stdout)
				time.Sleep(30 * time.Second)
			}

		}()
	}

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
