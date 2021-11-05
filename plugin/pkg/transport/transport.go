package transport

import (
	"net"
	"strings"
)

var (
	transports = make(map[string]string)
)

// RegisterTransport registers a transport scheme and default port
func RegisterTransport(scheme, port string) {
	transports[scheme] = port
}

// ParseTransport returns the transport defined in s and a string where the
// transport prefix is removed (if there was any). If no transport is defined
// we default to TransportDNS
func ParseTransport(s string) (scheme string, addr string) {
	for t := range transports {
		if strings.HasPrefix(s, t+"://") {
			s = strings.TrimSpace(s[len(t+"://"):])
			return t, s
		}
	}
	return DNS, s
}

// TransportPort returns the default port for the scheme provided
func TransportPort(scheme string) string {
	if c, ok := transports[scheme]; ok {
		return c
	}
	return Port
}

// TransportHostPort returns the formatted scheme://host:port for the scheme and host provided
func TransportHostPort(scheme, host string) string {
	port := TransportPort(scheme)
	if scheme == DNS {
		return net.JoinHostPort(host, port)
	}
	return scheme + "://" + net.JoinHostPort(host, port)
}
