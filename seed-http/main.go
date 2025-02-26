package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"

	"bsc.es/colmena/local-torrent/common"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

var filename = flag.String("filename", "", "file to serve")

func main() {
	flag.Parse()

	mi := createTorrent(*filename)
	
	str := mi.HashInfoBytes().HexString()
	bencoded, err := mi.InfoBytes.MarshalBencode()
	if err != nil {
		log.Fatal("could not get infohash")
	}

	url := fmt.Sprintf("http://localhost:8080/metainfo?ih=%s", str)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bencoded))
	if err != nil {
		fmt.Println("Error making POST request:", err.Error())
		return
	}
	defer resp.Body.Close()

	fmt.Println("Response Status:", resp.Status)
}

func createTorrent(filename string) metainfo.MetaInfo {
	var info metainfo.Info
	err := info.BuildFromFilePath(filename)
	common.AssertNil(err)
	var mi metainfo.MetaInfo
	mi.InfoBytes, err = bencode.Marshal(info)
	common.AssertNil(err)
	return mi
}
