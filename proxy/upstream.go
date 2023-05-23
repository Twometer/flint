package proxy

import (
	"flint/config"
	"flint/mc"
	"time"
)

const pingTimeout = 30 * time.Second

type upstreamWithStatus struct {
	upstream config.Upstream
	status   mc.ServerStatus
}

type upstreamTracker struct {
	upstreams map[string]upstreamWithStatus
}

func newUpstreamTracker() *upstreamTracker {
	return &upstreamTracker{
		upstreams: make(map[string]upstreamWithStatus),
	}
}

func (ut *upstreamTracker) run() {
	ticker := time.NewTicker(pingTimeout)
	for range ticker.C {
		ut.pingServers()
	}
}

func (ut *upstreamTracker) setUpstreams(upstreams []config.Upstream) {

}

func (ut *upstreamTracker) pingServers() {

}
