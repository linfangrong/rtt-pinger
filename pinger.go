package pinger

import (
	"net"
	"time"

	"github.com/tatsushid/go-fastping"
)

type RttPinger struct {
	pinger         *fastping.Pinger
	rttDataManager *RttDataManager
}

func NewRttPinger(maxRtt time.Duration) (s *RttPinger) {
	s = &RttPinger{
		pinger:         fastping.NewPinger(),
		rttDataManager: NewRttDataManager(),
	}
	s.pinger.MaxRTT = maxRtt
	s.pinger.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		s.rttDataManager.Add(addr.String(), rtt)
	}
	return
}

func (s *RttPinger) Pinger() (pinger *fastping.Pinger) {
	return s.pinger
}

func (s *RttPinger) Data() (rttDataManager *RttDataManager) {
	return s.rttDataManager
}
