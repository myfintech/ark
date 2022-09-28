package main

import (
	"context"
	"encoding/json"
	"net"
	"strings"
	"time"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_jwt "github.com/myfintech/ark/src/go/lib/grpc/middleware/jwt"

	"github.com/moby/buildkit/util/appcontext"
	"golang.org/x/sync/errgroup"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"

	"google.golang.org/grpc"

	"cloud.google.com/go/pubsub"

	"github.com/myfintech/ark/src/go/lib/ark/components/log_sink_server"

	"github.com/myfintech/ark/src/go/lib/log"
	"github.com/myfintech/ark/src/go/lib/utils"
)

func createNewGRPCServer(ctx context.Context, grpcServer *grpc.Server) error {
	grpcPort := utils.EnvLookup("LOG_SINK_GRPC_PORT", "9000")
	listener, err := net.Listen("tcp", net.JoinHostPort("", grpcPort))
	if err != nil {
		return err
	}

	go func() {
		select {
		case <-ctx.Done():
			log.Infof("shutting down log sink gRPC server")
			timer := time.AfterFunc(time.Second*10, grpcServer.Stop)
			defer timer.Stop()
			defer log.Infof("gRPC server shutdown")
			grpcServer.GracefulStop()
		}
	}()

	if err = grpcServer.Serve(listener); err != nil {
		return err
	}

	return nil
}

func createSubscription(ctx context.Context, client *pubsub.Client, topic *pubsub.Topic, subscription *pubsub.Subscription, es *elasticsearch.Client) error {
	exists, err := subscription.Exists(ctx)
	if err != nil {
		log.Info("subscription already exists")
	}

	if !exists {
		if _, err = client.CreateSubscription(ctx, subscription.ID(), pubsub.SubscriptionConfig{
			Topic:               topic,
			AckDeadline:         time.Second * 10,
			RetainAckedMessages: true,
			RetentionDuration:   time.Hour * 24 * 7,
			ExpirationPolicy:    time.Duration(0),
		}); err != nil {
			return err
		}
	}

	return subscription.Receive(ctx, func(ctx context.Context, message *pubsub.Message) {
		log.Info(message.ID)
		var line *log_sink_server.LogLine
		if err = json.Unmarshal(message.Data, &line); err != nil {
			log.Errorf("there was an error deserializing message data: %v", err)
			message.Nack()
			return
		}

		lineString, convertErr := line.ToJson()
		if convertErr != nil {
			log.Errorf("there was an error converting the log data into a string: %v", convertErr)
			message.Nack()
			return
		}

		req := esapi.IndexRequest{
			Index:      "ark-logs",
			DocumentID: message.ID,
			Body:       strings.NewReader(lineString),
			Pipeline:   "ark-logs-pipeline",
			Refresh:    "true",
			Pretty:     true,
		}

		res, doErr := req.Do(ctx, es)
		defer func() {
			if res != nil {
				_ = res.Body.Close()
			}
		}()
		if doErr != nil {
			log.Errorf("there was an error making a request to ElasticSearch: %v", doErr)
			message.Nack()
			return
		}

		if res.IsError() {
			log.Errorf("there was an error indexing a log line: %s; %s", res.Status(), res.String())
			message.Nack()
			return
		}

		message.Ack()
	})
}

func main() {
	log.Info("getting the party started")
	client, err := pubsub.NewClient(context.Background(), "test")
	if err != nil {
		log.Fatalf("there was an error creating the PubSub client: %v", err)
	}

	logTopic := utils.EnvLookup("ARK_LOG_SINK_TOPIC", "ark-log-sink")
	topic := client.Topic(logTopic)

	logSubscription := utils.EnvLookup("ARK_LOG_SINK_SUBSCRIPTION", "ark-log-sink")
	subscription := client.Subscription(logSubscription)

	eg, ctx := errgroup.WithContext(appcontext.Context())
	logSinkServer := log_sink_server.New(ctx, client, topic, subscription)
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(grpc_jwt.UpdateCtx())),
	)

	log_sink_server.RegisterLogSinkServer(grpcServer, logSinkServer)

	exists, err := topic.Exists(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if !exists {
		if _, err = client.CreateTopic(ctx, logTopic); err != nil {
			log.Fatal(err)
		}
	}

	esConfig := elasticsearch.Config{
		Addresses: []string{utils.EnvLookup("ELASTICSEARCH_URL", "")},
		Username:  utils.EnvLookup("ELASTICSEARCH_USER", ""),
		Password:  utils.EnvLookup("ELASTICSEARCH_PASSWORD", ""),
	}
	es, err := elasticsearch.NewClient(esConfig)

	eg.Go(func() error {
		return createNewGRPCServer(ctx, grpcServer)
	})

	eg.Go(func() error {
		return createSubscription(ctx, client, topic, subscription, es)
	})

	if err = eg.Wait(); err != nil {
		log.Fatal(err)
	}
}
