package aggregate

import (
	"context"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	otracer "go.opentelemetry.io/otel/trace"
	"main/internal/order/domain/models"
	"main/pkg/es"
	"strings"
)

func GetShopItemsTotalPrice(shopItems []*models.ShopItem) float64 {
	var totalPrice float64 = 0
	for _, item := range shopItems {
		totalPrice += item.Price * float64(item.Quantity)
	}
	return totalPrice
}

// GetOrderAggregateID get order aggregate id for eventstoredb
func GetOrderAggregateID(eventAggregateID string) string {
	return strings.ReplaceAll(eventAggregateID, "order-", "")
}

func IsAggregateNotFound(aggregate es.Aggregate) bool {
	return aggregate.GetVersion() == 0
}

func LoadOrderAggregate(ctx context.Context, t *trace.TracerProvider, eventStore es.AggregateStore, aggregateID string) (*OrderAggregate, error) {
	var span otracer.Span
	tracer := t.Tracer("Utils LoadOrderAggregate")

	ctx, span = tracer.Start(ctx, "utils.LoadOrderAggregate", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", aggregateID))

	order := NewOrderAggregateWithID(aggregateID)

	err := eventStore.Exists(ctx, t, order.GetID())
	errApi, _ := kurrentdb.FromError(err)
	if err != nil && errApi != nil && errApi.Code() != kurrentdb.ErrorCodeStreamDeleted {
		return nil, err
	}

	if err := eventStore.Load(ctx, t, order); err != nil {
		return nil, err
	}

	return order, nil
}
