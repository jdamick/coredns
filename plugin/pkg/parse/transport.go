package parse

import (
	"github.com/coredns/coredns/plugin/pkg/transport"
)

// Transport returns the transport defined in s and a string where the
// transport prefix is removed (if there was any). If no transport is defined
// we default to TransportDNS
func Transport(s string) (trans string, addr string) {
	return transport.ParseTransport(s)
}
