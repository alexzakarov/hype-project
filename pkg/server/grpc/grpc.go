package grpc

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"log"
	"net"
	"time"
)

/*
type server struct {
	hello.UnimplementedGreeterServer
}

var tracer = otel.Tracer("order-service")

func (s *server) SayHello(ctx context.Context, in *hello.HelloRequest) (*hello.HelloReply, error) {
	var span trace.Span
	md, ok := metadata.FromIncomingContext(ctx)

	if ok {
		ctx = otel.GetTextMapPropagator().Extract(
			ctx,
			propagation.HeaderCarrier(md),
		)
	}

	ctx, span = tracer.Start(ctx, "process-order",
		trace.WithAttributes(
			attribute.String("hello.name", in.Name),
			attribute.String("hello.string", in.String()),
		),
	)
	defer span.End()

	conn, client := NewGrpcClient("server3:50053")

	defer func() { _ = conn.Close() }()

	err := callSayHello(ctx, client)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return &hello.HelloReply{Message: "Hello " + in.Name}, nil
}
*/

const (
	maxConnectionIdle = 5
	gRPCTimeout       = 40
	maxConnectionAge  = 5
	gRPCTime          = 10
)

type ServerGrpc struct {
	listener net.Listener
	Server   *grpc.Server
}

func NewGrpcServer(port string) *ServerGrpc {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.MaxRecvMsgSize(1*10000000000),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: maxConnectionIdle * time.Minute,
			Timeout:           gRPCTimeout * time.Second,
			MaxConnectionAge:  maxConnectionAge * time.Minute,
			Time:              gRPCTime * time.Minute,
		}),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	//hello.RegisterGreeterServer(s, &server{})

	return &ServerGrpc{listener: lis, Server: s}
}

func (s *ServerGrpc) Serve() error {
	log.Printf("server listening at %v", s.listener.Addr())
	if err := s.Server.Serve(s.listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
		return err
	}
	return nil
}

func NewGrpcClient(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		panic(err)
	}
	//client := hello.NewGreeterClient(conn)
	return conn
}

/*
	func callSayHello(ctx context.Context, c hello.GreeterClient) error {
		ctx = injectTraceContext(ctx)

		response, err := c.SayHello(ctx, &hello.HelloRequest{Name: "World"})
		if err != nil {
			return fmt.Errorf("calling SayHello: %w", err)
		}
		log.Printf("Response from server: %s", response.Message)
		return nil
	}
*/
