package log_sink_server

import (
	"context"
	"encoding/json"
	"io"
	"time"

	grpc_jwt "github.com/myfintech/ark/src/go/lib/grpc/middleware/jwt"

	"github.com/myfintech/ark/src/go/lib/log"

	"cloud.google.com/go/pubsub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type logSinkServer struct {
	Ctx          context.Context
	Client       *pubsub.Client
	Topic        *pubsub.Topic
	Subscription *pubsub.Subscription
}

// Record publishes log data to pubsub for later consumption and storing in ElasticSearch
func (server *logSinkServer) Record(stream LogSink_RecordServer) error {
	log.Info("Connected")
	for {
		line, err := stream.Recv()
		if err == io.EOF {
			log.Info("Disconnected")
			return nil
		}
		if err != nil {
			log.Errorf("there was an error receiving from the stream: %v", err)
			return status.Error(codes.Internal, err.Error())
		}

		if len(line.Data) == 0 {
			return status.Error(codes.InvalidArgument, "data field is empty")
		}

		userId := stream.Context().Value(grpc_jwt.ContextKey("user_id"))
		if userId == nil || userId.(string) == "" {
			log.Info("The 'user_id' field was empty or not set")
			return status.Error(codes.InvalidArgument, "'user_id' field is empty or not set")
		}

		// records the user ID pulled from the middleware and sets the line reception time
		line.UserId = userId.(string)
		line.ReceivedAt = time.Now().Format(time.RFC3339)

		logBytes, err := json.Marshal(line)
		if err != nil {
			log.Errorf("there was an error serializing log data to json: %v", err)
			return status.Errorf(codes.Internal, "failed to serialize log data into json: %v", err)
		}

		pubRes := server.Topic.Publish(stream.Context(), &pubsub.Message{
			Data: logBytes,
		})
		if _, err = pubRes.Get(stream.Context()); err != nil {
			log.Errorf("there was an error publishing to pubsub: %v", err)
			return status.Errorf(codes.Internal, "failed to publish log data to topic: %v", err)
		}

		if err = stream.Send(&RecordLogLineAck{}); err != nil {
			log.Errorf("there was an error acking a log line: %v", err)
			return status.Errorf(codes.Unknown, err.Error())
		}
	}
}

// New creates a new log sink server
func New(ctx context.Context, client *pubsub.Client, topic *pubsub.Topic, subscription *pubsub.Subscription) *logSinkServer {
	return &logSinkServer{
		Ctx:          ctx,
		Client:       client,
		Topic:        topic,
		Subscription: subscription,
	}
}

// ToJson maps the LogLine struct into a JSON object
func (l *LogLine) ToJson() (string, error) {
	blob := map[string]interface{}{
		"user_id":        l.UserId,
		"target_address": l.TargetAddress,
		"target_hash":    l.TargetHash,
		"data":           string(l.Data),
		"created_at":     l.CreatedAt,
		"received_at":    l.ReceivedAt,
		"org_id":         l.OrgId,
		"project_id":     l.ProjectId,
	}
	jsonBytes, err := json.Marshal(blob)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
