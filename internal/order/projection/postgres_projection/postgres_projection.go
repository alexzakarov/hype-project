package postgres_projection

import (
	"context"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	otracer "go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
	"main/config"
	_interface "main/internal/order/domain/interface"
	v1 "main/internal/order/events/v1"
	"main/pkg/constants"
	"main/pkg/es"
	"main/pkg/logger"
	"main/pkg/server/grpc"
)

type PostgresProjection struct {
	log          logger.Logger
	db           *kurrentdb.Client
	cfg          *config.Config
	postgresRepo _interface.IPostgresRepository
}

func NewOrderProjection(log logger.Logger, db *kurrentdb.Client, postgresRepo _interface.IPostgresRepository, cfg *config.Config) *PostgresProjection {
	return &PostgresProjection{log: log, db: db, postgresRepo: postgresRepo, cfg: cfg}
}

type Worker func(ctx context.Context, t *trace.TracerProvider, stream *kurrentdb.PersistentSubscription, workerID int) error

func (o *PostgresProjection) Subscribe(ctx context.Context, t *trace.TracerProvider, prefixes []string, poolSize int, worker Worker) error {
	o.log.Infof("(starting order subscription) prefixes: {%+v}", prefixes)

	err := o.db.CreatePersistentSubscriptionToAll(ctx, o.cfg.Subscriptions.PostgresProjectionGroupName, kurrentdb.PersistentAllSubscriptionOptions{
		Filter: &kurrentdb.SubscriptionFilter{Type: kurrentdb.StreamFilterType, Prefixes: prefixes},
	})
	if err != nil {
		if subscriptionError, ok := err.(*kurrentdb.Error); !ok {
			o.log.Errorf("(CreatePersistentSubscriptionAll) err: {%v}", subscriptionError.Error())
		}
	}

	stream, err := o.db.SubscribeToPersistentSubscriptionToAll(
		ctx,
		o.cfg.Subscriptions.PostgresProjectionGroupName,
		kurrentdb.SubscribeToPersistentSubscriptionOptions{},
	)
	if err != nil {
		return err
	}
	defer stream.Close()

	g, ctx := errgroup.WithContext(ctx)
	for i := 0; i <= poolSize; i++ {
		g.Go(o.runWorker(ctx, t, worker, stream, i))
	}
	return g.Wait()

	/*
		g, ctx := errgroup.WithContext(ctx)
		for i := 0; i <= poolSize; i++ {
			stream, err := o.db.SubscribeToPersistentSubscriptionToAll(
				ctx,
				o.cfg.Subscriptions.PostgresProjectionGroupName,
				kurrentdb.SubscribeToPersistentSubscriptionOptions{},
			)
			if err != nil {
				return err
			}
			g.Go(func() error {
				defer stream.Close()
				return worker(ctx, t, stream, i)
			})
		}
	*/
}

func (o *PostgresProjection) runWorker(ctx context.Context, t *trace.TracerProvider, worker Worker, stream *kurrentdb.PersistentSubscription, i int) func() error {
	return func() error {
		return worker(ctx, t, stream, i)
	}
}

func (o *PostgresProjection) ProcessEvents(ctx context.Context, t *trace.TracerProvider, stream *kurrentdb.PersistentSubscription, workerID int) error {
	for {
		event := stream.Recv()
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:

			if event.SubscriptionDropped != nil {
				o.log.Errorf("(SubscriptionDropped) err: {%v}", event.SubscriptionDropped.Error)
				return errors.Wrap(event.SubscriptionDropped.Error, "Subscription Dropped")
			}

			if event.EventAppeared == nil {
				continue
			}

			var span otracer.Span

			ctx = grpc.ExtractTraceContextFromEvent(ctx, event.EventAppeared)
			tracer := t.Tracer("PostgresProjection ProcessEvents")
			ctx, span = tracer.Start(ctx, "postgresProjection.ProcessEvents", otracer.WithSpanKind(otracer.SpanKindServer))

			o.log.ProjectionEvent(constants.PostgresProjection, o.cfg.Subscriptions.PostgresProjectionGroupName, event.EventAppeared.Event, workerID)

			err := o.When(ctx, t, es.NewEventFromRecorded(event.EventAppeared.Event.Event))
			if err != nil {
				//span.RecordError(event.SubscriptionDropped.Error)
				o.log.Errorf("(mongoProjection.when) err: {%v}", err)

				if err := stream.Nack(err.Error(), kurrentdb.NackActionRetry, event.EventAppeared.Event); err != nil {
					o.log.Errorf("(stream.Nack) err: {%v}", err)
					return errors.Wrap(err, "stream.Nack")
				}
			}

			err = stream.Ack(event.EventAppeared.Event)
			if err != nil {
				//span.RecordError(event.SubscriptionDropped.Error)
				o.log.Errorf("(stream.Ack) err: {%v}", err)
				return errors.Wrap(err, "stream.Ack")
			}
			o.log.Infof("(ACK) event commit: {%v}", *event.EventAppeared.Event.Commit)
			span.End()
		}
	}
}

func (o *PostgresProjection) When(ctx context.Context, t *trace.TracerProvider, evt es.Event) error {
	var span otracer.Span
	tracer := t.Tracer("PostgresProjection When")

	ctx, span = tracer.Start(ctx, "postgresProjection.When", otracer.WithSpanKind(otracer.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("AggregateID", evt.GetAggregateID()), attribute.String("EventType", evt.GetEventType()))

	switch evt.GetEventType() {

	case v1.OrderCreated:
		return o.onOrderCreate(ctx, t, evt)
	case v1.OrderPaid:
		return o.onOrderPaid(ctx, t, evt)
	case v1.OrderSubmitted:
		return o.onSubmit(ctx, t, evt)
	case v1.ShoppingCartUpdated:
		return o.onShoppingCartUpdate(ctx, t, evt)
	case v1.OrderCanceled:
		return o.onCancel(ctx, t, evt)
	case v1.OrderCompleted:
		return o.onCompleted(ctx, t, evt)
	case v1.DeliveryAddressChanged:
		return o.onDeliveryAddressChnaged(ctx, t, evt)

	default:
		o.log.Warnf("(mongoProjection) [When unknown EventType] eventType: {%s}", evt.EventType)
		return es.ErrInvalidEventType
	}
}
