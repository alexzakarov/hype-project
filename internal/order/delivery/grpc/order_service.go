package grpc

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	otracer "go.opentelemetry.io/otel/trace"

	"main/config"
	"main/internal/mappers"
	"main/internal/metrics"
	v1 "main/internal/order/commands/v1"
	"main/internal/order/domain/models"
	"main/internal/order/queries"
	"main/internal/order/service"
	grpcErrors "main/pkg/grpc_errors"
	"main/pkg/logger"
	"main/pkg/server/grpc"
	utils "main/pkg/utils/pagination"
	orderService "main/proto"
	"time"

	uuid "github.com/satori/go.uuid"
)

type OrderGrpcService struct {
	cfg     *config.Config
	log     logger.Logger
	tracer  *trace.TracerProvider
	metrics *metrics.ESMicroserviceMetrics
	os      *service.OrderService
}

func NewOrderGrpcService(cfg *config.Config, log logger.Logger, tracer *trace.TracerProvider, metrics *metrics.ESMicroserviceMetrics, os *service.OrderService) *OrderGrpcService {

	return &OrderGrpcService{log: log, tracer: tracer, metrics: metrics, os: os}
}

func (s *OrderGrpcService) CreateOrder(ctx context.Context, req *orderService.CreateOrderReq) (*orderService.CreateOrderRes, error) {
	var span otracer.Span
	ctx = grpc.ExtractTraceContext(ctx)
	tracer := s.tracer.Tracer("OrderGrpcService CreateOrder")

	ctx, span = tracer.Start(ctx, "orderGrpcService.CreateOrder", otracer.WithSpanKind(otracer.SpanKindServer))
	span.SetAttributes(attribute.String("req", req.String()))
	s.metrics.CreateOrderGrpcRequests.Inc()

	aggregateID := uuid.NewV4().String()
	command := v1.NewCreateOrderCommand(aggregateID, models.ShopItemsFromProto(req.GetShopItems()), req.GetAccountEmail(), req.GetDeliveryAddress())
	/*
		if err := s.v.StructCtx(ctx, command); err != nil {
			s.log.Errorf("(validate) aggregateID: {%s}, err: {%v}", aggregateID, err)
			span.RecordError(err)
			return nil, s.errResponse(err)
		}
	*/

	if err := s.os.Commands.CreateOrder.Handle(ctx, s.tracer, command); err != nil {
		span.RecordError(err)
		s.log.Errorf("(CreateOrder.Handle) orderID: {%s}, err: {%v}", aggregateID, err)
		return nil, s.errResponse(err)
	}

	s.log.Infof("(created order): orderID: {%s}", aggregateID)
	return &orderService.CreateOrderRes{AggregateID: aggregateID}, nil
}

func (s *OrderGrpcService) PayOrder(ctx context.Context, req *orderService.PayOrderReq) (*orderService.PayOrderRes, error) {
	var span otracer.Span
	ctx = grpc.ExtractTraceContext(ctx)
	tracer := s.tracer.Tracer("OrderGrpcService PayOrder")

	ctx, span = tracer.Start(ctx, "orderGrpcService.PayOrder", otracer.WithSpanKind(otracer.SpanKindServer))
	span.SetAttributes(attribute.String("req", req.String()))
	s.metrics.PayOrderGrpcRequests.Inc()

	payment := models.Payment{PaymentID: req.GetPayment().GetID(), Timestamp: time.Now()}
	command := v1.NewPayOrderCommand(payment, req.GetAggregateID())
	/*
		if err := s.v.StructCtx(ctx, command); err != nil {
			s.log.Errorf("(validate) err: {%v}", err)
			span.RecordError(err)
			return nil, s.errResponse(err)
		}
	*/

	if err := s.os.Commands.OrderPaid.Handle(ctx, s.tracer, command); err != nil {
		span.RecordError(err)
		s.log.Errorf("(OrderPaid.Handle) orderID: {%s}, err: {%v}", req.GetAggregateID(), err)
		return nil, s.errResponse(err)
	}

	s.log.Infof("(paid order): orderID: {%s}", req.GetAggregateID())
	return &orderService.PayOrderRes{AggregateID: req.GetAggregateID()}, nil
}

func (s *OrderGrpcService) SubmitOrder(ctx context.Context, req *orderService.SubmitOrderReq) (*orderService.SubmitOrderRes, error) {
	var span otracer.Span
	ctx = grpc.ExtractTraceContext(ctx)
	tracer := s.tracer.Tracer("OrderGrpcService SubmitOrder")

	ctx, span = tracer.Start(ctx, "orderGrpcService.SubmitOrder", otracer.WithSpanKind(otracer.SpanKindServer))
	span.SetAttributes(attribute.String("req", req.String()))
	s.metrics.SubmitOrderGrpcRequests.Inc()

	command := v1.NewSubmitOrderCommand(req.GetAggregateID())
	/*
		if err := s.v.StructCtx(ctx, command); err != nil {
			s.log.Errorf("(validate) err: {%v}", err)
			span.RecordError(err)
			return nil, s.errResponse(err)
		}
	*/

	if err := s.os.Commands.SubmitOrder.Handle(ctx, s.tracer, command); err != nil {
		span.RecordError(err)
		s.log.Errorf("(SubmitOrder.Handle) orderID: {%s}, err: {%v}", req.GetAggregateID(), err)
		return nil, s.errResponse(err)
	}

	s.log.Infof("(submitted order): orderID: {%s}", req.GetAggregateID())
	return &orderService.SubmitOrderRes{AggregateID: req.GetAggregateID()}, nil
}

func (s *OrderGrpcService) GetOrderByID(ctx context.Context, req *orderService.GetOrderByIDReq) (*orderService.GetOrderByIDRes, error) {
	var span otracer.Span
	ctx = grpc.ExtractTraceContext(ctx)
	tracer := s.tracer.Tracer("OrderGrpcService GetOrderByID")

	ctx, span = tracer.Start(ctx, "orderGrpcService.GetOrderByID", otracer.WithSpanKind(otracer.SpanKindServer))
	span.SetAttributes(attribute.String("req", req.String()))
	s.metrics.GetOrderByIdGrpcRequests.Inc()

	query := queries.NewGetOrderByIDQuery(req.GetAggregateID())
	/*
		if err := s.v.StructCtx(ctx, query); err != nil {
			s.log.Errorf("(validate) err: {%v}", err)
			span.RecordError(err)
			return nil, s.errResponse(err)
		}
	*/

	orderProjection, err := s.os.Queries.GetOrderByID.Handle(ctx, s.tracer, query)
	if err != nil {
		span.RecordError(err)
		s.log.Errorf("(GetOrderByID.Handle) orderID: {%s}, err: {%v}", req.GetAggregateID(), err)
		return nil, s.errResponse(err)
	}

	s.log.Infof("(GetOrderByID) AggregateID: {%s}", req.GetAggregateID())
	s.log.Debugf("(GetOrderByID) orderProjection: {%s}", orderProjection.String())
	return &orderService.GetOrderByIDRes{Order: models.OrderProjectionToProto(orderProjection)}, nil
}

func (s *OrderGrpcService) UpdateShoppingCart(ctx context.Context, req *orderService.UpdateShoppingCartReq) (*orderService.UpdateShoppingCartRes, error) {
	var span otracer.Span
	ctx = grpc.ExtractTraceContext(ctx)
	tracer := s.tracer.Tracer("OrderGrpcService UpdateShoppingCart")

	ctx, span = tracer.Start(ctx, "orderGrpcService.UpdateShoppingCart", otracer.WithSpanKind(otracer.SpanKindServer))
	span.SetAttributes(attribute.String("req", req.String()))
	s.metrics.UpdateOrderGrpcRequests.Inc()

	command := v1.NewUpdateShoppingCartCommand(req.GetAggregateID(), models.ShopItemsFromProto(req.GetShopItems()))
	/*
		if err := s.v.StructCtx(ctx, command); err != nil {
			s.log.Errorf("(validate) err: {%v}", err)
			span.RecordError(err)
			return nil, s.errResponse(err)
		}
	*/

	if err := s.os.Commands.UpdateOrder.Handle(ctx, s.tracer, command); err != nil {
		span.RecordError(err)
		s.log.Errorf("(UpdateShoppingCart.Handle) orderID: {%s}, err: {%v}", req.GetAggregateID(), err)
		return nil, s.errResponse(err)
	}

	s.log.Infof("(UpdateShoppingCart): AggregateID: {%s}", req.GetAggregateID())
	return &orderService.UpdateShoppingCartRes{}, nil
}

func (s *OrderGrpcService) CancelOrder(ctx context.Context, req *orderService.CancelOrderReq) (*orderService.CancelOrderRes, error) {
	var span otracer.Span
	ctx = grpc.ExtractTraceContext(ctx)
	tracer := s.tracer.Tracer("OrderGrpcService CancelOrder")

	ctx, span = tracer.Start(ctx, "orderGrpcService.CancelOrder", otracer.WithSpanKind(otracer.SpanKindServer))
	span.SetAttributes(attribute.String("req", req.String()))
	s.metrics.CancelOrderGrpcRequests.Inc()

	command := v1.NewCancelOrderCommand(req.GetAggregateID(), req.GetCancelReason())
	/*
	   if err := s.v.StructCtx(ctx, command); err != nil {

	   		s.log.Errorf("(validate) err: {%v}", err)
	   		span.RecordError(err)
	   		return nil, s.errResponse(err)
	   	}
	*/
	if err := s.os.Commands.CancelOrder.Handle(ctx, s.tracer, command); err != nil {
		span.RecordError(err)
		s.log.Errorf("(CancelOrder.Handle) orderID: {%s}, err: {%v}", req.GetAggregateID(), err)
		return nil, s.errResponse(err)
	}

	s.log.Infof("(CancelOrder): AggregateID: {%s}", req.GetAggregateID())
	return &orderService.CancelOrderRes{}, nil
}

func (s *OrderGrpcService) CompleteOrder(ctx context.Context, req *orderService.CompleteOrderReq) (*orderService.CompleteOrderRes, error) {
	var span otracer.Span
	ctx = grpc.ExtractTraceContext(ctx)
	tracer := s.tracer.Tracer("OrderGrpcService CompleteOrder")

	ctx, span = tracer.Start(ctx, "orderGrpcService.CompleteOrder", otracer.WithSpanKind(otracer.SpanKindServer))
	span.SetAttributes(attribute.String("req", req.String()))
	s.metrics.CompleteOrderGrpcRequests.Inc()

	command := v1.NewCompleteOrderCommand(req.GetAggregateID(), time.Now())
	/*
	   if err := s.v.StructCtx(ctx, command); err != nil {
	   		s.log.Errorf("(validate) err: {%v}", err)
	   		span.RecordError(err)
	   		return nil, s.errResponse(err)
	   	}
	*/

	if err := s.os.Commands.CompleteOrder.Handle(ctx, s.tracer, command); err != nil {
		span.RecordError(err)
		s.log.Errorf("(CompleteOrder.Handle) orderID: {%s}, err: {%v}", req.GetAggregateID(), err)
		return nil, s.errResponse(err)
	}

	s.log.Infof("(CompleteOrder): AggregateID: {%s}", req.GetAggregateID())
	return &orderService.CompleteOrderRes{}, nil
}

func (s *OrderGrpcService) ChangeDeliveryAddress(ctx context.Context, req *orderService.ChangeDeliveryAddressReq) (*orderService.ChangeDeliveryAddressRes, error) {
	var span otracer.Span
	ctx = grpc.ExtractTraceContext(ctx)
	tracer := s.tracer.Tracer("OrderGrpcService ChangeDeliveryAddress")

	ctx, span = tracer.Start(ctx, "orderGrpcService.ChangeDeliveryAddress", otracer.WithSpanKind(otracer.SpanKindServer))
	span.SetAttributes(attribute.String("req", req.String()))
	s.metrics.ChangeAddressOrderGrpcRequests.Inc()

	command := v1.NewChangeDeliveryAddressCommand(req.GetAggregateID(), req.GetDeliveryAddress())
	/*
		if err := s.v.StructCtx(ctx, command); err != nil {
			s.log.Errorf("(validate) err: {%v}", err)
			span.RecordError(err)
			return nil, s.errResponse(err)
		}
	*/

	if err := s.os.Commands.ChangeOrderDeliveryAddress.Handle(ctx, s.tracer, command); err != nil {
		span.RecordError(err)
		s.log.Errorf("(ChangeOrderDeliveryAddress.Handle) orderID: {%s}, err: {%v}", req.GetAggregateID(), err)
		return nil, s.errResponse(err)
	}

	s.log.Infof("(ChangeDeliveryAddress): AggregateID: {%s}", req.GetAggregateID())
	return &orderService.ChangeDeliveryAddressRes{}, nil
}

func (s *OrderGrpcService) Search(ctx context.Context, req *orderService.SearchReq) (*orderService.SearchRes, error) {
	var span otracer.Span
	ctx = grpc.ExtractTraceContext(ctx)
	tracer := s.tracer.Tracer("OrderGrpcService Search")

	ctx, span = tracer.Start(ctx, "orderGrpcService.Search", otracer.WithSpanKind(otracer.SpanKindServer))
	span.SetAttributes(attribute.String("req", req.String()))
	s.metrics.ChangeAddressOrderGrpcRequests.Inc()

	query := queries.NewSearchOrdersQuery(req.GetSearchText(), utils.NewPaginationQuery(int(req.GetSize()), int(req.GetPage())))
	/*
		if err := s.v.StructCtx(ctx, query); err != nil {
			s.log.Errorf("(validate) err: {%v}", err)
			span.RecordError(err)
			return nil, s.errResponse(err)
		}
	*/

	searchResult, err := s.os.Queries.SearchOrders.Handle(ctx, s.tracer, query)
	if err != nil {
		s.log.Errorf("(SearchOrders.Handle) text: {%s}, err: {%v}", req.GetSearchText(), err)
		return nil, s.errResponse(err)
	}

	s.log.Infof("(Search result): searchText: {%s}, pagination: {%+v}", req.GetSearchText(), searchResult.Pagination)
	return mappers.SearchResponseToProto(searchResult), nil
}

func (s *OrderGrpcService) errResponse(err error) error {
	return grpcErrors.ErrResponse(err)
}
