package queries

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	otracer "go.opentelemetry.io/otel/trace"
	"main/config"
	"main/internal/dto"
	_interface "main/internal/order/domain/interface"
	"main/pkg/es"
	"main/pkg/logger"
)

type SearchOrdersQueryHandler interface {
	Handle(ctx context.Context, t *trace.TracerProvider, command *SearchOrdersQuery) (*dto.OrderSearchResponseDto, error)
}

type searchOrdersHandler struct {
	log               logger.Logger
	cfg               *config.Config
	es                es.AggregateStore
	elasticRepository _interface.IElasticOrderRepository
}

func NewSearchOrdersHandler(log logger.Logger, cfg *config.Config, es es.AggregateStore, elasticRepository _interface.IElasticOrderRepository) *searchOrdersHandler {
	return &searchOrdersHandler{log: log, cfg: cfg, es: es, elasticRepository: elasticRepository}
}

func (s *searchOrdersHandler) Handle(ctx context.Context, t *trace.TracerProvider, command *SearchOrdersQuery) (*dto.OrderSearchResponseDto, error) {
	var span otracer.Span
	tracer := t.Tracer("SearchOrdersHandler Handle")

	ctx, span = tracer.Start(ctx, "searchOrdersHandler.Handle", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("SearchText", command.SearchText))

	return s.elasticRepository.Search(ctx, t, command.SearchText, command.Pq)
}
