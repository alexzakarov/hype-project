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

type ICompleteOrderCommandHandler interface {
	Handle(ctx context.Context, t *trace.TracerProvider, command *CompleteOrderCommand) error
}

type completeOrderCommandHandler struct {
	log logger.Logger
	cfg *config.Config
	es  es.AggregateStore
}

func NewCompleteOrderCommandHandler(log logger.Logger, cfg *config.Config, es es.AggregateStore) ICompleteOrderCommandHandler {
	return &completeOrderCommandHandler{log: log, cfg: cfg, es: es}
}

func (c *completeOrderCommandHandler) Handle(ctx context.Context, t *trace.TracerProvider, command *CompleteOrderCommand) error {
	var span otracer.Span
	tracer := t.Tracer("completeOrderCommandHandler Handle")

	ctx, span = tracer.Start(ctx, "completeOrderCommandHandler.Handle", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", command.GetAggregateID()))

	order, err := aggregate.LoadOrderAggregate(ctx, t, c.es, command.GetAggregateID())
	if err != nil {
		return err
	}

	if err := order.CompleteOrder(ctx, t, command.DeliveryTimestamp); err != nil {
		return err
	}

	return c.es.Save(ctx, t, order)
}
