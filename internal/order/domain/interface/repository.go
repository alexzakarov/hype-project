package _interface

import (
	"context"
	"go.opentelemetry.io/otel/sdk/trace"
	"main/internal/dto"
	"main/internal/order/domain/models"
	utils "main/pkg/utils/pagination"
)

type IPostgresRepository interface {
	GetById(context.Context, *trace.TracerProvider, string) (*models.OrderProjection, error)
	GetAll(context.Context, *trace.TracerProvider) ([]*models.OrderProjection, error)
	Create(context.Context, *trace.TracerProvider, *models.OrderProjection) (string, error)
	UpdateOrder(context.Context, *trace.TracerProvider, *models.OrderProjection) error
	UpdateCancel(context.Context, *trace.TracerProvider, *models.OrderProjection) error
	UpdatePayment(context.Context, *trace.TracerProvider, *models.OrderProjection) error
	Complete(context.Context, *trace.TracerProvider, *models.OrderProjection) error
	UpdateDeliveryAddress(context.Context, *trace.TracerProvider, *models.OrderProjection) error
	UpdateSubmit(context.Context, *trace.TracerProvider, *models.OrderProjection) error
	Delete(context.Context, *trace.TracerProvider, *models.OrderProjection) error
}

type IElasticOrderRepository interface {
	IndexOrder(context.Context, *trace.TracerProvider, *models.OrderProjection) error
	GetByID(context.Context, *trace.TracerProvider, string) (*models.OrderProjection, error)
	UpdateOrder(context.Context, *trace.TracerProvider, *models.OrderProjection) error
	Search(context.Context, *trace.TracerProvider, string, *utils.Pagination) (*dto.OrderSearchResponseDto, error)
}
