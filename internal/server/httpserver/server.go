package httpserver

import (
	"context"
	"errors"
	"github.com/arxon31/metrics-collector/internal/repository/memory"
	handlers2 "github.com/arxon31/metrics-collector/internal/server/handlers"
	middlewares2 "github.com/arxon31/metrics-collector/internal/server/handlers/middlewares"
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
	HashKey         string
}

type Dumper interface {
	Dump(ctx context.Context, path string) error
}

type Restorer interface {
	Restore(ctx context.Context, path string) error
}

func New(p *Params, logger *zap.SugaredLogger, storage handlers2.MetricCollector, provider handlers2.MetricProvider, pinger handlers2.Pinger) *Server {

	mux := chi.NewRouter()
	postGaugeMetricHandler := &handlers2.PostGaugeMetric{Storage: storage, Provider: provider, Logger: logger}
	postCounterMetricHandler := &handlers2.PostCounterMetrics{Storage: storage, Provider: provider, Logger: logger}
	getMetricHandler := &handlers2.GetMetricHandler{Storage: storage, Provider: provider, Logger: logger}
	getMetricsHandler := &handlers2.GetMetricsHandler{Storage: storage, Provider: provider, Logger: logger}
	notImplementedHandler := &handlers2.NotImplementedHandler{Storage: storage, Provider: provider, Logger: logger}
	postJSONHandler := &handlers2.PostJSONMetric{Storage: storage, Provider: provider, Logger: logger}
	getJSONHandler := &handlers2.GetJSONMetric{Storage: storage, Provider: provider, Logger: logger}
	pingHandler := &handlers2.Ping{Pinger: pinger}
	postBatchJSON := &handlers2.PostJSONBatch{Storage: storage, Provider: provider, Logger: logger}

	mux.Post(postGaugeMetricPath, middlewares2.WithLogging(logger, postGaugeMetricHandler).ServeHTTP)
	mux.Post(postCounterMetricPath, middlewares2.WithLogging(logger, postCounterMetricHandler).ServeHTTP)
	mux.Post(postUnknownMetricPath, middlewares2.WithLogging(logger, notImplementedHandler).ServeHTTP)
	mux.Post(postJSONPath, middlewares2.WithLogging(logger, postJSONHandler).ServeHTTP)
	mux.Get(getMetricPath, middlewares2.WithLogging(logger, getMetricHandler).ServeHTTP)
	mux.Get(getMetricsPath, middlewares2.WithLogging(logger, getMetricsHandler).ServeHTTP)
	mux.Post(getJSONPath, middlewares2.WithLogging(logger, getJSONHandler).ServeHTTP)
	mux.Get(pingPath, middlewares2.WithLogging(logger, pingHandler).ServeHTTP)
	mux.Post(postJSONBatch, middlewares2.WithHash(p.HashKey, middlewares2.WithLogging(logger, postBatchJSON)).ServeHTTP)

	return &Server{
		server: &http.Server{
			Addr:    p.Address,
			Handler: middlewares2.WithCompressing(mux),
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
			if errors.Is(err, memory.ErrIsNotFound) {
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
