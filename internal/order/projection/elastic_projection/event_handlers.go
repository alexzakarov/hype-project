package elastic_projection

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

func (o *ElasticProjection) onOrderCreate(ctx context.Context, t *trace.TracerProvider, evt es.Event) error {
	var span otracer.Span
	tracer := t.Tracer("ElasticProjection onOrderCreate")

	ctx, span = tracer.Start(ctx, "elasticProjection.onOrderCreate", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", evt.GetAggregateID()))

	var eventData v1.OrderCreatedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "evt.GetJsonData")
	}

	op := &models.OrderProjection{
		OrderID:      aggregate.GetOrderAggregateID(evt.AggregateID),
		ShopItems:    eventData.ShopItems,
		AccountEmail: eventData.AccountEmail,
		TotalPrice:   aggregate.GetShopItemsTotalPrice(eventData.ShopItems),
	}

	return o.elasticRepository.IndexOrder(ctx, t, op)
}

func (o *ElasticProjection) onOrderPaid(ctx context.Context, t *trace.TracerProvider, evt es.Event) error {
	var span otracer.Span
	tracer := t.Tracer("ElasticProjection onOrderPaid")

	ctx, span = tracer.Start(ctx, "elasticProjection.onOrderPaid", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", evt.GetAggregateID()))

	var payment models.Payment
	if err := evt.GetJsonData(&payment); err != nil {
		return errors.Wrap(err, "GetJsonData")
	}

	projection, err := o.elasticRepository.GetByID(ctx, t, aggregate.GetOrderAggregateID(evt.AggregateID))
	if err != nil {
		return err
	}
	projection.Paid = true
	projection.Payment = payment

	return o.elasticRepository.UpdateOrder(ctx, t, projection)
}

func (o *ElasticProjection) onSubmit(ctx context.Context, t *trace.TracerProvider, evt es.Event) error {
	var span otracer.Span
	tracer := t.Tracer("ElasticProjection onSubmit")

	ctx, span = tracer.Start(ctx, "elasticProjection.onSubmit", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", evt.GetAggregateID()))

	projection, err := o.elasticRepository.GetByID(ctx, t, aggregate.GetOrderAggregateID(evt.AggregateID))
	if err != nil {
		return err
	}
	projection.Submitted = true

	return o.elasticRepository.UpdateOrder(ctx, t, projection)
}

func (o *ElasticProjection) onShoppingCartUpdate(ctx context.Context, t *trace.TracerProvider, evt es.Event) error {
	var span otracer.Span
	tracer := t.Tracer("ElasticProjection onShoppingCartUpdate")

	ctx, span = tracer.Start(ctx, "elasticProjection.onShoppingCartUpdate", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", evt.GetAggregateID()))

	var eventData v1.ShoppingCartUpdatedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "evt.GetJsonData")
	}

	projection, err := o.elasticRepository.GetByID(ctx, t, aggregate.GetOrderAggregateID(evt.AggregateID))
	if err != nil {
		return err
	}
	projection.ShopItems = eventData.ShopItems
	projection.TotalPrice = aggregate.GetShopItemsTotalPrice(eventData.ShopItems)

	return o.elasticRepository.UpdateOrder(ctx, t, projection)
}

func (o *ElasticProjection) onCancel(ctx context.Context, t *trace.TracerProvider, evt es.Event) error {
	var span otracer.Span
	tracer := t.Tracer("ElasticProjection onCancel")

	ctx, span = tracer.Start(ctx, "elasticProjection.onCancel", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", evt.GetAggregateID()))

	var eventData v1.OrderCanceledEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "evt.GetJsonData")
	}

	projection, err := o.elasticRepository.GetByID(ctx, t, aggregate.GetOrderAggregateID(evt.AggregateID))
	if err != nil {
		return err
	}
	projection.Canceled = true
	projection.Completed = false
	projection.CancelReason = eventData.CancelReason

	return o.elasticRepository.UpdateOrder(ctx, t, projection)
}

func (o *ElasticProjection) onComplete(ctx context.Context, t *trace.TracerProvider, evt es.Event) error {
	var span otracer.Span
	tracer := t.Tracer("ElasticProjection onComplete")

	ctx, span = tracer.Start(ctx, "elasticProjection.onComplete", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", evt.GetAggregateID()))

	var eventData v1.OrderCompletedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "evt.GetJsonData")
	}

	projection, err := o.elasticRepository.GetByID(ctx, t, aggregate.GetOrderAggregateID(evt.AggregateID))
	if err != nil {
		return err
	}
	projection.Completed = true
	projection.DeliveredTime = eventData.DeliveryTimestamp

	return o.elasticRepository.UpdateOrder(ctx, t, projection)
}

func (o *ElasticProjection) onDeliveryAddressChanged(ctx context.Context, t *trace.TracerProvider, evt es.Event) error {
	var span otracer.Span
	tracer := t.Tracer("ElasticProjection onDeliveryAddressChanged")

	ctx, span = tracer.Start(ctx, "elasticProjection.onDeliveryAddressChanged", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", evt.GetAggregateID()))

	var eventData v1.OrderDeliveryAddressChangedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "evt.GetJsonData")
	}

	projection, err := o.elasticRepository.GetByID(ctx, t, aggregate.GetOrderAggregateID(evt.AggregateID))
	if err != nil {
		return err
	}
	projection.DeliveryAddress = eventData.DeliveryAddress

	return o.elasticRepository.UpdateOrder(ctx, t, projection)

}
