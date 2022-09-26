package kafka

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/google/uuid"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func TestBroker(t *testing.T) {
	if os.Getenv("KAFKA_TESTS_ENABLED") != "true" {
		t.Skip("kafka tests not enabled")
		return
	}

	ctx := context.Background()

	config := &kafka.ConfigMap{
		"allow.auto.create.topics": true,
		"auto.commit.interval.ms":  5000,
		"auto.offset.reset":        "earliest",
		"bootstrap.servers":        "localhost:29092",
		"enable.auto.commit":       false,
		"enable.auto.offset.store": true,
		"group.id":                 uuid.New().String(),
	}

	client, err := kafka.NewAdminClient(config)
	require.NoError(t, err)

	topic := fmt.Sprintf("cqrs.kafka.integration.test.%s", uuid.New().String())
	t.Log(topic)

	_, _ = client.CreateTopics(ctx, []kafka.TopicSpecification{{
		Topic:         topic,
		NumPartitions: 1,
	}})

	producer, err := kafka.NewProducer(config)
	require.NoError(t, err)

	consumer, err := kafka.NewConsumer(config)
	require.NoError(t, err)

	broker := Broker{
		Producer: producer,
		Consumer: consumer,
		Logger:   logz.NoOpLogger{},
	}

	waitToPublish := make(chan struct{})

	go func() {
		<-waitToPublish
		err = broker.Publish(cqrs.RouteKey(topic), cqrs.NewDefaultEnvelope(
			cqrs.WithSource("source"),
			cqrs.WithSubject("subject"),
			cqrs.WithType("test.event"),
			cqrs.WithData(cqrs.TextPlain, "This is a test message"),
		))
		require.NoError(t, err)
	}()

	ackDeadline := time.Second * 60

	messages, err := broker.Subscribe(ctx, cqrs.RouteKey(topic), &ackDeadline)
	require.NoError(t, err)
	close(waitToPublish)

	var retryErr = errors.New("retry error")
	select {
	case message := <-messages:
		require.NotNil(t, message)
		require.Equal(t, "source", message.Source())
		require.Equal(t, "subject", message.Subject())
		require.Equal(t, "test.event", message.Type())
		require.Equal(t, "This is a test message", string(message.Data()))
		message.Reject(retryErr)
	case <-time.After(time.Second * 30):
		t.Error("did not receive test message within timeout")
		return
	}
	select {
	case message := <-messages:
		require.NotNil(t, message)
		require.Equal(t, "0", message.Header.Get(RetryCountKey))
		require.Equal(t, "60s", message.Header.Get(RetryDelayKey))
	case <-time.After(time.Second * 30):
		t.Error("did not receive test message within timeout")
		return
	}
}

func TestHeaders(t *testing.T) {
	message := cqrs.NewDefaultEnvelope(
		cqrs.WithSource("source"),
		cqrs.WithSubject("subject"),
		cqrs.WithType("test.event"),
		cqrs.WithData(cqrs.TextPlain, "This is a test message"),
	)

	message.Header.Add(RetryCountKey, "0")
	message.Header.Add(RetryDelayKey, "60s")

	kHeaders := toKafkaHeaders(message)
	sort.Slice(kHeaders, func(i, j int) bool {
		return kHeaders[i].Key > kHeaders[j].Key
	})
	require.Equal(t, []kafka.Header(
		[]kafka.Header{
			{
				Key:   RetryDelayKey,
				Value: []byte("60s"),
			},
			{
				Key:   RetryCountKey,
				Value: []byte("0"),
			},
		}), kHeaders)

	envelope := cqrs.NewDefaultEnvelope(fromKafkaHeaders(kHeaders))
	require.Equal(t, http.Header{
		RetryCountKey: []string{"0"},
		RetryDelayKey: []string{"60s"},
	}, envelope.Header)
}

func TestGetRetryHeaders(t *testing.T) {
	retryCount, retryDelay, err := getRetryHeaders(http.Header{})
	require.Equal(t, 0, retryCount)
	require.Equal(t, time.Duration(0), retryDelay)
	require.NoError(t, err)

	retryCount, retryDelay, err = getRetryHeaders(http.Header{
		RetryCountKey: []string{"12"},
		RetryDelayKey: []string{"60s"},
	})
	require.Equal(t, 12, retryCount)
	require.Equal(t, time.Second*60, retryDelay)
	require.NoError(t, err)
}
