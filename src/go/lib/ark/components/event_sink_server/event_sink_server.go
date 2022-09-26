package event_sink_server

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cloud.google.com/go/pubsub"
)

type eventSinkServer struct {
	Ctx          context.Context
	Client       *pubsub.Client
	Topic        *pubsub.Topic
	Subscription *pubsub.Subscription
}

// RecordCanonicalEvent receives an event and publishes it to a given topic as a PubSub Message
func (server *eventSinkServer) RecordCanonicalEvent(ctx context.Context, event *CanonicalEvent) (*RecordCanonicalEventResponse, error) {
	userId := server.Ctx.Value("user_id").(string)
	event.UserId = userId

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to serialize event into json: %v", err)
	}

	pubRes := server.Topic.Publish(server.Ctx, &pubsub.Message{
		Data: eventBytes,
	})

	if _, err = pubRes.Get(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish event to topic: %v", err)
	}

	return &RecordCanonicalEventResponse{
		Ok: true,
	}, nil
}

// New creates a new event sink server
func New(ctx context.Context, client *pubsub.Client, topic *pubsub.Topic, subscription *pubsub.Subscription) *eventSinkServer {
	return &eventSinkServer{
		Ctx:          ctx,
		Client:       client,
		Topic:        topic,
		Subscription: subscription,
	}
}
