package aggregate

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	otracer "go.opentelemetry.io/otel/trace"
	"main/internal/order/domain/models"
	eventsV1 "main/internal/order/events/v1"
	"main/pkg/server/grpc"
	"time"

	"github.com/pkg/errors"
)

func (a *OrderAggregate) CreateOrder(ctx context.Context, t *trace.TracerProvider, shopItems []*models.ShopItem, accountEmail, deliveryAddress string) error {
	var span otracer.Span
	tracer := t.Tracer("OrderAggregate CreateOrder")

	ctx, span = tracer.Start(ctx, "orderAggregate.CreateOrder", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", a.GetID()))

	if shopItems == nil {
		return ErrOrderShopItemsIsRequired
	}
	if deliveryAddress == "" {
		return ErrInvalidDeliveryAddress
	}

	event, err := eventsV1.NewOrderCreatedEvent(a, shopItems, accountEmail, deliveryAddress)
	if err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "NewOrderCreatedEvent")
	}

	if err := event.SetMetadata(grpc.InjectTraceContextAndConvertToMD(ctx)); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "SetMetadata")
	}

	return a.Apply(event)
}

func (a *OrderAggregate) PayOrder(ctx context.Context, t *trace.TracerProvider, payment models.Payment) error {
	var span otracer.Span
	tracer := t.Tracer("OrderAggregate PayOrder")

	ctx, span = tracer.Start(ctx, "orderAggregate.PayOrder", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", a.GetID()))

	if a.Order.Canceled {
		return ErrOrderAlreadyCancelled
	}
	if a.Order.Paid {
		return ErrAlreadyPaid
	}
	if a.Order.Submitted {
		return ErrAlreadySubmitted
	}

	event, err := eventsV1.NewOrderPaidEvent(a, &payment)
	if err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "NewOrderPaidEvent")
	}

	if err := event.SetMetadata(grpc.InjectTraceContextAndConvertToMD(ctx)); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "SetMetadata")
	}

	return a.Apply(event)
}

func (a *OrderAggregate) SubmitOrder(ctx context.Context, t *trace.TracerProvider) error {
	var span otracer.Span
	tracer := t.Tracer("OrderAggregate SubmitOrder")

	ctx, span = tracer.Start(ctx, "orderAggregate.SubmitOrder", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", a.GetID()))

	if a.Order.Canceled {
		return ErrOrderAlreadyCancelled
	}
	if !a.Order.Paid {
		return ErrOrderNotPaid
	}
	if a.Order.Submitted {
		return ErrAlreadySubmitted
	}

	event, err := eventsV1.NewSubmitOrderEvent(a)
	if err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "NewSubmitOrderEvent")
	}

	if err := event.SetMetadata(grpc.InjectTraceContextAndConvertToMD(ctx)); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "SetMetadata")
	}

	return a.Apply(event)
}

func (a *OrderAggregate) UpdateShoppingCart(ctx context.Context, t *trace.TracerProvider, shopItems []*models.ShopItem) error {
	var span otracer.Span
	tracer := t.Tracer("OrderAggregate UpdateShoppingCart")

	ctx, span = tracer.Start(ctx, "orderAggregate.UpdateShoppingCart", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", a.GetID()))

	if a.Order.Canceled {
		return ErrOrderAlreadyCancelled
	}
	if a.Order.Submitted {
		return ErrAlreadySubmitted
	}

	event, err := eventsV1.NewShoppingCartUpdatedEvent(a, shopItems)
	if err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "NewShoppingCartUpdatedEvent")
	}

	if err := event.SetMetadata(grpc.InjectTraceContextAndConvertToMD(ctx)); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "SetMetadata")
	}

	return a.Apply(event)
}

func (a *OrderAggregate) CancelOrder(ctx context.Context, t *trace.TracerProvider, cancelReason string) error {
	var span otracer.Span
	tracer := t.Tracer("OrderAggregate CancelOrder")

	ctx, span = tracer.Start(ctx, "orderAggregate.CancelOrder", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", a.GetID()))

	if a.Order.Completed {
		return ErrOrderAlreadyCompleted
	}
	if cancelReason == "" {
		return ErrCancelReasonRequired
	}

	event, err := eventsV1.NewOrderCanceledEvent(a, cancelReason)
	if err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "NewOrderCanceledEvent")
	}

	if err := event.SetMetadata(grpc.InjectTraceContextAndConvertToMD(ctx)); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "SetMetadata")
	}

	return a.Apply(event)
}

func (a *OrderAggregate) CompleteOrder(ctx context.Context, t *trace.TracerProvider, deliveryTimestamp time.Time) error {
	var span otracer.Span
	tracer := t.Tracer("OrderAggregate CompleteOrder")

	ctx, span = tracer.Start(ctx, "orderAggregate.CompleteOrder", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", a.GetID()))

	if a.Order.Completed {
		return ErrOrderAlreadyCompleted
	}
	if a.Order.Canceled {
		return ErrOrderAlreadyCanceled
	}
	if !a.Order.Paid {
		return ErrOrderMustBePaidBeforeDelivered
	}

	event, err := eventsV1.NewOrderCompletedEvent(a, deliveryTimestamp)
	if err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "NewOrderCompletedEvent")
	}

	if err := event.SetMetadata(grpc.InjectTraceContextAndConvertToMD(ctx)); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "SetMetadata")
	}

	return a.Apply(event)
}

func (a *OrderAggregate) ChangeDeliveryAddress(ctx context.Context, t *trace.TracerProvider, deliveryAddress string) error {
	var span otracer.Span
	tracer := t.Tracer("OrderAggregate ChangeDeliveryAddress")

	ctx, span = tracer.Start(ctx, "orderAggregate.ChangeDeliveryAddress", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", a.GetID()))

	if a.Order.Completed {
		return ErrOrderAlreadyCompleted
	}

	event, err := eventsV1.NewDeliveryAddressChangedEvent(a, deliveryAddress)
	if err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "NewDeliveryAddressChangedEvent")
	}

	if err := event.SetMetadata(grpc.InjectTraceContextAndConvertToMD(ctx)); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "SetMetadata")
	}

	return a.Apply(event)
}
