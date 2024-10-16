package common

import (
	"context"
	"log/slog"
	"net"
	"strconv"
	"strings"

	"github.com/anacrolix/torrent"
	"github.com/philrhc/zyre"
)

type FoundPeer struct {
	Magnet string
	Peer torrent.PeerInfo
}

//magnet link cannot be directly used as a group name, let's use the urn
func ParseMagnetLink(magnet string) string {
	spec, err := torrent.TorrentSpecFromMagnetUri(magnet)
	AssertNil(err)
	return spec.InfoHash.HexString()
}

func FindPeers(name string, interfc string, peerFound chan FoundPeer, magnet chan string) {
	node := zyre.NewZyre(context.Background())
	node.SetName(name)
	node.SetInterface(interfc)
	//TODO: defer node.Stop()
	err := node.Start()
	AssertNil(err)

	go func() {
		for {
			select {
			case msg := <-node.Events():
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
						Addr: IpPortAddr{IP: net.ParseIP(host), Port: hostport},
					}
					slog.Info("peer found", slog.Any("ip", peer.Addr.String()))
					peerFound <- FoundPeer{Magnet: msg.Group, Peer: peer}
				}
			case toJoin := <-magnet:
				groupId := ParseMagnetLink(toJoin)
				slog.Info("joining group", slog.String("groupId", groupId), slog.String("nodeId", node.Name()))
				node.Join(toJoin)
			}
		}
	}()
}

