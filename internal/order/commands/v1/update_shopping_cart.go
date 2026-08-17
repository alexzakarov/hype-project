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

type IUpdateShoppingCartCommandHandler interface {
	Handle(ctx context.Context, t *trace.TracerProvider, command *UpdateShoppingCartCommand) error
}

type updateShoppingCartCmdHandler struct {
	log logger.Logger
	cfg *config.Config
	es  es.AggregateStore
}

func NewUpdateShoppingCartCmdHandler(log logger.Logger, cfg *config.Config, es es.AggregateStore) IUpdateShoppingCartCommandHandler {
	return &updateShoppingCartCmdHandler{log: log, cfg: cfg, es: es}
}

func (c *updateShoppingCartCmdHandler) Handle(ctx context.Context, t *trace.TracerProvider, command *UpdateShoppingCartCommand) error {
	var span otracer.Span
	tracer := t.Tracer("updateShoppingCartCmdHandler Handle")

	ctx, span = tracer.Start(ctx, "updateShoppingCartCmdHandler.Handle", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", command.GetAggregateID()))

	order, err := aggregate.LoadOrderAggregate(ctx, t, c.es, command.GetAggregateID())
	if err != nil {
		return err
	}

	if err := order.UpdateShoppingCart(ctx, t, command.ShopItems); err != nil {
		return err
	}

	return c.es.Save(ctx, t, order)
}
