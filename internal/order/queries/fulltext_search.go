package queries

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"go.opentelemetry.io/otel/sdk/trace"
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
	span, ctx := opentracing.StartSpanFromContext(ctx, "searchOrdersHandler.Handle")
	defer span.Finish()
	span.LogFields(log.String("SearchText", command.SearchText))

	return s.elasticRepository.Search(ctx, t, command.SearchText, command.Pq)
}
