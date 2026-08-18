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
	"strings"
)

type ICreateOrderCommandHandler interface {
	Handle(ctx context.Context, t *trace.TracerProvider, command *CreateOrderCommand) error
}

type createOrderHandler struct {
	log logger.Logger
	cfg *config.Config
	es  es.AggregateStore
}

func NewCreateOrderHandler(log logger.Logger, cfg *config.Config, es es.AggregateStore) ICreateOrderCommandHandler {
	return &createOrderHandler{log: log, cfg: cfg, es: es}
}

func (c *createOrderHandler) Handle(ctx context.Context, t *trace.TracerProvider, command *CreateOrderCommand) error {
	var span otracer.Span
	tracer := t.Tracer("createOrderHandler Handle")
	ctx, span = tracer.Start(ctx, "createOrderHandler.Handle", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", command.GetAggregateID()))

	order := aggregate.NewOrderAggregateWithID(command.AggregateID)
	err := c.es.Exists(ctx, t, order.GetID())
	if err != nil && !strings.Contains(err.Error(), "is not found") {
		return err
	}

	if err := order.CreateOrder(ctx, t, command.ShopItems, command.AccountEmail, command.DeliveryAddress); err != nil {
		return err
	}

	span.SetAttributes(attribute.String("order", order.String()))
	return c.es.Save(ctx, t, order)
}
