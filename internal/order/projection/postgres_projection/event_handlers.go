package postgres_projection

import (
	"context"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	otracer "go.opentelemetry.io/otel/trace"
	"main/internal/order/aggregate"
	"main/internal/order/domain/models"
	v1 "main/internal/order/events/v1"
	"main/pkg/es"
)

func (o *PostgresProjection) onOrderCreate(ctx context.Context, t *trace.TracerProvider, evt es.Event) error {
	var span otracer.Span
	tracer := t.Tracer("PostgresProjection onOrderCreate")

	ctx, span = tracer.Start(ctx, "postgresProjection.onOrderCreate", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", evt.GetAggregateID()), attribute.String("EventType", evt.GetEventType()))

	var eventData v1.OrderCreatedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "evt.GetJsonData")
	}
	span.SetAttributes(attribute.String("AccountEmail", eventData.AccountEmail))

	op := &models.OrderProjection{
		OrderID:         aggregate.GetOrderAggregateID(evt.AggregateID),
		ShopItems:       eventData.ShopItems,
		AccountEmail:    eventData.AccountEmail,
		TotalPrice:      aggregate.GetShopItemsTotalPrice(eventData.ShopItems),
		DeliveryAddress: eventData.DeliveryAddress,
	}

	_, err := o.postgresRepo.Create(ctx, t, op)
	if err != nil {
		return err
	}

	return nil
}

func (o *PostgresProjection) onOrderPaid(ctx context.Context, t *trace.TracerProvider, evt es.Event) error {
	var span otracer.Span
	tracer := t.Tracer("PostgresProjection onOrderPaid")

	ctx, span = tracer.Start(ctx, "postgresProjection.onOrderPaid", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", evt.GetAggregateID()))

	var payment models.Payment
	if err := evt.GetJsonData(&payment); err != nil {
		return errors.Wrap(err, "GetJsonData")
	}

	op := &models.OrderProjection{OrderID: aggregate.GetOrderAggregateID(evt.AggregateID), Paid: true, Payment: payment}
	return o.postgresRepo.UpdatePayment(ctx, t, op)
}

func (o *PostgresProjection) onSubmit(ctx context.Context, t *trace.TracerProvider, evt es.Event) error {
	var span otracer.Span
	tracer := t.Tracer("PostgresProjection onSubmit")

	ctx, span = tracer.Start(ctx, "postgresProjection.onSubmit", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", evt.GetAggregateID()))

	op := &models.OrderProjection{OrderID: aggregate.GetOrderAggregateID(evt.AggregateID), Submitted: true}
	return o.postgresRepo.UpdateSubmit(ctx, t, op)
}

func (o *PostgresProjection) onShoppingCartUpdate(ctx context.Context, t *trace.TracerProvider, evt es.Event) error {
	var span otracer.Span
	tracer := t.Tracer("PostgresProjection onShoppingCartUpdate")

	ctx, span = tracer.Start(ctx, "postgresProjection.onShoppingCartUpdate", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", evt.GetAggregateID()))

	var eventData v1.ShoppingCartUpdatedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "evt.GetJsonData")
	}

	op := &models.OrderProjection{OrderID: aggregate.GetOrderAggregateID(evt.AggregateID), ShopItems: eventData.ShopItems}
	op.TotalPrice = aggregate.GetShopItemsTotalPrice(eventData.ShopItems)
	return o.postgresRepo.UpdateOrder(ctx, t, op)
}

func (o *PostgresProjection) onCancel(ctx context.Context, t *trace.TracerProvider, evt es.Event) error {
	var span otracer.Span
	tracer := t.Tracer("PostgresProjection onCancel")

	ctx, span = tracer.Start(ctx, "postgresProjection.onCancel", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", evt.GetAggregateID()))

	var eventData v1.OrderCanceledEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "evt.GetJsonData")
	}

	op := &models.OrderProjection{
		OrderID:      aggregate.GetOrderAggregateID(evt.AggregateID),
		Canceled:     true,
		Completed:    false,
		CancelReason: eventData.CancelReason,
	}
	return o.postgresRepo.UpdateCancel(ctx, t, op)
}

func (o *PostgresProjection) onCompleted(ctx context.Context, t *trace.TracerProvider, evt es.Event) error {
	var span otracer.Span
	tracer := t.Tracer("PostgresProjection onCompleted")

	ctx, span = tracer.Start(ctx, "postgresProjection.onCompleted", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", evt.GetAggregateID()))

	var eventData v1.OrderCompletedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "evt.GetJsonData")
	}

	op := &models.OrderProjection{
		OrderID:       aggregate.GetOrderAggregateID(evt.AggregateID),
		Canceled:      false,
		Completed:     true,
		DeliveredTime: eventData.DeliveryTimestamp,
	}
	return o.postgresRepo.Complete(ctx, t, op)
}

func (o *PostgresProjection) onDeliveryAddressChnaged(ctx context.Context, t *trace.TracerProvider, evt es.Event) error {
	var span otracer.Span
	tracer := t.Tracer("PostgresProjection onCompleted")

	ctx, span = tracer.Start(ctx, "postgresProjection.onCompleted", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", evt.GetAggregateID()))

	var eventData v1.OrderDeliveryAddressChangedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "evt.GetJsonData")
	}

	op := &models.OrderProjection{
		OrderID:         aggregate.GetOrderAggregateID(evt.AggregateID),
		DeliveryAddress: eventData.DeliveryAddress,
	}
	return o.postgresRepo.UpdateDeliveryAddress(ctx, t, op)
}
