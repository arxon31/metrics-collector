package httpserver

import (
	"context"
	"errors"
	"github.com/arxon31/metrics-collector/internal/handlers"
	"github.com/arxon31/metrics-collector/internal/handlers/middlewares"
	"github.com/arxon31/metrics-collector/pkg/e"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	postCounterMetricPath = "/update/counter/{name}/{value}"
	postGaugeMetricPath   = "/update/gauge/{name}/{value}"
	postUnknownMetricPath = "/update/{type}/{name}/{value}"
	getMetricPath         = "/value/{type}/{name}"
	getMetricsPath        = "/"
	postJSONPath          = "/update/"
	getJSONPath           = "/value/"
	shutdownTimeout       = 3 * time.Second
)

var ErrIsNotFound = errors.New("not found")

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
}

func New(p *Params, logger *zap.SugaredLogger, storage handlers.MetricCollector, provider handlers.MetricProvider) *Server {

	mux := chi.NewRouter()
	postGaugeMetricHandler := &handlers.PostGaugeMetric{Storage: storage, Provider: provider}
	postCounterMetricHandler := &handlers.PostCounterMetrics{Storage: storage, Provider: provider}
	getMetricHandler := &handlers.GetMetricHandler{Storage: storage, Provider: provider}
	getMetricsHandler := &handlers.GetMetricsHandler{Storage: storage, Provider: provider}
	notImplementedHandler := &handlers.NotImplementedHandler{Storage: storage, Provider: provider}
	postJSONHandler := &handlers.PostJSONMetric{Storage: storage, Provider: provider}
	getJSONHandler := &handlers.GetJSONMetric{Storage: storage, Provider: provider}

	mux.Post(postGaugeMetricPath, middlewares.WithLogging(logger, postGaugeMetricHandler).ServeHTTP)
	mux.Post(postCounterMetricPath, middlewares.WithLogging(logger, postCounterMetricHandler).ServeHTTP)
	mux.Post(postUnknownMetricPath, middlewares.WithLogging(logger, notImplementedHandler).ServeHTTP)
	mux.Post(postJSONPath, middlewares.WithLogging(logger, postJSONHandler).ServeHTTP)
	mux.Get(getMetricPath, middlewares.WithLogging(logger, getMetricHandler).ServeHTTP)
	mux.Get(getMetricsPath, middlewares.WithLogging(logger, getMetricsHandler).ServeHTTP)
	mux.Post(getJSONPath, middlewares.WithLogging(logger, getJSONHandler).ServeHTTP)

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
		err := s.restore(ctx, s.params.FileStoragePath, restorer)
		if err != nil {
			if errors.Is(err, ErrIsNotFound) {
				s.logger.Infoln(e.WrapError(op, "nothing to restore", err))
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

	go s.dumpRoutine(ctx, dumper)

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		log.Println(err)
	}

	s.logger.Infoln(op, "trying to dump data to file:", s.params.FileStoragePath)
	if err := s.dump(ctx, s.params.FileStoragePath, dumper); err != nil {
		s.logger.Errorln(e.WrapError(op, "failed to dump data", err))
	}

	s.logger.Infoln(op, " server gracefully stopped")
}

type Dumper interface {
	ValuesJSON(ctx context.Context) (string, error)
}

func (s *Server) dump(ctx context.Context, path string, dumper Dumper) error {
	const op = "httpserver.Server.dump()"

	// create directory if not exists
	dir := filepath.Dir(path)
	myPath, err := os.Getwd()
	if err != nil {
		s.logger.Errorln(e.WrapError(op, "failed to get current path", err))
		return err
	}
	dirPath := filepath.Join(myPath, dir)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		s.logger.Errorln(e.WrapError(op, "failed to create directory", err))
		return err
	}
	fPath := filepath.Join(myPath, path)
	file, err := os.OpenFile(fPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	if err != nil {
		s.logger.Errorln(e.WrapError(op, "failed to open file", err))
		return err
	}
	data, err := dumper.ValuesJSON(ctx)
	if err != nil {
		s.logger.Errorln(e.WrapError(op, "failed to dump data", err))
		return err
	}
	n, err := file.Write([]byte(data))
	if err != nil {
		s.logger.Errorln(e.WrapError(op, "failed to write data", err))
		return err

	}
	if n < len(data) {
		s.logger.Errorln(e.WrapError(op, "failed to write all data", err))
		return err
	}
	return nil
}

func (s *Server) dumpRoutine(ctx context.Context, dumper Dumper) {
	const op = "httpserver.Server.dumpRoutine()"
	if s.params.StoreInterval != 0 {
		timer := time.NewTicker(s.params.StoreInterval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				s.logger.Infoln(op, "trying to dump data to file:", s.params.FileStoragePath)
				if err := s.dump(ctx, s.params.FileStoragePath, dumper); err != nil {
					s.logger.Errorln(e.WrapError(op, "failed to dump data", err))
				}
			}
		}
	} else {
		s.logger.Infoln(op, "synchronously dumping data to file:", s.params.FileStoragePath)
		for {
			if err := s.dump(ctx, s.params.FileStoragePath, dumper); err != nil {
				s.logger.Errorln(e.WrapError(op, "failed to dump data synchronously", err))
			}
		}
	}
}

type Restorer interface {
	RestoreFromJSON(ctx context.Context, values string) error
}

func (s *Server) restore(ctx context.Context, path string, restorer Restorer) error {
	const op = "httpserver.Server.restore()"
	// check file existence
	fPath, err := os.Getwd()
	if err != nil {
		s.logger.Errorln(e.WrapError(op, "failed to get current path", err))
		return err
	}
	if _, err := os.Stat(filepath.Join(fPath, path)); errors.Is(err, os.ErrNotExist) {
		s.logger.Infoln(e.WrapError(op, "file not found", err))
		return ErrIsNotFound
	}
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		s.logger.Errorln(e.WrapError(op, "failed to open file", err))
		return err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		s.logger.Errorln(e.WrapError(op, "failed to read data", err))
		return err

	}
	err = restorer.RestoreFromJSON(ctx, string(data))
	if err != nil {
		s.logger.Errorln(e.WrapError(op, "failed to restore data", err))
		return err
	}
	return nil
}
