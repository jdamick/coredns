package parse

import (
	"testing"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin/pkg/transport"
)

func TestTransport(t *testing.T) {
	for i, test := range []struct {
		input    string
		expected string
	}{
		{"dns://.:53", transport.DNS},
		{"2003::1/64.:53", transport.DNS},
		{"grpc://example.org:1443 ", dnsserver.GRPC},
		{"tls://example.org ", dnsserver.TLS},
		{"https://example.org ", dnsserver.HTTPS},
	} {
		actual, _ := Transport(test.input)
		if actual != test.expected {
			t.Errorf("Test %d: Expected %s but got %s", i, test.expected, actual)
		}
	}
}
