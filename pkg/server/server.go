package server

import (
	"context"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc/reflection"
	"main/config"
	"main/internal/metrics"
	grpcOrder "main/internal/order/delivery/grpc"
	v1 "main/internal/order/delivery/http/v1"
	"main/internal/order/service"
	"main/pkg/graceful_exit"
	"main/pkg/logger"
	"main/pkg/middlewares/trace"
	"main/pkg/server/grpc"
	"main/pkg/server/health_check"
	server "main/pkg/server/http"
	orderService "main/proto"
)

const (
	certFile        = "ssl/server.crt"
	keyFile         = "ssl/server.pem"
	maxHeaderBytes  = 1 << 20
	gzipLevel       = 5
	stackSize       = 1 << 10 // 1 KB
	csrfTokenHeader = "X-CSRF-Token"
	bodyLimit       = "2M"
)

type Handlers struct {
	OrderHttpHandlers v1.IOrderHandlers
}

// Server
type Server struct {
	ctx               context.Context
	cfg               *config.Config
	logger            logger.Logger
	tracerMiddleware  *trace.Middleware
	serverHttp        *server.ServerHttp
	serverGrpc        *grpc.ServerGrpc
	serverHealthCheck *health_check.ServerHealthCheck
	orderService      *service.OrderService
	metrics           *metrics.ESMicroserviceMetrics
	Handlers          Handlers
}

// NewServer constructor
func NewServer(
	ctx context.Context,
	cfg *config.Config,
	logger logger.Logger,
	tracerMiddleware *trace.Middleware,
	serverHttp *server.ServerHttp,
	serverGrpc *grpc.ServerGrpc,
	serverHealthCheck *health_check.ServerHealthCheck,
	orderService *service.OrderService,
	metrics *metrics.ESMicroserviceMetrics,
	handlers Handlers,
) *Server {

	return &Server{
		cfg:               cfg,
		ctx:               ctx,
		logger:            logger,
		tracerMiddleware:  tracerMiddleware,
		serverHttp:        serverHttp,
		serverGrpc:        serverGrpc,
		serverHealthCheck: serverHealthCheck,
		orderService:      orderService,
		metrics:           metrics,
		Handlers:          handlers,
	}
}

func (s *Server) ListenAll() {

	versioning := s.serverHttp.Server.Group(s.cfg.ApiVersion)

	s.Handlers.OrderHttpHandlers.MapRoutes(s.tracerMiddleware, versioning)

	go s.serverHttp.Listen(s.serverHttp.Server)

	orderGrpcService := grpcOrder.NewOrderGrpcService(s.cfg, s.logger, s.tracerMiddleware.TracerProvider, s.metrics, s.orderService)
	orderService.RegisterOrderServiceServer(s.serverGrpc.Server, orderGrpcService)
	grpc_prometheus.Register(s.serverGrpc.Server)

	if s.cfg.GRPC.Development {
		reflection.Register(s.serverGrpc.Server)
	}
	go s.serverGrpc.Serve()

	go s.serverHealthCheck.RunHealthCheck(s.ctx)
	go s.serverHealthCheck.RunMetricsServer()

	graceful_exit.TerminateApp(s.ctx)

}
