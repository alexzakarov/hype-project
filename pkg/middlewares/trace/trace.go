package trace

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/sdk/trace"
	trace_init "main/pkg/trace"
)

type Middleware struct {
	TracerProvider *trace.TracerProvider
}

func (m *Middleware) Public(next Route) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		err := next(ctx, m.TracerProvider)
		return err
	}
}

type Route func(*fiber.Ctx, *trace.TracerProvider) error

func New(endpoint string, serviceName string) *Middleware {
	_, tracerProvider := trace_init.InitTracer(serviceName, endpoint)
	return &Middleware{TracerProvider: tracerProvider}
}
