package v1

import (
	"github.com/gofiber/fiber/v2"
	uuid "github.com/satori/go.uuid"
	"go.opentelemetry.io/otel/sdk/trace"
	otracer "go.opentelemetry.io/otel/trace"
	"main/config"
	"main/internal/dto"
	"main/internal/mappers"
	"main/internal/metrics"
	v1 "main/internal/order/commands/v1"
	"main/internal/order/domain/models"
	"main/internal/order/queries"
	"main/internal/order/service"
	"main/pkg/constants"
	httpErrors "main/pkg/http_errors"
	"main/pkg/logger"
	utils "main/pkg/utils/pagination"
	"net/http"
	"time"
)

type OrderHandlers struct {
	cfg     *config.Config
	log     logger.Logger
	os      *service.OrderService
	metrics *metrics.ESMicroserviceMetrics
}

func NewOrderHandlers(
	cfg *config.Config,
	log logger.Logger,
	os *service.OrderService,
	metrics *metrics.ESMicroserviceMetrics,
) IOrderHandlers {
	return &OrderHandlers{cfg: cfg, log: log, os: os, metrics: metrics}
}

// CreateOrder
// @Tags Orders
// @Summary Create order
// @Description Create new order
// @Param order body dto.CreateOrderReqDto true "create order"
// @Accept json
// @Produce json
// @Success 201 {string} id ""
// @Router /orders [post]
func (h *OrderHandlers) CreateOrder(ctx *fiber.Ctx, t *trace.TracerProvider) error {
	var span otracer.Span
	tracer := t.Tracer("OrderHandlers CreateOrder")
	userContext := ctx.UserContext()
	userContext, span = tracer.Start(userContext, "orderHandlers.CreateOrder", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	h.metrics.CreateOrderHttpRequests.Inc()

	var reqDto dto.CreateOrderReqDto
	if err := ctx.BodyParser(&reqDto); err != nil {
		h.log.Errorf("(Bind) err: {%v}", err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	/*
		if err := h.v.StructCtx(ctx, reqDto); err != nil {
			h.log.Errorf("(validate) err: {%v}", err)
			span.RecordError(err)
			return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
		}
	*/

	id := uuid.NewV4().String()
	command := v1.NewCreateOrderCommand(id, reqDto.ShopItems, reqDto.AccountEmail, reqDto.DeliveryAddress)
	err := h.os.Commands.CreateOrder.Handle(userContext, t, command)
	if err != nil {
		h.log.Errorf("(CreateOrder.Handle) id: {%s}, err: {%v}", id, err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	query := queries.NewGetOrderByIDQuery(id)
	data, errQuery := h.os.Queries.GetOrderByID.Handle(userContext, t, query)
	if errQuery != nil {
		h.log.Errorf("(CreateOrder.Query.Handle) id: {%s}, err: {%v}", id, errQuery)
		span.RecordError(errQuery)
		return httpErrors.ErrorCtxResponse(ctx, errQuery, h.cfg.Http.DebugErrorsResponse)

	}

	h.log.Infof("(order created) id: {%s}", id)
	return ctx.Status(http.StatusCreated).JSON(data)
}

// PayOrder
// @Tags Orders
// @Summary Pay order
// @Description Pay existing order
// @Accept json
// @Produce json
// @Param order body dto.Payment true "create order"
// @Param id path string true "Order ID"
// @Success 200 {string} id ""
// @Router /orders/pay/{id} [put]
func (h *OrderHandlers) PayOrder(ctx *fiber.Ctx, t *trace.TracerProvider) error {
	var span otracer.Span
	tracer := t.Tracer("OrderHandlers PayOrder")
	userContext := ctx.UserContext()
	userContext, span = tracer.Start(userContext, "orderHandlers.PayOrder", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	h.metrics.PayOrderHttpRequests.Inc()

	orderID, err := uuid.FromString(ctx.Params(constants.ID, ""))
	if err != nil {
		h.log.Errorf("(uuid.FromString) err: {%v}", err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	var payment dto.Payment
	if err := ctx.BodyParser(&payment); err != nil {
		h.log.Errorf("(Bind) err: {%v}", err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	command := v1.NewPayOrderCommand(models.Payment{PaymentID: payment.PaymentID, Timestamp: payment.Timestamp}, orderID.String())
	/*
		if err := h.v.StructCtx(ctx, command); err != nil {
			h.log.Errorf("(validate) err: {%v}", err)
			span.RecordError(err)
			return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
		}
	*/

	err = h.os.Commands.OrderPaid.Handle(userContext, t, command)
	if err != nil {
		h.log.Errorf("(OrderPaid.Handle) id: {%s}, err: {%v}", orderID.String(), err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	h.log.Infof("(order paid) id: {%s}", orderID.String())
	return ctx.Status(http.StatusOK).JSON(orderID.String())

}

// SubmitOrder
// @Tags Orders
// @Summary Submit order
// @Description Submit existing order
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {string} id ""
// @Router /orders/submit/{id} [put]
func (h *OrderHandlers) SubmitOrder(ctx *fiber.Ctx, t *trace.TracerProvider) error {
	var span otracer.Span
	tracer := t.Tracer("OrderHandlers SubmitOrder")
	userContext := ctx.UserContext()
	userContext, span = tracer.Start(userContext, "orderHandlers.SubmitOrder", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	h.metrics.SubmitOrderHttpRequests.Inc()

	orderID, err := uuid.FromString(ctx.Params(constants.ID, ""))
	if err != nil {
		h.log.Errorf("(uuid.FromString) err: {%v}", err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	command := v1.NewSubmitOrderCommand(orderID.String())
	/*
		if err := h.v.StructCtx(ctx, command); err != nil {
			h.log.Errorf("(validate) err: {%v}", err)
			span.RecordError(err)
			return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
		}
	*/

	err = h.os.Commands.SubmitOrder.Handle(userContext, t, command)
	if err != nil {
		h.log.Errorf("(SubmitOrder.Handle) id: {%s}, err: {%v}", orderID.String(), err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	h.log.Infof("(order submitted) id: {%s}", orderID.String())
	return ctx.Status(http.StatusOK).JSON(orderID.String())
}

// CancelOrder
// @Tags Orders
// @Summary Cancel order
// @Description Cancel existing order
// @Accept json
// @Produce json
// @Param order body dto.CancelOrderReqDto true "cancel order reason"
// @Param id path string true "Order ID"
// @Success 200 {string} id ""
// @Router /orders/cancel/{id} [post]
func (h *OrderHandlers) CancelOrder(ctx *fiber.Ctx, t *trace.TracerProvider) error {
	var span otracer.Span
	tracer := t.Tracer("OrderHandlers CancelOrder")
	userContext := ctx.UserContext()
	userContext, span = tracer.Start(userContext, "orderHandlers.CancelOrder", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	h.metrics.SubmitOrderHttpRequests.Inc()

	orderID, err := uuid.FromString(ctx.Params(constants.ID, ""))
	if err != nil {
		h.log.Errorf("(uuid.FromString) err: {%v}", err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	var data dto.CancelOrderReqDto
	if err := ctx.QueryParser(&data); err != nil {
		h.log.Errorf("(Bind) err: {%v}", err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	command := v1.NewCancelOrderCommand(orderID.String(), data.CancelReason)
	/*
		if err := h.v.StructCtx(ctx, command); err != nil {
			h.log.Errorf("(validate) err: {%v}", err)
			span.RecordError(err)
			return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
		}
	*/
	err = h.os.Commands.CancelOrder.Handle(userContext, t, command)
	if err != nil {
		h.log.Errorf("(CancelOrder.Handle) id: {%s}, err: {%v}", orderID.String(), err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	h.log.Infof("(order canceled) id: {%s}", orderID.String())
	return ctx.Status(http.StatusOK).JSON(orderID.String())
}

// CompleteOrder
// @Tags Orders
// @Summary Complete order
// @Description Complete existing order
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {string} id ""
// @Router /orders/complete/{id} [post]
func (h *OrderHandlers) CompleteOrder(ctx *fiber.Ctx, t *trace.TracerProvider) error {
	var span otracer.Span
	tracer := t.Tracer("OrderHandlers CompleteOrder")
	userContext := ctx.UserContext()
	userContext, span = tracer.Start(userContext, "orderHandlers.CompleteOrder", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	h.metrics.CompleteOrderHttpRequests.Inc()

	orderID, err := uuid.FromString(ctx.Params(constants.ID, ""))
	if err != nil {
		h.log.Errorf("(uuid.FromString) err: {%v}", err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	command := v1.NewCompleteOrderCommand(orderID.String(), time.Now())
	/*
		if err := h.v.StructCtx(ctx, command); err != nil {
			h.log.Errorf("(validate) err: {%v}", err)
			span.RecordError(err)
			return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
		}
	*/

	err = h.os.Commands.CompleteOrder.Handle(userContext, t, command)
	if err != nil {
		h.log.Errorf("(CompleteOrder.Handle) id: {%s}, err: {%v}", orderID.String(), err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	h.log.Infof("(order delivered) id: {%s}", orderID.String())
	return ctx.Status(http.StatusOK).JSON(orderID.String())
}

// ChangeDeliveryAddress
// @Tags Orders
// @Summary Change delivery address order
// @Description Deliver existing order
// @Accept json
// @Produce json
// @Param order body dto.ChangeDeliveryAddressReqDto true "change delivery address"
// @Param id path string true "Order ID"
// @Success 200 {string} id ""
// @Router /orders/address/{id} [put]
func (h *OrderHandlers) ChangeDeliveryAddress(ctx *fiber.Ctx, t *trace.TracerProvider) error {
	var span otracer.Span
	tracer := t.Tracer("OrderHandlers ChangeDeliveryAddress")
	userContext := ctx.UserContext()
	userContext, span = tracer.Start(userContext, "orderHandlers.ChangeDeliveryAddress", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	h.metrics.ChangeAddressOrderHttpRequests.Inc()

	param := ctx.Params(constants.ID, "")
	orderID, err := uuid.FromString(param)
	if err != nil {
		h.log.Errorf("(uuid.FromString) err: {%v}", err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	var data dto.ChangeDeliveryAddressReqDto
	if err := ctx.QueryParser(&data); err != nil {
		h.log.Errorf("(Bind) err: {%v}", err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	command := v1.NewChangeDeliveryAddressCommand(orderID.String(), data.DeliveryAddress)
	/*
		if err := h.v.StructCtx(ctx, command); err != nil {
			h.log.Errorf("(validate) err: {%v}", err)
			span.RecordError(err)
			return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
		}
	*/

	err = h.os.Commands.ChangeOrderDeliveryAddress.Handle(userContext, t, command)
	if err != nil {
		h.log.Errorf("(ChangeOrderDeliveryAddress.Handle) id: {%s}, err: {%v}", orderID.String(), err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	h.log.Infof("(ChangeDeliveryAddress) id: {%s}", orderID.String())
	return ctx.Status(http.StatusOK).JSON(orderID.String())
}

// UpdateShoppingCart
// @Tags Orders
// @Summary Update order shopping cart
// @Description Update existing order shopping cart
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param order body dto.UpdateShoppingItemsReqDto true "update order"
// @Success 200 {string} id ""
// @Router /orders/cart/{id} [put]
func (h *OrderHandlers) UpdateShoppingCart(ctx *fiber.Ctx, t *trace.TracerProvider) error {
	var span otracer.Span
	tracer := t.Tracer("OrderHandlers UpdateShoppingCart")
	userContext := ctx.UserContext()
	userContext, span = tracer.Start(userContext, "orderHandlers.UpdateShoppingCart", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	h.metrics.UpdateOrderHttpRequests.Inc()

	orderID, err := uuid.FromString(ctx.Params(constants.ID, ""))
	if err != nil {
		h.log.Errorf("(uuid.FromString) err: {%v}", err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	var reqDto dto.UpdateShoppingItemsReqDto
	if err := ctx.QueryParser(&reqDto); err != nil {
		h.log.Errorf("(Bind) err: {%v}", err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	/*
		if err := h.v.StructCtx(ctx, reqDto); err != nil {
			h.log.Errorf("(validate) err: {%v}", err)
			span.RecordError(err)
			return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
		}
	*/

	command := v1.NewUpdateShoppingCartCommand(orderID.String(), reqDto.ShopItems)
	err = h.os.Commands.UpdateOrder.Handle(userContext, t, command)
	if err != nil {
		h.log.Errorf("(UpdateShoppingCart.Handle) id: {%s}, err: {%v}", orderID.String(), err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	h.log.Infof("(order updated) id: {%s}", orderID.String())
	return ctx.Status(http.StatusOK).JSON(orderID.String())

}

// GetOrderByID
// @Tags Orders
// @Summary Get order
// @Description Get order by id
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} dto.OrderResponseDto
// @Router /orders/{id} [get]
func (h *OrderHandlers) GetOrderByID(ctx *fiber.Ctx, t *trace.TracerProvider) error {
	var span otracer.Span
	tracer := t.Tracer("OrderHandlers GetOrderByID")
	userContext := ctx.UserContext()
	userContext, span = tracer.Start(userContext, "orderHandlers.GetOrderByID", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	h.metrics.GetOrderByIdHttpRequests.Inc()

	param := ctx.Params(constants.ID, "")
	orderID, err := uuid.FromString(param)
	if err != nil {
		h.log.Errorf("(uuid.FromString) err: {%v}", err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	query := queries.NewGetOrderByIDQuery(orderID.String())
	/*
		if err := h.v.StructCtx(ctx, query); err != nil {
			h.log.Errorf("(validate) err: {%v}", err)
			span.RecordError(err)
			return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
		}
	*/

	orderProjection, err := h.os.Queries.GetOrderByID.Handle(userContext, t, query)
	if err != nil {
		h.log.Errorf("(GetOrderByID.Handle) id: {%s}, err: {%v}", orderID.String(), err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	h.log.Infof("(get order by id) orderID: {%s}", orderID.String())
	return ctx.Status(http.StatusOK).JSON(mappers.OrderResponseFromProjection(orderProjection))
}

// Search
// @Tags Orders
// @Summary Search orders
// @Description Full text search by title and description
// @Accept json
// @Produce json
// @Param search query string false "search text"
// @Param page query string false "page number"
// @Param size query string false "number of elements"
// @Success 200 {object} dto.OrderSearchResponseDto
// @Router /orders/search [get]
func (h *OrderHandlers) Search(ctx *fiber.Ctx, t *trace.TracerProvider) error {
	var span otracer.Span
	tracer := t.Tracer("OrderHandlers GetOrderByID")
	userContext := ctx.UserContext()
	userContext, span = tracer.Start(userContext, "orderHandlers.GetOrderByID", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	h.metrics.SearchOrderHttpRequests.Inc()

	pq := utils.NewPaginationFromQueryParams(ctx.Params(constants.Size, "1"), ctx.Params(constants.Page, "1"))

	query := queries.NewSearchOrdersQuery(ctx.Params(constants.Search), pq)
	/*
		if err := h.v.StructCtx(ctx, query); err != nil {
			h.log.Errorf("(validate) err: {%v}", err)
			span.RecordError(err)
			return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
		}
	*/

	searchRes, err := h.os.Queries.SearchOrders.Handle(userContext, t, query)
	if err != nil {
		h.log.Errorf("(SearchOrders.Handle): Search: {%s}, err: {%v}", ctx.Params(constants.Search, ""), err)
		span.RecordError(err)
		return httpErrors.ErrorCtxResponse(ctx, err, h.cfg.Http.DebugErrorsResponse)
	}

	h.log.Infof("(search) result: {%+v}", searchRes)
	return ctx.Status(http.StatusOK).JSON(searchRes)

}
