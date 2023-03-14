package forward

import (
	"testing"
	"time"

	"github.com/coredns/coredns/plugin/pkg/dnstest"

	"github.com/miekg/dns"
)

func BenchmarkPersistentHandoff(b *testing.B) {
	s := dnstest.NewServer(func(w dns.ResponseWriter, r *dns.Msg) {
		ret := new(dns.Msg)
		ret.SetReply(r)
		w.WriteMsg(ret)
	})

	tr := newTransport(s.Addr)
	tr.Start()
	b.Cleanup(func() { tr.Stop(); s.Close() })
	// setup complete, reset timing
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c1 /*cache1*/, _, err := tr.Dial("udp")
			if c1 == nil {
				b.Errorf("c1 is nil")
			}
			if err != nil {
				b.Errorf("Dial Error: %v", err)
			}
			tr.Yield(c1)
		}
	})
}

func TestCached(t *testing.T) {
	s := dnstest.NewServer(func(w dns.ResponseWriter, r *dns.Msg) {
		ret := new(dns.Msg)
		ret.SetReply(r)
		w.WriteMsg(ret)
	})
	defer s.Close()

	tr := newTransport(s.Addr)
	tr.Start()
	defer tr.Stop()

	c1, cache1, _ := tr.Dial("udp")
	c2, cache2, _ := tr.Dial("udp")

	if cache1 || cache2 {
		t.Errorf("Expected non-cached connection")
	}

	tr.Yield(c1)
	tr.Yield(c2)
	c3, cached3, _ := tr.Dial("udp")
	if !cached3 {
		t.Error("Expected cached connection (c3)")
	}
	if c1 != c3 {
		t.Error("Expected c1 == c3")
	}

	tr.Yield(c3)

	// dial another protocol
	c4, cached4, _ := tr.Dial("tcp")
	if cached4 {
		t.Errorf("Expected non-cached connection (c4)")
	}
	tr.Yield(c4)
}

func TestCleanupByTimer(t *testing.T) {
	s := dnstest.NewServer(func(w dns.ResponseWriter, r *dns.Msg) {
		ret := new(dns.Msg)
		ret.SetReply(r)
		w.WriteMsg(ret)
	})
	defer s.Close()

	tr := newTransport(s.Addr)
	tr.SetExpire(100 * time.Millisecond)
	tr.Start()
	defer tr.Stop()

	c1, _, _ := tr.Dial("udp")
	c2, _, _ := tr.Dial("udp")
	tr.Yield(c1)
	time.Sleep(10 * time.Millisecond)
	tr.Yield(c2)

	time.Sleep(120 * time.Millisecond)
	c3, cached, _ := tr.Dial("udp")
	if cached {
		t.Error("Expected non-cached connection (c3)")
	}
	tr.Yield(c3)

	time.Sleep(120 * time.Millisecond)
	c4, cached, _ := tr.Dial("udp")
	if cached {
		t.Error("Expected non-cached connection (c4)")
	}
	tr.Yield(c4)
}

func TestCleanupAll(t *testing.T) {
	s := dnstest.NewServer(func(w dns.ResponseWriter, r *dns.Msg) {
		ret := new(dns.Msg)
		ret.SetReply(r)
		w.WriteMsg(ret)
	})
	defer s.Close()

	tr := newTransport(s.Addr)

	c1, _ := dns.DialTimeout("udp", tr.addr, defaultMaxDialTimeout)
	c2, _ := dns.DialTimeout("udp", tr.addr, defaultMaxDialTimeout)
	c3, _ := dns.DialTimeout("udp", tr.addr, defaultMaxDialTimeout)

	//tr.conns[typeUDP] = []*persistConn{{c1, time.Now()}, {c2, time.Now()}, {c3, time.Now()}}
	tr.conns[typeUDP].Put(&persistConn{c1, time.Now()})
	tr.conns[typeUDP].Put(&persistConn{c2, time.Now()})
	tr.conns[typeUDP].Put(&persistConn{c3, time.Now()})

	// if len(tr.conns[typeUDP]) != 3 {
	// 	t.Error("Expected 3 connections")
	// }
	tr.cleanup(true)

	// if len(tr.conns[typeUDP]) > 0 {
	// if tr.conns[typeUDP].Get() != nil {
	// 	t.Error("Expected no cached connections")
	// }
}
