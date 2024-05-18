package httpserver

import "time"

type Option func(s *server)

func (s *server) WithShutdownTimeout(t time.Duration) {
	s.shutdownTimeout = t
}

func (s *server) WithAddr(addr string) {
	s.server.Addr = addr

}
