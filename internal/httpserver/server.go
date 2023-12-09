package httpserver

import (
	"context"
	"github.com/arxon31/metrics-collector/internal/handlers"
	"github.com/arxon31/metrics-collector/internal/storage/mem"
	"github.com/arxon31/metrics-collector/pkg/e"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	GaugePath   = "/update/gauge/"
	CounterPath = "/update/counter/"
)

func notImplementedHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

type Server struct {
	server *http.Server
	params *Params
}

type Params struct {
	Address string
	Port    string
}

func New(p *Params) *Server {
	st := mem.NewMapStorage()

	mux := http.NewServeMux()
	gaugeHandler := &handlers.GaugeHandler{Storage: st}
	counterHandler := &handlers.CounterHandler{Storage: st}

	mux.Handle(GaugePath, Chain(gaugeHandler, reqParamsCheck))
	mux.Handle(CounterPath, Chain(counterHandler, reqParamsCheck))
	mux.Handle("/update/", http.HandlerFunc(notImplementedHandler))

	return &Server{
		server: &http.Server{
			Addr:    p.Address + ":" + p.Port,
			Handler: mux,
		},
		params: p,
	}
}

func (s *Server) Run() {
	const op = "httpserver.Server.Run()"
	done := make(chan struct{})

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop

		if err := s.server.Shutdown(context.Background()); err != nil {
			log.Fatal(e.Wrap(op, "failed to shutdown server", err))
		}
		close(done)
	}()

	err := s.server.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Fatal(e.Wrap(op, "failed to start server", err))
	}

	<-done

	log.Fatal(op, " server gracefully stopped")
}
