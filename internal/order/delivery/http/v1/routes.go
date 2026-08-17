package v1

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"main/pkg/middlewares/trace"
)

func (h *OrderHandlers) MapRoutes(tracerMiddleware *trace.Middleware, router fiber.Router) {
	core := router.Group("orders")

	core.Post("", tracerMiddleware.Public(h.CreateOrder))
	core.Put("/pay/:id", tracerMiddleware.Public(h.PayOrder))
	core.Put("/submit/:id", tracerMiddleware.Public(h.SubmitOrder))
	core.Post("/cancel/:id", tracerMiddleware.Public(h.CancelOrder))
	core.Post("/complete/:id", tracerMiddleware.Public(h.CompleteOrder))
	core.Put("/address/:id", tracerMiddleware.Public(h.ChangeDeliveryAddress))
	core.Get("/:id", tracerMiddleware.Public(h.GetOrderByID))
	core.Get("/search", tracerMiddleware.Public(h.Search))
	fmt.Println("Routes are done")
}
