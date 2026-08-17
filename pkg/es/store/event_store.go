package store

import (
	"context"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	otracer "go.opentelemetry.io/otel/trace"
	"io"
	"main/pkg/es"
	"main/pkg/logger"
)

type EventStore struct {
	log logger.Logger
	db  *kurrentdb.Client
}

func NewEventStore(log logger.Logger, db *kurrentdb.Client) *EventStore {
	return &EventStore{log: log, db: db}
}

func (e *EventStore) SaveEvents(ctx context.Context, t *trace.TracerProvider, streamID string, events []es.Event) error {
	var span otracer.Span
	tracer := t.Tracer("EventStore SaveEvents")

	ctx, span = tracer.Start(ctx, "eventStore.SaveEvents", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", streamID))

	eventsData := make([]kurrentdb.EventData, 0, len(events))
	for _, event := range events {
		eventsData = append(eventsData, event.ToEventData())
	}

	stream, err := e.db.AppendToStream(ctx, streamID, kurrentdb.AppendToStreamOptions{}, eventsData...)
	if err != nil {
		span.RecordError(err)
		return err
	}

	e.log.Debugf("SaveEvents stream: %+v", stream)
	return nil
}

func (e *EventStore) LoadEvents(ctx context.Context, t *trace.TracerProvider, streamID string) ([]es.Event, error) {
	var span otracer.Span
	tracer := t.Tracer("EventStore SaveEvents")

	ctx, span = tracer.Start(ctx, "eventStore.LoadEvents", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", streamID))

	stream, err := e.db.ReadStream(ctx, streamID, kurrentdb.ReadStreamOptions{
		Direction: kurrentdb.Forwards,
		From:      kurrentdb.Revision(1),
	}, 100)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer stream.Close()

	events := make([]es.Event, 0, 100)
	for {
		event, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		events = append(events, es.NewEventFromRecorded(event.Event))
	}

	return events, nil
}
