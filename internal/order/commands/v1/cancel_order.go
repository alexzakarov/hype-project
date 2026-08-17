package v1

import (
	"context"
	"github.com/spiffe/go-spiffe/v2/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	otracer "go.opentelemetry.io/otel/trace"
	"main/config"
	"main/internal/order/aggregate"
	"main/pkg/es"
)

type ICancelOrderCommandHandler interface {
	Handle(ctx context.Context, t *trace.TracerProvider, command *CancelOrderCommand) error
}

type cancelOrderCommandHandler struct {
	log logger.Logger
	cfg *config.Config
	es  es.AggregateStore
}

func NewCancelOrderCommandHandler(log logger.Logger, cfg *config.Config, es es.AggregateStore) ICancelOrderCommandHandler {
	return &cancelOrderCommandHandler{log: log, cfg: cfg, es: es}
}

func (c *cancelOrderCommandHandler) Handle(ctx context.Context, t *trace.TracerProvider, command *CancelOrderCommand) error {
	var span otracer.Span
	tracer := t.Tracer("cancelOrderCommandHandler Handle")

	ctx, span = tracer.Start(ctx, "cancelOrderCommandHandler.Handle", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", command.GetAggregateID()))

	order, err := aggregate.LoadOrderAggregate(ctx, t, c.es, command.GetAggregateID())
	if err != nil {
		return err
	}

	if err := order.CancelOrder(ctx, t, command.CancelReason); err != nil {
		return err
	}

	return c.es.Save(ctx, t, order)
}
