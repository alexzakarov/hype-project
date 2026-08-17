package v1

type OrderCommands struct {
	CreateOrder                ICreateOrderCommandHandler
	OrderPaid                  IPayOrderCommandHandler
	SubmitOrder                ISubmitOrderCommandHandler
	UpdateOrder                IUpdateShoppingCartCommandHandler
	CancelOrder                ICancelOrderCommandHandler
	CompleteOrder              ICompleteOrderCommandHandler
	ChangeOrderDeliveryAddress IChangeDeliveryAddressCommandHandler
}

func NewOrderCommands(
	createOrder ICreateOrderCommandHandler,
	orderPaid IPayOrderCommandHandler,
	submitOrder ISubmitOrderCommandHandler,
	updateOrder IUpdateShoppingCartCommandHandler,
	cancelOrder ICancelOrderCommandHandler,
	deliveryOrder ICompleteOrderCommandHandler,
	changeOrderDeliveryAddress IChangeDeliveryAddressCommandHandler,
) *OrderCommands {
	return &OrderCommands{
		CreateOrder:                createOrder,
		OrderPaid:                  orderPaid,
		SubmitOrder:                submitOrder,
		UpdateOrder:                updateOrder,
		CancelOrder:                cancelOrder,
		CompleteOrder:              deliveryOrder,
		ChangeOrderDeliveryAddress: changeOrderDeliveryAddress,
	}
}
