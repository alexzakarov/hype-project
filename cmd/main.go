package main

import (
	"context"
	"fmt"
	"log"
	"main/config"
	"main/internal/metrics"
	v1 "main/internal/order/delivery/http/v1"
	"main/internal/order/projection/elastic_projection"
	"main/internal/order/projection/postgres_projection"
	"main/internal/order/repository"
	"main/internal/order/service"
	"main/pkg/databases/elastic"
	"main/pkg/databases/postgres"
	"main/pkg/es/store"
	"main/pkg/eventstroredb"
	"main/pkg/logger"
	"main/pkg/middlewares/trace"
	"main/pkg/server"
	"main/pkg/server/grpc"
	"main/pkg/server/health_check"
	"main/pkg/server/http"
	"main/pkg/utils/common"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	appLogger := logger.NewAppLogger(cfg.Logger)
	appLogger.InitLogger()
	appLogger.WithName(common.GetMicroserviceName(cfg))

	metricsCounters := metrics.NewESMicroserviceMetrics(cfg)

	postgresDb, errPostgres := postgres.NewPostgresDB(cfg)
	if errPostgres != nil {
		appLogger.Fatal("Error connecting to Postgres: ", errPostgres)
	} else {
		appLogger.Info("Postgres connection established")
	}
	defer postgresDb.Close()

	elasticClient, errElastic := elastic.InitElasticClient(ctx, cfg, appLogger)
	if errElastic != nil {
		appLogger.Fatal("Error connecting to Elastic: ", errElastic)
	} else {
		appLogger.Info("Elastic connection established")
	}

	otelEndpoint := fmt.Sprintf("%s:%s", cfg.OtelCollector.Host, cfg.OtelCollector.Port)
	m := trace.New(otelEndpoint, "order-service")

	postgresRepo := repository.NewPostgresRepository(appLogger, postgresDb)
	elasticRepository := repository.NewElasticRepository(appLogger, cfg, elasticClient)

	esDb, errEs := eventstroredb.NewEventStoreDB(cfg.EventStoreConfig)
	if errEs != nil {
		appLogger.Fatal(errEs)
		return
	}
	defer esDb.Close() // nolint: errcheck
	aggregateStore := store.NewAggregateStore(appLogger, esDb)

	orderProjection := postgres_projection.NewOrderProjection(appLogger, esDb, postgresRepo, cfg)
	elasticProjection := elastic_projection.NewElasticProjection(appLogger, esDb, elasticRepository, cfg)

	go func() {
		err := orderProjection.Subscribe(ctx, m.TracerProvider, []string{cfg.Subscriptions.OrderPrefix}, cfg.Subscriptions.PoolSize, orderProjection.ProcessEvents)
		if err != nil {
			appLogger.Errorf("(orderProjection.Subscribe) err: {%v}", err)
			cancel()
		}
	}()
	go func() {
		err := elasticProjection.Subscribe(ctx, m.TracerProvider, []string{cfg.Subscriptions.OrderPrefix}, cfg.Subscriptions.PoolSize, elasticProjection.ProcessEvents)
		if err != nil {
			appLogger.Errorf("(elasticProjection.Subscribe) err: {%v}", err)
			cancel()
		}
	}()

	orderService := service.NewOrderService(cfg, appLogger, aggregateStore, postgresRepo, elasticRepository)

	orderHandlers := v1.NewOrderHandlers(cfg, appLogger, orderService, metricsCounters)

	httpServer := http.NewHttpServer(cfg, appLogger)
	grpcServer := grpc.NewGrpcServer(cfg.GRPC.Port)
	healthCheckServer := health_check.NewHealthCheckServer(ctx, cfg, appLogger, postgresDb, elasticClient)
	servers := server.NewServer(ctx, cfg, appLogger, m, httpServer, grpcServer, healthCheckServer, orderService, metricsCounters, server.Handlers{
		OrderHttpHandlers: orderHandlers,
	})
	servers.ListenAll()

}
