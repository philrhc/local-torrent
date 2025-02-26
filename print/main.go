package main

import (
	"io"
	"net"
	"os"
	"log"

	"bsc.es/colmena/local-torrent/common"
	"github.com/anacrolix/torrent"
)

func main() {
	cfg := torrent.NewDefaultClientConfig()
	cfg.ListenPort = 0
	cfg.Seed = true
	cfg.AlwaysWantConns = true
	cfg.AcceptPeerConnections = true
	c, _ := torrent.NewClient(cfg)
	defer c.Close()
	t, _ := c.AddMagnet(os.Args[1])
	t.AddPeers([]torrent.PeerInfo{torrent.PeerInfo{Addr: common.IpPortAddr{IP: net.ParseIP("127.0.0.1"), Port: 42069}}})
	<-t.GotInfo()

	r := t.Files()[0].NewReader()
	defer r.Close()

	_, err := io.Copy(os.Stdout, r)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
}