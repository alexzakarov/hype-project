package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.opentelemetry.io/otel/sdk/trace"
	_interface "main/internal/order/domain/interface"
	"main/internal/order/domain/models"
	"main/pkg/logger"
)

type PostgresRepository struct {
	db     *pgxpool.Pool
	logger logger.Logger
}

func NewPostgresRepository(logger logger.Logger, db *pgxpool.Pool) _interface.IPostgresRepository {
	return &PostgresRepository{
		db:     db,
		logger: logger,
	}
}

func (p PostgresRepository) GetById(ctx context.Context, provider *trace.TracerProvider, s string) (*models.OrderProjection, error) {
	data := models.OrderProjection{}
	fmt.Println("Retrieving Order ID: ", s)

	query := `SELECT id, account_email FROM public.orders WHERE id=$1`
	errDb := p.db.QueryRow(ctx, query, s).Scan(&data.ID, &data.AccountEmail)
	fmt.Println(errDb, query)
	return &data, errDb
}

func (p PostgresRepository) GetAll(ctx context.Context, provider *trace.TracerProvider) ([]*models.OrderProjection, error) {
	return nil, nil

}

func (p PostgresRepository) Create(ctx context.Context, provider *trace.TracerProvider, projection *models.OrderProjection) (string, error) {
	var errDb error
	var query string
	query = `INSERT INTO public.orders (id, account_email) VALUES (trim($1),$2)`
	_, errDb = p.db.Exec(ctx, query, projection.OrderID, projection.AccountEmail)
	return projection.ID, errDb
}

func (p PostgresRepository) UpdateOrder(ctx context.Context, provider *trace.TracerProvider, projection *models.OrderProjection) error {
	//TODO implement me
	panic("implement me")
}

func (p PostgresRepository) UpdateCancel(ctx context.Context, provider *trace.TracerProvider, projection *models.OrderProjection) error {
	//TODO implement me
	panic("implement me")
}

func (p PostgresRepository) UpdatePayment(ctx context.Context, provider *trace.TracerProvider, projection *models.OrderProjection) error {
	//TODO implement me
	panic("implement me")
}

func (p PostgresRepository) Complete(ctx context.Context, provider *trace.TracerProvider, projection *models.OrderProjection) error {
	//TODO implement me
	panic("implement me")
}

func (p PostgresRepository) UpdateDeliveryAddress(ctx context.Context, provider *trace.TracerProvider, projection *models.OrderProjection) error {
	//TODO implement me
	panic("implement me")
}

func (p PostgresRepository) UpdateSubmit(ctx context.Context, provider *trace.TracerProvider, projection *models.OrderProjection) error {
	//TODO implement me
	panic("implement me")
}

func (p PostgresRepository) Delete(ctx context.Context, provider *trace.TracerProvider, projection *models.OrderProjection) error {
	//TODO implement me
	panic("implement me")
}
