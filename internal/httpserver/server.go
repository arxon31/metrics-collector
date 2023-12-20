package httpserver

import (
	"context"
	"errors"
	"github.com/arxon31/metrics-collector/internal/handlers"
	"github.com/arxon31/metrics-collector/pkg/e"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"time"
)

const (
	postCounterMetricPath = "/update/counter/{name}/{value}"
	postGaugeMetricPath   = "/update/gauge/{name}/{value}"
	postUnknownMetricPath = "/update/{type}/{name}/{value}"
	getMetricPath         = "/value/{type}/{name}"
	getMetricsPath        = "/"
	shutdownTimeout       = 3 * time.Second
)

type Server struct {
	server *http.Server
	params *Params
}

type Params struct {
	Address string
}

func New(p *Params, storage handlers.MetricCollector, provider handlers.MetricProvider) *Server {

	mux := chi.NewRouter()
	postGaugeMetricHandler := &handlers.PostGaugeMetric{Storage: storage, Provider: provider}
	postCounterMetricHandler := &handlers.PostCounterMetrics{Storage: storage, Provider: provider}
	getMetricHandler := &handlers.GetMetricHandler{Storage: storage, Provider: provider}
	getMetricsHandler := &handlers.GetMetricsHandler{Storage: storage, Provider: provider}
	notImplementedHandler := &handlers.NotImplementedHandler{Storage: storage, Provider: provider}

	mux.Post(postGaugeMetricPath, postGaugeMetricHandler.ServeHTTP)
	mux.Post(postCounterMetricPath, postCounterMetricHandler.ServeHTTP)
	mux.Post(postUnknownMetricPath, notImplementedHandler.ServeHTTP)
	mux.Get(getMetricPath, getMetricHandler.ServeHTTP)
	mux.Get(getMetricsPath, getMetricsHandler.ServeHTTP)

	return &Server{
		server: &http.Server{
			Addr:    p.Address,
			Handler: mux,
		},
		params: p,
	}
}

func (s *Server) Run(ctx context.Context) {
	const op = "httpserver.Server.Run()"
	go func() {

		err := s.server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(e.Wrap(op, "failed to start server", err))
		}

	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		log.Println(err)
	}

	log.Println(op, " server gracefully stopped")
}
