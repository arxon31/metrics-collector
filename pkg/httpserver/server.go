package httpserver

import (
	"context"
	"net/http"
	"time"
)

const (
	_defaultShutdownTimeout = 10 * time.Second
	_defaultAddr            = ":80"
)

type server struct {
	server          *http.Server
	notify          chan error
	shutdownTimeout time.Duration
}

func NewHTTPServer(handler http.Handler, opts ...Option) *server {
	httpServer := &http.Server{
		Addr:    _defaultAddr,
		Handler: handler,
	}

	s := server{
		server:          httpServer,
		notify:          make(chan error, 1),
		shutdownTimeout: _defaultShutdownTimeout,
	}

	for _, opt := range opts {
		opt(&s)
	}

	s.start()

	return &s

}

func (s *server) start() {
	go func() {
		s.notify <- s.server.ListenAndServe()
		close(s.notify)
	}()
}

func (s *server) Notify() chan error {
	return s.notify
}

func (s *server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}
