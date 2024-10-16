package common

import (
	"net"
	"strconv"
	"strings"
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
	AssertNil(err)
	return IpPortAddr{ip, port}
}