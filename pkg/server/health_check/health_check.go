package health_check

import (
	"context"
	"github.com/heptiolabs/healthcheck"
	"github.com/jackc/pgx/v4/pgxpool"
	v7 "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"main/config"
	"main/pkg/constants"
	"main/pkg/logger"
	"net/http"
	"time"
)

const (
	maxHeaderBytes = 1 << 20
	stackSize      = 1 << 10 // 1 KB
	bodyLimit      = "2M"
	readTimeout    = 15 * time.Second
	writeTimeout   = 15 * time.Second
	gzipLevel      = 5
)

type ServerHealthCheck struct {
	ctx            context.Context
	cfg            *config.Config
	logger         logger.Logger
	ps             *http.Server
	psMetrics      *http.Server
	health         healthcheck.Handler
	postgresClient *pgxpool.Pool
	elasticClient  *v7.Client
}

func NewHealthCheckServer(ctx context.Context, cfg *config.Config, logger logger.Logger, postgresClient *pgxpool.Pool, elasticClient *v7.Client) *ServerHealthCheck {
	health := healthcheck.NewHandler()

	mux := http.NewServeMux()
	ps := &http.Server{
		Handler:      mux,
		Addr:         cfg.Probes.Port,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
	}

	metricsMux := http.NewServeMux()
	metricsMux.Handle(cfg.Probes.PrometheusPath, promhttp.Handler())
	psMetrics := &http.Server{
		Handler:      metricsMux,
		Addr:         cfg.Probes.PrometheusPort,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
	}

	mux.HandleFunc(cfg.Probes.LivenessPath, health.LiveEndpoint)
	mux.HandleFunc(cfg.Probes.ReadinessPath, health.ReadyEndpoint)

	configureHealthCheckEndpoints(ctx, cfg, logger, health, postgresClient, elasticClient)

	return &ServerHealthCheck{
		ctx:            ctx,
		cfg:            cfg,
		logger:         logger,
		ps:             ps,
		psMetrics:      psMetrics,
		health:         health,
		postgresClient: postgresClient,
		elasticClient:  elasticClient,
	}
}

func (s *ServerHealthCheck) RunHealthCheck(ctx context.Context) {

	go func() {
		s.logger.Infof("(%s) Kubernetes probes listening on port: {%s}", s.cfg.ServiceName, s.cfg.Probes.Port)
		if err := s.ps.ListenAndServe(); err != nil {
			s.logger.Errorf("(ListenAndServe) err: {%v}", err)
		}
	}()
}

func (s *ServerHealthCheck) RunMetricsServer() {

	go func() {
		s.logger.Infof("(%s) Prometheus metrics listening on port: {%s}", s.cfg.ServiceName, s.cfg.Probes.PrometheusPort)
		if err := s.psMetrics.ListenAndServe(); err != nil {
			s.logger.Errorf("(RunMetricsServer) err: {%v}", err)
		}
	}()
}

func configureHealthCheckEndpoints(ctx context.Context, cfg *config.Config, logger logger.Logger, health healthcheck.Handler, postgresClient *pgxpool.Pool, elasticClient *v7.Client) {

	health.AddReadinessCheck(constants.MongoDB, healthcheck.AsyncWithContext(ctx, func() error {
		if err := postgresClient.Ping(ctx); err != nil {
			logger.Warnf("(MongoDB Readiness Check) err: {%v}", err)
			return err
		}
		return nil
	}, time.Duration(cfg.Probes.CheckIntervalSeconds)*time.Second))

	health.AddLivenessCheck(constants.MongoDB, healthcheck.AsyncWithContext(ctx, func() error {
		if err := postgresClient.Ping(ctx); err != nil {
			logger.Warnf("(MongoDB Liveness Check) err: {%v}", err)
			return err
		}
		return nil
	}, time.Duration(cfg.Probes.CheckIntervalSeconds)*time.Second))

	health.AddReadinessCheck(constants.ElasticSearch, healthcheck.AsyncWithContext(ctx, func() error {
		_, _, err := elasticClient.Ping(cfg.Elastic.URL).Do(ctx)
		if err != nil {
			logger.Warnf("(ElasticSearch Readiness Check) err: {%v}", err)
			return errors.Wrap(err, "client.Ping")
		}
		return nil
	}, time.Duration(cfg.Probes.CheckIntervalSeconds)*time.Second))

	health.AddLivenessCheck(constants.ElasticSearch, healthcheck.AsyncWithContext(ctx, func() error {
		_, _, err := elasticClient.Ping(cfg.Elastic.URL).Do(ctx)
		if err != nil {
			logger.Warnf("(ElasticSearch Liveness Check) err: {%v}", err)
			return errors.Wrap(err, "client.Ping")
		}
		return nil
	}, time.Duration(cfg.Probes.CheckIntervalSeconds)*time.Second))
}

func (s *ServerHealthCheck) ShutDownHealthCheckServer(ctx context.Context) error {
	return s.ps.Shutdown(ctx)
}

func (s *ServerHealthCheck) ShutDownMetricsServer(ctx context.Context) error {
	return s.psMetrics.Shutdown(ctx)
}
