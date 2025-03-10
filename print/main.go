package main

import (
	"io"
	"os"
	"log"

	"bsc.es/colmena/local-torrent/common"
	"github.com/anacrolix/torrent"
)

func main() {
	cfg := common.NewClientConfig()
	c, err := torrent.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
	defer c.Close()
	t, _ := c.AddMagnet(os.Args[1])
	<-t.GotInfo()

	r := t.Files()[0].NewReader()
	defer r.Close()

	_, err = io.Copy(os.Stdout, r)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
}