package httpserver

import (
	"context"
	"errors"
	"github.com/arxon31/metrics-collector/internal/handlers"
	"github.com/arxon31/metrics-collector/internal/handlers/middlewares"
	"github.com/arxon31/metrics-collector/internal/storage/mem"
	"github.com/arxon31/metrics-collector/pkg/e"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const (
	postCounterMetricPath = "/update/counter/{name}/{value}"
	postGaugeMetricPath   = "/update/gauge/{name}/{value}"
	postUnknownMetricPath = "/update/{type}/{name}/{value}"
	getMetricPath         = "/value/{type}/{name}"
	getMetricsPath        = "/"
	postJSONPath          = "/update/"
	postJSONBatch         = "/updates/"
	getJSONPath           = "/value/"
	pingPath              = "/ping"
	shutdownTimeout       = 3 * time.Second
)

type Server struct {
	server *http.Server
	params *Params
	logger *zap.SugaredLogger
}

type Params struct {
	Address         string
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
	DBString        string
}

type Dumper interface {
	Dump(ctx context.Context, path string) error
}

type Restorer interface {
	Restore(ctx context.Context, path string) error
}

func New(p *Params, logger *zap.SugaredLogger, storage handlers.MetricCollector, provider handlers.MetricProvider, pinger handlers.Pinger) *Server {

	mux := chi.NewRouter()
	postGaugeMetricHandler := &handlers.PostGaugeMetric{Storage: storage, Provider: provider, Logger: logger}
	postCounterMetricHandler := &handlers.PostCounterMetrics{Storage: storage, Provider: provider, Logger: logger}
	getMetricHandler := &handlers.GetMetricHandler{Storage: storage, Provider: provider, Logger: logger}
	getMetricsHandler := &handlers.GetMetricsHandler{Storage: storage, Provider: provider, Logger: logger}
	notImplementedHandler := &handlers.NotImplementedHandler{Storage: storage, Provider: provider, Logger: logger}
	postJSONHandler := &handlers.PostJSONMetric{Storage: storage, Provider: provider, Logger: logger}
	getJSONHandler := &handlers.GetJSONMetric{Storage: storage, Provider: provider, Logger: logger}
	pingHandler := &handlers.Ping{Pinger: pinger}
	postBatchJSON := &handlers.PostJSONBatch{Storage: storage, Provider: provider, Logger: logger}

	mux.Post(postGaugeMetricPath, middlewares.WithLogging(logger, postGaugeMetricHandler).ServeHTTP)
	mux.Post(postCounterMetricPath, middlewares.WithLogging(logger, postCounterMetricHandler).ServeHTTP)
	mux.Post(postUnknownMetricPath, middlewares.WithLogging(logger, notImplementedHandler).ServeHTTP)
	mux.Post(postJSONPath, middlewares.WithLogging(logger, postJSONHandler).ServeHTTP)
	mux.Get(getMetricPath, middlewares.WithLogging(logger, getMetricHandler).ServeHTTP)
	mux.Get(getMetricsPath, middlewares.WithLogging(logger, getMetricsHandler).ServeHTTP)
	mux.Post(getJSONPath, middlewares.WithLogging(logger, getJSONHandler).ServeHTTP)
	mux.Get(pingPath, middlewares.WithLogging(logger, pingHandler).ServeHTTP)
	mux.Post(postJSONBatch, middlewares.WithLogging(logger, postBatchJSON).ServeHTTP)

	return &Server{
		server: &http.Server{
			Addr:    p.Address,
			Handler: middlewares.WithCompressing(mux),
		},
		params: p,
		logger: logger,
	}
}

func (s *Server) Run(ctx context.Context, restorer Restorer, dumper Dumper) {
	const op = "httpserver.Server.Run()"

	if s.params.Restore {
		s.logger.Infoln(op, "trying to restore data from file:", s.params.FileStoragePath)
		err := restorer.Restore(ctx, s.params.FileStoragePath)
		if err != nil {
			if errors.Is(err, mem.ErrIsNotFound) {
				s.logger.Infoln(e.WrapString(op, "nothing to restore", err))
			} else {
				s.logger.Errorln(e.WrapError(op, "failed to restore data", err))
			}
		}
	}

	go func() {

		err := s.server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			s.logger.Errorln(e.WrapError(op, "failed to start server", err))
		}

	}()
	if s.params.FileStoragePath != "" {
		if s.params.StoreInterval != 0 {
			timer := time.NewTicker(s.params.StoreInterval)
			for {
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
					s.logger.Infoln(op, "trying to dump data to file:", s.params.FileStoragePath)
					if err := dumper.Dump(ctx, s.params.FileStoragePath); err != nil {
						s.logger.Errorln(e.WrapError(op, "failed to dump data", err))
					}
				}
			}
		} else {
			s.logger.Infoln(op, "synchronously dumping data to file:", s.params.FileStoragePath)
			for {
				if err := dumper.Dump(ctx, s.params.FileStoragePath); err != nil {
					s.logger.Errorln(e.WrapError(op, "failed to dump data synchronously", err))
				}
			}
		}
	}

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Errorln(e.WrapError(op, "failed to shutdown http server", err))
	}
	if s.params.FileStoragePath != "" {
		s.logger.Infoln(op, "trying to dump data to file:", s.params.FileStoragePath)
		if err := dumper.Dump(ctx, s.params.FileStoragePath); err != nil {
			s.logger.Errorln(e.WrapError(op, "failed to dump data", err))
		}
	}

	s.logger.Infoln(op, " server gracefully stopped")
}
