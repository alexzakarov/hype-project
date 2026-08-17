package service

import (
	"main/config"
	v1 "main/internal/order/commands/v1"
	_interface "main/internal/order/domain/interface"
	"main/internal/order/queries"
	"main/pkg/es"
	"main/pkg/logger"
)

type OrderService struct {
	Commands *v1.OrderCommands
	Queries  *queries.OrderQueries
}

func NewOrderService(
	cfg *config.Config,
	log logger.Logger,
	es es.AggregateStore,
	postgresRepository _interface.IPostgresRepository,
	elasticRepository _interface.IElasticOrderRepository,
) *OrderService {

	createOrderHandler := v1.NewCreateOrderHandler(log, cfg, es)
	orderPaidHandler := v1.NewOrderPaidHandler(log, cfg, es)
	submitOrderHandler := v1.NewSubmitOrderHandler(log, cfg, es)
	updateOrderCmdHandler := v1.NewUpdateShoppingCartCmdHandler(log, cfg, es)
	cancelOrderCommandHandler := v1.NewCancelOrderCommandHandler(log, cfg, es)
	deliveryOrderCommandHandler := v1.NewCompleteOrderCommandHandler(log, cfg, es)
	changeOrderDeliveryAddressCmdHandler := v1.NewChangeDeliveryAddressCmdHandler(log, cfg, es)

	getOrderByIDHandler := queries.NewGetOrderByIDHandler(log, cfg, es, postgresRepository)
	searchOrdersHandler := queries.NewSearchOrdersHandler(log, cfg, es, elasticRepository)

	orderCommands := v1.NewOrderCommands(
		createOrderHandler,
		orderPaidHandler,
		submitOrderHandler,
		updateOrderCmdHandler,
		cancelOrderCommandHandler,
		deliveryOrderCommandHandler,
		changeOrderDeliveryAddressCmdHandler,
	)
	orderQueries := queries.NewOrderQueries(getOrderByIDHandler, searchOrdersHandler)

	return &OrderService{Commands: orderCommands, Queries: orderQueries}
}
