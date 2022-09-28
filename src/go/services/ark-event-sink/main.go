package main

import (
	"context"
	"encoding/json"
	"net"
	"time"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_jwt "github.com/myfintech/ark/src/go/lib/grpc/middleware/jwt"

	"cloud.google.com/go/bigquery"

	"cloud.google.com/go/pubsub"
	"github.com/moby/buildkit/util/appcontext"
	"golang.org/x/sync/errgroup"

	"google.golang.org/grpc"

	"github.com/myfintech/ark/src/go/lib/log"
	"github.com/myfintech/ark/src/go/lib/utils"

	"github.com/myfintech/ark/src/go/lib/ark/components/event_sink_server"
)

func createNewGrpcServer(ctx context.Context, grpcServer *grpc.Server) error {
	grpcPort := utils.EnvLookup("EVENT_SINK_GRPC_PORT", "9000")
	lis, err := net.Listen("tcp", net.JoinHostPort("", grpcPort))
	if err != nil {
		return err
	}

	go func() {
		select {
		case <-ctx.Done():
			log.Infof("shutting down event sink gRPC server")
			timer := time.AfterFunc(time.Second*10, grpcServer.Stop)
			defer timer.Stop()
			defer log.Infof("gRPC server shutdown")
			grpcServer.GracefulStop()
		}
	}()

	if err = grpcServer.Serve(lis); err != nil {
		return err
	}
	return nil
}

func createSubscription(ctx context.Context, client *pubsub.Client, topic *pubsub.Topic, subscription *pubsub.Subscription) error {
	exists, err := subscription.Exists(ctx)
	if err != nil {
		log.Print("subscription already exists")
	}

	if !exists {
		_, subErr := client.CreateSubscription(ctx, subscription.ID(), pubsub.SubscriptionConfig{
			Topic:               topic,
			AckDeadline:         time.Second * 10,
			RetainAckedMessages: true,
			RetentionDuration:   time.Hour * 24 * 7,
			ExpirationPolicy:    time.Duration(0),
		})
		if subErr != nil {
			return subErr
		}
	}

	return subscription.Receive(ctx, func(ctx context.Context, message *pubsub.Message) {
		log.Info(message.ID)
		var event *event_sink_server.CanonicalEvent
		if err = json.Unmarshal(message.Data, &event); err != nil {
			log.Errorf("there was an error deserializing message data: %v", err)
			return
		}
		bqClient, bqErr := bigquery.NewClient(ctx, "[insert-google-project]")
		if bqErr != nil {
			log.Errorf("there was an error creating BigQuery client: %v", err)
			return
		}

		uploader := bqClient.Dataset("ark_dev").Table("events").Inserter()
		if err = uploader.Put(ctx, event); err != nil {
			log.Errorf("there was an error uploading message to BigQuery: %v", err)
			return
		}

		message.Ack()
	})
}

func main() {
	client, err := pubsub.NewClient(context.Background(), "test")
	if err != nil {
		log.Fatalf("there was an error creating the PubSub client: %v", err)
	}

	topicID := utils.EnvLookup("ARK_EVENT_SINK_TOPIC", "ark-event-sink")
	topic := client.Topic(topicID)

	subscriptionID := utils.EnvLookup("ARK_EVENT_SINK_SUBSCRIPTION", "ark-event-sink")
	subscription := client.Subscription(subscriptionID)

	eg, ctx := errgroup.WithContext(appcontext.Context())
	eventSinkServer := event_sink_server.New(ctx, client, topic, subscription)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(grpc_jwt.UpdateCtx())),
	)

	event_sink_server.RegisterEventSinkServer(grpcServer, eventSinkServer)

	exists, err := topic.Exists(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if !exists {
		_, err = client.CreateTopic(ctx, topicID)
		if err != nil {
			log.Fatal(err)
		}
	}

	eg.Go(func() error {
		return createNewGrpcServer(ctx, grpcServer)
	})

	eg.Go(func() error {
		return createSubscription(ctx, client, topic, subscription)
	})

	if err = eg.Wait(); err != nil {
		log.Fatal(err)
	}
}
