package v1

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	otracer "go.opentelemetry.io/otel/trace"
	"main/config"
	"main/internal/order/aggregate"
	"main/pkg/es"
	"main/pkg/logger"
)

type IPayOrderCommandHandler interface {
	Handle(ctx context.Context, t *trace.TracerProvider, command *PayOrderCommand) error
}

type payOrderCommandHandler struct {
	log logger.Logger
	cfg *config.Config
	es  es.AggregateStore
}

func NewOrderPaidHandler(log logger.Logger, cfg *config.Config, es es.AggregateStore) IPayOrderCommandHandler {
	return &payOrderCommandHandler{log: log, cfg: cfg, es: es}
}

func (c *payOrderCommandHandler) Handle(ctx context.Context, t *trace.TracerProvider, command *PayOrderCommand) error {
	var span otracer.Span
	tracer := t.Tracer("payOrderCommandHandler Handle")

	ctx, span = tracer.Start(ctx, "payOrderCommandHandler.Handle", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", command.GetAggregateID()))

	order, err := aggregate.LoadOrderAggregate(ctx, t, c.es, command.GetAggregateID())
	if err != nil {
		return err
	}

	if err := order.PayOrder(ctx, t, command.Payment); err != nil {
		return err
	}

	return c.es.Save(ctx, t, order)
}
