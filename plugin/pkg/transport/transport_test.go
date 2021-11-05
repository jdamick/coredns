package transport

import (
	"testing"
)

func init() {
	RegisterTransport("dns", "53")
	RegisterTransport("grpc", "443")
	RegisterTransport("tls", "853")
	RegisterTransport("https", "443")
}

func TestTransport(t *testing.T) {
	for i, test := range []struct {
		input        string
		expected     string
		expectedAddr string
	}{
		{"dns://.:53", DNS, ".:53"},
		{"2003::1/64.:53", DNS, "2003::1/64.:53"},
		{"grpc://example.org:1443 ", GRPC, "example.org:1443"},
		{"tls://example.org ", TLS, "example.org"},
		{"https://example.org ", HTTPS, "example.org"},
	} {
		actual, addr := ParseTransport(test.input)
		if actual != test.expected {
			t.Errorf("Test %d: Expected %s but got %s", i, test.expected, actual)
		}
		if addr != test.expectedAddr {
			t.Errorf("Test %d: Expected addr %s but got %s", i, test.expectedAddr, addr)
		}
	}
}

func TestTransportPort(t *testing.T) {
	for i, test := range []struct {
		scheme   string
		expected string
	}{
		{DNS, "53"},
		{GRPC, "443"},
		{TLS, "853"},
		{HTTPS, "443"},
	} {
		actual := TransportPort(test.scheme)
		if actual != test.expected {
			t.Errorf("Test %d: Expected %s but got %s", i, test.expected, actual)
		}
	}
}

func TestTransportHostPort(t *testing.T) {
	for i, test := range []struct {
		scheme   string
		hostname string
		expected string
	}{
		{DNS, "example.com", "example.com:53"}, // special case for dns
		{GRPC, "example.org", "grpc://example.org:443"},
		{TLS, "example.net", "tls://example.net:853"},
		{HTTPS, "example.co.uk", "https://example.co.uk:443"},
	} {
		actual := TransportHostPort(test.scheme, test.hostname)
		if actual != test.expected {
			t.Errorf("Test %d: Expected %s but got %s", i, test.expected, actual)
		}
	}
}
