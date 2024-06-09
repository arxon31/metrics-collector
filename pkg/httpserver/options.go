package httpserver

import "time"

type Option func(s *server)

func WithShutdownTimeout(t time.Duration) Option {
	return func(s *server) {
		s.shutdownTimeout = t
	}
}

func WithAddr(addr string) Option {
	return func(s *server) {
		s.server.Addr = addr
	}

}
