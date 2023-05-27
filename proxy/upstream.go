package proxy

import (
	"flint/config"
	"flint/mc"
	"strings"
	"sync"
	"time"
)

const pingTimeout = 30 * time.Second

type upstreamWithStatus struct {
	config config.Upstream
	status mc.ServerStatus
}

type upstreamTracker struct {
	upstreamMutex sync.Mutex
	upstreams     map[string]upstreamWithStatus
	ticker        *time.Ticker
	pingRequests  chan interface{}
}

func newUpstreamTracker() *upstreamTracker {
	tracker := &upstreamTracker{
		upstreamMutex: sync.Mutex{},
		upstreams:     make(map[string]upstreamWithStatus),
		ticker:        time.NewTicker(pingTimeout),
		pingRequests:  make(chan interface{}, 4),
	}
	go tracker.run()
	tracker.schedulePing()
	return tracker
}

func (ut *upstreamTracker) run() {
	defer ut.ticker.Stop()

	for {
		select {
		case <-ut.ticker.C:
			ut.pingServers()
		case <-ut.pingRequests:
			ut.pingServers()
		}
	}
}

func (ut *upstreamTracker) schedulePing() {
	ut.ticker.Reset(pingTimeout)
	ut.pingRequests <- 0
}

func (ut *upstreamTracker) setUpstreams(upstreams []config.Upstream) {
	ut.upstreamMutex.Lock()
	defer ut.upstreamMutex.Unlock()

	ut.upstreams = make(map[string]upstreamWithStatus)
	for _, upstream := range upstreams {
		ut.upstreams[upstream.Host] = upstreamWithStatus{
			status: mc.ServerOffline,
			config: upstream,
		}
	}

	ut.schedulePing()
}

func (ut *upstreamTracker) pingServers() {
	ut.upstreamMutex.Lock()
	defer ut.upstreamMutex.Unlock()

	for host, upstream := range ut.upstreams {
		upstream.status = mc.PingServer(upstream.config.Address)
		ut.upstreams[strings.ToLower(host)] = upstream
	}
}

func (ut *upstreamTracker) findUpstream(host string) (upstreamWithStatus, bool) {
	us, ok := ut.upstreams[strings.ToLower(host)]
	return us, ok
}
