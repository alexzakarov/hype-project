package v1

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/sdk/trace"
	tracerMiddleware "main/pkg/middlewares/trace"
)

type IOrderHandlers interface {
	MapRoutes(*tracerMiddleware.Middleware, fiber.Router)
	CreateOrder(*fiber.Ctx, *trace.TracerProvider) error
	PayOrder(*fiber.Ctx, *trace.TracerProvider) error
	SubmitOrder(*fiber.Ctx, *trace.TracerProvider) error
	UpdateShoppingCart(*fiber.Ctx, *trace.TracerProvider) error

	GetOrderByID(*fiber.Ctx, *trace.TracerProvider) error
	Search(*fiber.Ctx, *trace.TracerProvider) error
}
