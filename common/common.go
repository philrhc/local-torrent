package common

import (
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/anacrolix/envpprof"
)

type IpPortAddr struct {
	IP   net.IP
	Port int
}

func (IpPortAddr) Network() string {
	return ""
}

func (me IpPortAddr) String() string {
	return net.JoinHostPort(me.IP.String(), strconv.FormatInt(int64(me.Port), 10))
}

func ParseIpAddress(ipAddress string) IpPortAddr {
	split := strings.Split(ipAddress, ":")
	ip := net.ParseIP(split[0])
	port, err := strconv.Atoi(split[1])
	assertNil(err)
	return IpPortAddr{ip, port}
}

func SetupTmpFolder() string {
	defer envpprof.Stop()
	tmpDir, err := os.MkdirTemp("", "torrent-zyre")
	assertNil(err)
	slog.Info("made temp dir", slog.String("tmpDir", tmpDir))
	return tmpDir
}

func assertNil(x any) {
	if x != nil {
		panic(x)
	}
}