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

type IChangeDeliveryAddressCommandHandler interface {
	Handle(ctx context.Context, t *trace.TracerProvider, command *ChangeDeliveryAddressCommand) error
}

type changeDeliveryAddressCmdHandler struct {
	log logger.Logger
	cfg *config.Config
	es  es.AggregateStore
}

func NewChangeDeliveryAddressCmdHandler(log logger.Logger, cfg *config.Config, es es.AggregateStore) IChangeDeliveryAddressCommandHandler {
	return &changeDeliveryAddressCmdHandler{log: log, cfg: cfg, es: es}
}

func (c *changeDeliveryAddressCmdHandler) Handle(ctx context.Context, t *trace.TracerProvider, command *ChangeDeliveryAddressCommand) error {
	var span otracer.Span
	tracer := t.Tracer("changeDeliveryAddressCmdHandler Handle")

	ctx, span = tracer.Start(ctx, "changeDeliveryAddressCmdHandler.Handle", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", command.GetAggregateID()))

	order, err := aggregate.LoadOrderAggregate(ctx, t, c.es, command.GetAggregateID())
	if err != nil {
		return err
	}

	if err := order.ChangeDeliveryAddress(ctx, t, command.DeliveryAddress); err != nil {
		return err
	}

	return c.es.Save(ctx, t, order)
}
