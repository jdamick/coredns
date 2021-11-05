package transport

import (
	"net"
	"strings"
)

// These transports are supported by CoreDNS.
const (
	DNS = "dns"

//	TLS   = "tls"
//	GRPC  = "grpc"
//	HTTPS = "https"
)

// Port numbers for the various transports.
const (
	// Port is the default port for DNS
	Port = "53"
	// TLSPort is the default port for DNS-over-TLS.
	//TLSPort = "853"
	// GRPCPort is the default port for DNS-over-gRPC.
	//GRPCPort = "443"
	// HTTPSPort is the default port for DNS-over-HTTPS.
	//HTTPSPort = "443"
)

var (
	transports = make(map[string]string)
)

func RegisterTransport(scheme, port string) {
	transports[scheme] = port
}

func ParseTransport(s string) (trans string, addr string) {
	for t := range transports {
		if strings.HasPrefix(s, t+"://") {
			return t, s
		}
	}
	return DNS, s
}

func TransportPort(scheme string) string {
	if c, ok := transports[scheme]; ok {
		return c
	}
	return Port
}

func TransportHostPort(scheme, host string) string {
	port := TransportPort(scheme)
	if scheme == DNS {
		return net.JoinHostPort(host, port)
	}
	return scheme + "://" + net.JoinHostPort(host, port)
}
