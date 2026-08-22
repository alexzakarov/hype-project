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
	"math"
	"strings"
)

const (
	count = math.MaxInt64
)

type AggregateStore struct {
	log logger.Logger
	db  *kurrentdb.Client
}

func NewAggregateStore(log logger.Logger, db *kurrentdb.Client) *AggregateStore {
	return &AggregateStore{log: log, db: db}
}

func (a *AggregateStore) Load(ctx context.Context, t *trace.TracerProvider, aggregate es.Aggregate) error {
	var span otracer.Span
	tracer := t.Tracer("Aggregate Load")

	ctx, span = tracer.Start(ctx, "aggregateStore.Load", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", aggregate.GetID()))

	stream, err := a.db.ReadStream(ctx, aggregate.GetID(), kurrentdb.ReadStreamOptions{}, count)
	if err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "db.ReadStream")
	}
	defer stream.Close()

	for {
		event, errStream := stream.Recv()
		if errStream != nil {
			if strings.Contains(errStream.Error(), "is not found") {
				span.RecordError(errStream)
				return errors.Wrap(errStream, "stream.Recv")
			}
			if errors.Is(errStream, io.EOF) {
				break
			}
			span.RecordError(errStream)
			return errors.Wrap(errStream, "stream.Recv")
		}

		esEvent := es.NewEventFromRecorded(event.Event)
		if errStream = aggregate.RaiseEvent(esEvent); errStream != nil {
			span.RecordError(errStream)
			return errors.Wrap(errStream, "RaiseEvent")
		}
		a.log.Debugf("(Load) esEvent: {%s}", esEvent.String())
	}

	a.log.Debugf("(Load) aggregate: {%s}", aggregate.String())
	return nil
}

func (a *AggregateStore) Save(ctx context.Context, t *trace.TracerProvider, aggregate es.Aggregate) error {
	var span otracer.Span
	tracer := t.Tracer("Aggregate Save")

	ctx, span = tracer.Start(ctx, "aggregateStore.Save", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", aggregate.GetID()))

	if len(aggregate.GetUncommittedEvents()) == 0 {
		a.log.Debugf("(Save) [no uncommittedEvents] len: {%d}", len(aggregate.GetUncommittedEvents()))
		return nil
	}

	eventsData := make([]kurrentdb.EventData, 0, len(aggregate.GetUncommittedEvents()))
	for _, event := range aggregate.GetUncommittedEvents() {
		eventsData = append(eventsData, event.ToEventData())
	}

	// check for aggregate.GetVersion() == 0 or len(aggregate.GetAppliedEvents()) == 0 means new aggregate
	var expectedRevision kurrentdb.StreamState
	if aggregate.GetVersion() == 0 {
		expectedRevision = kurrentdb.NoStream{}
		a.log.Debugf("(Save) expectedRevision: {%T}", expectedRevision)

		appendStream, err := a.db.AppendToStream(
			ctx,
			aggregate.GetID(),
			kurrentdb.AppendToStreamOptions{StreamState: expectedRevision},
			eventsData...,
		)
		if err != nil {
			span.RecordError(err)
			return errors.Wrap(err, "db.AppendToStream")
		}

		a.log.Debugf("(Save) stream: {%+v}", appendStream)
		return nil
	}

	readOps := kurrentdb.ReadStreamOptions{Direction: kurrentdb.Backwards, From: kurrentdb.End{}}
	stream, err := a.db.ReadStream(context.Background(), aggregate.GetID(), readOps, 1)
	if err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "db.ReadStream")
	}
	defer stream.Close()

	lastEvent, err := stream.Recv()
	if err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "stream.Recv")
	}

	expectedRevision = kurrentdb.Revision(lastEvent.OriginalEvent().EventNumber)
	a.log.Debugf("(Save) expectedRevision: {%T}", expectedRevision)

	appendStream, err := a.db.AppendToStream(
		ctx,
		aggregate.GetID(),
		kurrentdb.AppendToStreamOptions{StreamState: expectedRevision},
		eventsData...,
	)
	if err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "db.AppendToStream")
	}

	a.log.Debugf("(Save) stream: {%+v}", appendStream)
	aggregate.ClearUncommittedEvents()
	return nil
}

func (a *AggregateStore) Exists(ctx context.Context, t *trace.TracerProvider, streamID string) error {
	var span otracer.Span
	tracer := t.Tracer("Aggregate Save")

	ctx, span = tracer.Start(ctx, "aggregateStore.Save", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", streamID))

	readStreamOptions := kurrentdb.ReadStreamOptions{Direction: kurrentdb.Backwards, From: kurrentdb.Revision(1)}

	stream, err := a.db.ReadStream(ctx, streamID, readStreamOptions, 1)
	if err != nil {
		return errors.Wrap(err, "db.ReadStream")
	}
	defer stream.Close()

	for {
		_, errStream := stream.Recv()
		errApi, _ := kurrentdb.FromError(errStream)
		if errStream != nil && errApi != nil {
			if strings.Contains(errApi.Err().Error(), "is not found") {
				span.RecordError(err)
				return errors.Wrap(errApi.Unwrap(), "stream.Recv")
			}
			if errors.Is(errStream, io.EOF) {
				err = nil
				break
			}
			if errStream != nil {
				span.RecordError(errStream)
				return errors.Wrap(errStream, "stream.Recv")
			}
		}

	}

	return nil
}
