package grpc

import (
	"context"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/metadata"
	"strings"
)

type metadataCarrier struct {
	md metadata.MD
}

func (c *metadataCarrier) Get(key string) string {
	values := c.md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (c *metadataCarrier) Set(key, value string) {
	c.md.Set(strings.ToLower(key), value)
}

func (c *metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(c.md))

	for key := range c.md {
		keys = append(keys, key)
	}

	return keys
}

func InjectTraceContext(ctx context.Context) context.Context {
	md := metadata.New(nil)

	otel.GetTextMapPropagator().Inject(
		ctx,
		&metadataCarrier{md: md},
	)

	return metadata.NewOutgoingContext(ctx, md)
}

func ExtractTraceContext(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)

	if ok {
		ctx = otel.GetTextMapPropagator().Extract(
			ctx,
			&metadataCarrier{md: md},
		)
	}
	return ctx
}
