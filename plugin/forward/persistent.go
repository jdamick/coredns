package forward

import (
	"crypto/tls"
	"sync"
	"time"

	"github.com/miekg/dns"
)

// a persistConn hold the dns.Conn and the last used time.
type persistConn struct {
	c    *dns.Conn
	used time.Time
}

// Transport hold the persistent cache.
type Transport struct {
	minDialTimeout time.Duration
	maxDialTimeout time.Duration
	avgDialTime    int64                      // kind of average time of dial time
	conns          [typeTotalCount]*sync.Pool // Buckets for udp, tcp and tcp-tls.
	connLck        *sync.RWMutex

	expire    time.Duration // After this duration a connection is expired.
	addr      string
	tlsConfig *tls.Config

	discard *sync.Pool
	stop    chan bool
}

func newTransport(addr string) *Transport {
	t := &Transport{
		minDialTimeout: defaultMinDialTimeout,
		maxDialTimeout: defaultMaxDialTimeout,
		avgDialTime:    int64(defaultMaxDialTimeout / 2),
		conns:          [typeTotalCount]*sync.Pool{},
		expire:         defaultExpire,
		addr:           addr,
		discard:        &sync.Pool{},
		stop:           make(chan bool),
		connLck:        &sync.RWMutex{},
	}
	for i := transportType(0); i < typeTotalCount; i++ {
		t.conns[i] = &sync.Pool{}
	}
	return t
}

func poolPersisConn(s *sync.Pool) *persistConn {
	if pc, ok := s.Get().(*persistConn); ok && pc != nil {
		return pc
	}
	return nil
}

const TRIES = 2

func (t *Transport) getConn(proto string) *persistConn {
	transtype := stringToTransportType(proto)
	t.connLck.RLock()
	pclist := t.conns[transtype]
	t.connLck.RUnlock()

	if pclist == nil {
		return nil
	}
	for retry := 0; retry < TRIES; retry++ {
		if pc := poolPersisConn(pclist); pc != nil {
			if time.Since(pc.used) < t.expire {
				return pc
			} else { // expired, remove it.
				t.discard.Put(pc)
			}
		}
	}
	return nil
}

// connManagers manages the persistent connection cache for UDP and TCP.
func (t *Transport) connManager() {
	ticker := time.NewTicker(defaultExpire)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			t.cleanup(false)

		case <-t.stop:
			t.cleanup(true)
			return
		}
	}
}

// closeConns closes connections.
func closeConns(conns ...*persistConn) {
	for _, pc := range conns {
		pc.c.Close()
	}
}

// cleanup removes connections from cache.
func (t *Transport) cleanup(all bool) {
	if all {
		t.connLck.Lock()
		defer t.connLck.Unlock()
		for transtype, stack := range t.conns {
			t.conns[transtype] = &sync.Pool{}
			for c := poolPersisConn(stack); c != nil; c = poolPersisConn(stack) {
				closeConns(c)
			}
		}
	} else {
		for conn := t.discard.Get(); conn != nil; conn = t.discard.Get() {
			closeConns(conn.(*persistConn))
		}
	}
}

// It is hard to pin a value to this, the import thing is to no block forever, losing at cached connection is not terrible.
const yieldTimeout = 25 * time.Millisecond

// Yield returns the connection to transport for reuse.
func (t *Transport) Yield(pc *persistConn) {
	pc.used = time.Now() // update used time

	transtype := t.transportTypeFromConn(pc)
	t.connLck.RLock()
	pclist := t.conns[transtype]
	t.connLck.RUnlock()
	if pclist != nil {
		pclist.Put(pc)
	}
}

// Start starts the transport's connection manager.
func (t *Transport) Start() { go t.connManager() }

// Stop stops the transport's connection manager.
func (t *Transport) Stop() { close(t.stop) }

// SetExpire sets the connection expire time in transport.
func (t *Transport) SetExpire(expire time.Duration) { t.expire = expire }

// SetTLSConfig sets the TLS config in transport.
func (t *Transport) SetTLSConfig(cfg *tls.Config) { t.tlsConfig = cfg }

const (
	defaultExpire         = 1 * time.Second
	defaultMinDialTimeout = 1 * time.Second
	defaultMaxDialTimeout = 30 * time.Second

	// Some resolves might take quite a while, usually (cached) responses are fast. Set to 2s to give us some time to retry a different upstream.
	defaultReadTimeout  = 2 * time.Second
	defaultWriteTimeout = 2 * time.Second
)
