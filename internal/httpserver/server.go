package httpserver

import (
	"context"
	"errors"
	"github.com/arxon31/metrics-collector/internal/handlers"
	"github.com/arxon31/metrics-collector/internal/handlers/middlewares"
	"github.com/arxon31/metrics-collector/pkg/e"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
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
	postJsonPath          = "/update/"
	getJsonPath           = "/value/"
	shutdownTimeout       = 3 * time.Second
)

type Server struct {
	server *http.Server
	params *Params
	logger *zap.SugaredLogger
}

type Params struct {
	Address string
}

func New(p *Params, logger *zap.SugaredLogger, storage handlers.MetricCollector, provider handlers.MetricProvider) *Server {

	mux := chi.NewRouter()
	postGaugeMetricHandler := &handlers.PostGaugeMetric{Storage: storage, Provider: provider}
	postCounterMetricHandler := &handlers.PostCounterMetrics{Storage: storage, Provider: provider}
	getMetricHandler := &handlers.GetMetricHandler{Storage: storage, Provider: provider}
	getMetricsHandler := &handlers.GetMetricsHandler{Storage: storage, Provider: provider}
	notImplementedHandler := &handlers.NotImplementedHandler{Storage: storage, Provider: provider}
	postJsonHandler := &handlers.PostJsonMetric{Storage: storage, Provider: provider}
	getJsonHandler := &handlers.GetJsonMetric{Storage: storage, Provider: provider}

	mux.Post(postGaugeMetricPath, middlewares.WithLogging(logger, postGaugeMetricHandler).ServeHTTP)
	mux.Post(postCounterMetricPath, middlewares.WithLogging(logger, postCounterMetricHandler).ServeHTTP)
	mux.Post(postUnknownMetricPath, middlewares.WithLogging(logger, notImplementedHandler).ServeHTTP)
	mux.Post(postJsonPath, middlewares.WithLogging(logger, postJsonHandler).ServeHTTP)
	mux.Get(getMetricPath, middlewares.WithLogging(logger, getMetricHandler).ServeHTTP)
	mux.Get(getMetricsPath, middlewares.WithLogging(logger, getMetricsHandler).ServeHTTP)
	mux.Get(getJsonPath, middlewares.WithLogging(logger, getJsonHandler).ServeHTTP)

	return &Server{
		server: &http.Server{
			Addr:    p.Address,
			Handler: mux,
		},
		params: p,
		logger: logger,
	}
}

func (s *Server) Run(ctx context.Context) {
	const op = "httpserver.Server.Run()"
	go func() {

		err := s.server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			s.logger.Errorln(e.WrapError(op, "failed to start server", err))
		}

	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		log.Println(err)
	}

	s.logger.Infoln(op, " server gracefully stopped")
}
