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

type ISubmitOrderCommandHandler interface {
	Handle(ctx context.Context, t *trace.TracerProvider, command *SubmitOrderCommand) error
}

type submitOrderHandler struct {
	log logger.Logger
	cfg *config.Config
	es  es.AggregateStore
}

func NewSubmitOrderHandler(log logger.Logger, cfg *config.Config, es es.AggregateStore) ISubmitOrderCommandHandler {
	return &submitOrderHandler{log: log, cfg: cfg, es: es}
}

func (c *submitOrderHandler) Handle(ctx context.Context, t *trace.TracerProvider, command *SubmitOrderCommand) error {
	var span otracer.Span
	tracer := t.Tracer("submitOrderHandler Handle")

	ctx, span = tracer.Start(ctx, "submitOrderHandler.Handle", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", command.GetAggregateID()))

	order, err := aggregate.LoadOrderAggregate(ctx, t, c.es, command.GetAggregateID())
	if err != nil {
		return err
	}

	if err := order.SubmitOrder(ctx, t); err != nil {
		return err
	}

	return c.es.Save(ctx, t, order)
}
