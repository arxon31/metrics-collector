// Package httpserver provides http server implementation
package server

import (
	"context"
	"errors"
	"github.com/arxon31/metrics-collector/internal/repository"
	"github.com/arxon31/metrics-collector/internal/server/config"
	middlewares2 "github.com/arxon31/metrics-collector/internal/server/controller/http/middlewares"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/arxon31/metrics-collector/internal/server/handlers"
	"github.com/arxon31/metrics-collector/pkg/e"
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
	storage repository.Repository
	server  *http.Server
	config  *config.Config
	logger  *zap.SugaredLogger
}

type Dumper interface {
	Dump(ctx context.Context, path string) error
}

type Restorer interface {
	Restore(ctx context.Context, path string) error
}

func New(cfg *config.Config, logger *zap.SugaredLogger, repo repository.Repository) *Server {

	mux := chi.NewRouter()

	mux.Mount("/debug", middleware.Profiler())

	postGaugeMetricHandler := &handlers.PostGaugeMetric{Storage: repo, Provider: repo, Logger: logger}
	postCounterMetricHandler := &handlers.PostCounterMetrics{Storage: repo, Provider: repo, Logger: logger}
	getMetricHandler := &handlers.GetMetricHandler{Storage: repo, Provider: repo, Logger: logger}
	getMetricsHandler := &handlers.GetMetricsHandler{Storage: repo, Provider: repo, Logger: logger}
	notImplementedHandler := &handlers.NotImplementedHandler{Storage: repo, Provider: repo, Logger: logger}
	postJSONHandler := &handlers.PostJSONMetric{Storage: repo, Provider: repo, Logger: logger}
	getJSONHandler := &handlers.GetJSONMetric{Storage: repo, Provider: repo, Logger: logger}
	pingHandler := &handlers.Ping{Pinger: repo}
	postBatchJSON := &handlers.PostJSONBatch{Storage: repo, Provider: repo, Logger: logger}

	mux.Post(postGaugeMetricPath, middlewares2.WithLogging(logger, postGaugeMetricHandler).ServeHTTP)
	mux.Post(postCounterMetricPath, middlewares2.WithLogging(logger, postCounterMetricHandler).ServeHTTP)
	mux.Post(postUnknownMetricPath, middlewares2.WithLogging(logger, notImplementedHandler).ServeHTTP)
	mux.Post(postJSONPath, middlewares2.WithLogging(logger, postJSONHandler).ServeHTTP)
	mux.Get(getMetricPath, middlewares2.WithLogging(logger, getMetricHandler).ServeHTTP)
	mux.Get(getMetricsPath, middlewares2.WithLogging(logger, getMetricsHandler).ServeHTTP)
	mux.Post(getJSONPath, middlewares2.WithLogging(logger, getJSONHandler).ServeHTTP)
	mux.Get(pingPath, middlewares2.WithLogging(logger, pingHandler).ServeHTTP)
	mux.Post(postJSONBatch, middlewares2.WithHash(cfg.HashKey, middlewares2.WithLogging(logger, postBatchJSON)).ServeHTTP)

	return &Server{
		server: &http.Server{
			Addr:    cfg.Address,
			Handler: middlewares2.WithCompressing(mux),
		},
		storage: repo,
		config:  cfg,
		logger:  logger,
	}
}

// Run function starts http server
func (s *Server) Run(ctx context.Context) {
	const op = "httpserver.Server.Run()"

	if s.config.Restore {
		s.logger.Infoln(op, "trying to restore data from file:", s.config.FileStoragePath)
		err := s.storage.Restore(ctx, s.config.FileStoragePath)
		if err != nil {
			if errors.Is(err, repository.ErrFileNotFound) {
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
	if s.config.FileStoragePath != "" {
		if s.config.StoreInterval != 0 {
			timer := time.NewTicker(s.config.StoreInterval)
			for {
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
					s.logger.Infoln(op, "trying to dump data to file:", s.config.FileStoragePath)
					if err := s.storage.Dump(ctx, s.config.FileStoragePath); err != nil {
						s.logger.Errorln(e.WrapError(op, "failed to dump data", err))
					}
				}
			}
		} else {
			s.logger.Infoln(op, "synchronously dumping data to file:", s.config.FileStoragePath)
			for {
				if err := s.storage.Dump(ctx, s.config.FileStoragePath); err != nil {
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
	if s.config.FileStoragePath != "" {
		s.logger.Infoln(op, "trying to dump data to file:", s.config.FileStoragePath)
		if err := s.storage.Dump(ctx, s.config.FileStoragePath); err != nil {
			s.logger.Errorln(e.WrapError(op, "failed to dump data", err))
		}
	}

	s.logger.Infoln(op, " server gracefully stopped")
}
