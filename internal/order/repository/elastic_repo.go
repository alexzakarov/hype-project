package repository

import (
	"context"
	"encoding/json"
	v7 "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	otracer "go.opentelemetry.io/otel/trace"
	"main/config"
	"main/internal/dto"
	"main/internal/mappers"
	_interface "main/internal/order/domain/interface"
	"main/internal/order/domain/models"
	"main/pkg/logger"
	utils "main/pkg/utils/pagination"
)

const (
	shopItemTitle            = "shopItems.title"
	shopItemDescription      = "shopItems.description"
	minimumNumberShouldMatch = 1
)

type ElasticRepository struct {
	log           logger.Logger
	cfg           *config.Config
	elasticClient *v7.Client
}

func NewElasticRepository(log logger.Logger, cfg *config.Config, elasticClient *v7.Client) _interface.IElasticOrderRepository {
	return &ElasticRepository{log: log, cfg: cfg, elasticClient: elasticClient}
}

func (e *ElasticRepository) IndexOrder(ctx context.Context, t *trace.TracerProvider, order *models.OrderProjection) error {
	var span otracer.Span
	tracer := t.Tracer("ElasticRepository IndexOrder")

	ctx, span = tracer.Start(ctx, "elasticRepository.IndexOrder", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("OrderID", order.OrderID))

	res, err := e.elasticClient.Index().Index(e.cfg.ElasticIndexes.Orders).BodyJson(order).Id(order.OrderID).Do(ctx)
	if err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "elasticClient.Index")
	}

	e.log.Debugf("(IndexOrder) result: {%s}", res.Result)
	return nil
}

func (e *ElasticRepository) GetByID(ctx context.Context, t *trace.TracerProvider, orderID string) (*models.OrderProjection, error) {
	var span otracer.Span
	tracer := t.Tracer("ElasticRepository GetByID")

	ctx, span = tracer.Start(ctx, "elasticRepository.GetByID", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("OrderID", orderID))

	result, err := e.elasticClient.Get().Index(e.cfg.ElasticIndexes.Orders).Id(orderID).FetchSource(true).Do(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, errors.Wrap(err, "elasticClient.Get")
	}

	jsonData, err := result.Source.MarshalJSON()
	if err != nil {
		span.RecordError(err)
		return nil, errors.Wrap(err, "Source.MarshalJSON")
	}

	var order models.OrderProjection
	if err := json.Unmarshal(jsonData, &order); err != nil {
		span.RecordError(err)
		return nil, errors.Wrap(err, "json.Unmarshal")
	}

	return &order, nil
}

func (e *ElasticRepository) UpdateOrder(ctx context.Context, t *trace.TracerProvider, order *models.OrderProjection) error {
	var span otracer.Span
	tracer := t.Tracer("ElasticRepository UpdateOrder")

	ctx, span = tracer.Start(ctx, "elasticRepository.UpdateOrder", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("OrderID", order.OrderID))

	res, err := e.elasticClient.Update().Index(e.cfg.ElasticIndexes.Orders).Id(order.OrderID).Doc(order).FetchSource(true).Do(ctx)
	if err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "elasticClient.Update")
	}

	e.log.Debugf("(UpdateShoppingCart) result: {%s}", res.Result)
	return nil
}

func (e *ElasticRepository) Search(ctx context.Context, t *trace.TracerProvider, text string, pq *utils.Pagination) (*dto.OrderSearchResponseDto, error) {
	var span otracer.Span
	tracer := t.Tracer("ElasticRepository Search")

	ctx, span = tracer.Start(ctx, "elasticRepository.Search", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("Search", text))

	shouldMatch := v7.NewBoolQuery().
		Should(v7.NewMatchPhrasePrefixQuery(shopItemTitle, text), v7.NewMatchPhrasePrefixQuery(shopItemDescription, text)).
		MinimumNumberShouldMatch(minimumNumberShouldMatch)

	searchResult, err := e.elasticClient.Search(e.cfg.ElasticIndexes.Orders).
		Query(shouldMatch).
		From(pq.GetOffset()).
		Explain(e.cfg.Elastic.Explain).
		FetchSource(e.cfg.Elastic.FetchSource).
		Version(e.cfg.Elastic.Version).
		Size(pq.GetSize()).
		Pretty(e.cfg.Elastic.Pretty).
		Do(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, errors.Wrap(err, "elasticClient.Search")
	}

	orders := make([]*models.OrderProjection, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		jsonBytes, err := hit.Source.MarshalJSON()
		if err != nil {
			span.RecordError(err)
			return nil, errors.Wrap(err, "Source.MarshalJSON")
		}
		var order models.OrderProjection
		if err := json.Unmarshal(jsonBytes, &order); err != nil {
			span.RecordError(err)
			return nil, errors.Wrap(err, "json.Unmarshal")
		}
		orders = append(orders, &order)
	}

	return &dto.OrderSearchResponseDto{
		Pagination: dto.Pagination{
			TotalCount: searchResult.TotalHits(),
			TotalPages: int64(pq.GetTotalPages(int(searchResult.TotalHits()))),
			Page:       int64(pq.GetPage()),
			Size:       int64(pq.GetSize()),
			HasMore:    pq.GetHasMore(int(searchResult.TotalHits())),
		},
		Orders: mappers.OrdersFromProjections(orders),
	}, nil
}
