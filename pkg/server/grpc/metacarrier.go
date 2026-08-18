package grpc

import (
	"context"
	"encoding/json"
	"github.com/google/martian/log"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/metadata"
	"strings"
)

type MetadataCarrier struct {
	Md metadata.MD
}

func (c *MetadataCarrier) Get(key string) string {
	values := c.Md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (c *MetadataCarrier) Set(key, value string) {
	c.Md.Set(strings.ToLower(key), value)
}

func (c *MetadataCarrier) Keys() []string {
	keys := make([]string, 0, len(c.Md))

	for key := range c.Md {
		keys = append(keys, key)
	}

	return keys
}

func InjectTraceContext(ctx context.Context) context.Context {
	md := metadata.New(nil)

	otel.GetTextMapPropagator().Inject(
		ctx,
		&MetadataCarrier{Md: md},
	)

	return metadata.NewOutgoingContext(ctx, md)
}

func InjectTraceContextAndConvertToMD(ctx context.Context) metadata.MD {
	md := metadata.New(nil)

	otel.GetTextMapPropagator().Inject(
		ctx,
		&MetadataCarrier{Md: md},
	)
	metadata.NewOutgoingContext(ctx, md)
	return md
}

func ExtractTraceContext(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)

	if ok {
		ctx = otel.GetTextMapPropagator().Extract(
			ctx,
			&MetadataCarrier{Md: md},
		)
	}
	return ctx
}

func ExtractTraceContextFromEvent(ctx context.Context, event *kurrentdb.EventAppeared) context.Context {
	md := metadata.New(nil)

	if event == nil {
		return ctx
	}
	errJsonParse := json.Unmarshal(event.Event.Event.UserMetadata, &md)
	if errJsonParse != nil {
		log.Errorf("(ProcessEvents) err: {%v}", errJsonParse.Error())
	}

	ctx = otel.GetTextMapPropagator().Extract(
		ctx,
		&MetadataCarrier{Md: md},
	)
	return ctx

}
