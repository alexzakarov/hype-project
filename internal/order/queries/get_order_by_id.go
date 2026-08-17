package queries

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/sdk/trace"
	"main/config"
	"main/internal/mappers"
	"main/internal/order/aggregate"
	_interface "main/internal/order/domain/interface"
	"main/internal/order/domain/models"
	"main/pkg/es"
	"main/pkg/logger"
)

type GetOrderByIDQueryHandler interface {
	Handle(ctx context.Context, t *trace.TracerProvider, command *GetOrderByIDQuery) (*models.OrderProjection, error)
}

type GetOrderByIDHandler struct {
	log          logger.Logger
	cfg          *config.Config
	es           es.AggregateStore
	postgresRepo _interface.IPostgresRepository
}

func NewGetOrderByIDHandler(log logger.Logger, cfg *config.Config, es es.AggregateStore, postgresRepo _interface.IPostgresRepository) *GetOrderByIDHandler {
	return &GetOrderByIDHandler{log: log, cfg: cfg, es: es, postgresRepo: postgresRepo}
}

func (q *GetOrderByIDHandler) Handle(ctx context.Context, t *trace.TracerProvider, query *GetOrderByIDQuery) (*models.OrderProjection, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "getOrderByIDHandler.Handle")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", query.ID))

	orderProjection, err := q.postgresRepo.GetById(ctx, t, query.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if orderProjection != nil {
		return orderProjection, nil
	}

	order := aggregate.NewOrderAggregateWithID(query.ID)
	if err := q.es.Load(ctx, t, order); err != nil {
		return nil, err
	}

	if aggregate.IsAggregateNotFound(order) {
		return nil, aggregate.ErrOrderNotFound
	}

	orderProjection = mappers.OrderProjectionFromAggregate(order)

	_, err = q.postgresRepo.Create(ctx, t, orderProjection)
	if err != nil {
		return nil, err
	}

	return orderProjection, nil
}
