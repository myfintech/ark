package kafka

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/pkg/errors"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
)

type Broker struct {
	Producer       *kafka.Producer
	Consumer       *kafka.Consumer
	PublishTimeout *time.Duration
	Logger         logz.FieldLogger
}

const RetryCountKey = "X-Retry-Count"
const RetryDelayKey = "X-Retry-Delay"

func (b Broker) Publish(topic cqrs.RouteKey, messages ...cqrs.Message) error {
	delivered := make(chan kafka.Event)
	for _, message := range messages {
		v, err := json.Marshal(message)
		if err != nil {
			return err
		}

		if err = b.Producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     topic.StringPtr(),
				Partition: kafka.PartitionAny,
			},
			Value:     v,
			Headers:   toKafkaHeaders(message),
			Timestamp: message.Time(),
		}, delivered); err != nil {
			return err
		}
		select {
		case <-timeoutWithDuration(b.PublishTimeout):
			return cqrs.PublishTimeoutErr
		case e := <-delivered:
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					return ev.TopicPartition.Error
				}
			default:
				return errors.Errorf("received unknown message(%s) type(%v), while waiting for delivery confirmation", e, e)
			}
		}
	}
	return nil
}

var defaultTimeout = time.Second * 30

func timeoutWithDuration(timeout *time.Duration) <-chan time.Time {
	if timeout == nil {
		timeout = &defaultTimeout
	}
	return time.After(*timeout)
}

func fromKafkaHeaders(headers []kafka.Header) cqrs.EnvelopeOption {
	return func(envelope *cqrs.Envelope) error {
		for _, h := range headers {
			envelope.Header.Set(h.Key, string(h.Value))
		}
		return nil
	}
}

func toKafkaHeaders(message cqrs.Message) (kHeaders []kafka.Header) {
	env, _ := message.(cqrs.Envelope)
	for k := range env.Header {
		kHeaders = append(kHeaders, kafka.Header{
			Key:   k,
			Value: []byte(env.Header.Get(k)),
		})
	}
	return
}

func getRetryHeaders(headers http.Header) (retryCount int, retryDelay time.Duration, err error) {
	retryCountStr := headers.Get(RetryCountKey)
	retryDelaySecondsStr := headers.Get(RetryDelayKey)
	if retryCountStr == "" || retryDelaySecondsStr == "" {
		return
	}

	retryCount, err = strconv.Atoi(retryCountStr)
	if err != nil {
		return
	}

	retryDelay, err = time.ParseDuration(retryDelaySecondsStr)
	if err != nil {
		return
	}

	return
}

func (b Broker) Subscribe(ctx context.Context, topic cqrs.RouteKey, ackDeadline *time.Duration) (<-chan cqrs.Envelope, error) {
	sub := make(chan cqrs.Envelope)

	if err := b.Consumer.Subscribe(topic.String(), nil); err != nil {
		close(sub)
		return sub, err
	}

	go func() {
		defer func() {
			if rErr := recover(); rErr != nil {
				b.Logger.Errorf("subscriber closed because of a panic %v", rErr)
			}
			_ = b.Close()
			close(sub)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:

			}
			envelope := cqrs.NewEnvelope()
			msg, err := b.Consumer.ReadMessage(time.Second * 30)
			if err != nil {
				if kErr, ok := err.(kafka.Error); ok && kErr.Code() == kafka.ErrTimedOut {
					continue
				}
				envelope.Error = err
				sub <- envelope
				continue
			}

			envelope = cqrs.NewEnvelope(cqrs.FromData(msg.Value), fromKafkaHeaders(msg.Headers))

			retryCount, retryDelay, err := getRetryHeaders(envelope.Header)
			if err != nil {
				b.Logger.Errorf("there was an error parsing the message headers %v", err)
				continue
			}

			if retryDelay > 0 {
				if retryCount > 0 {
					retryDelay *= time.Duration(retryCount)
				}
				retryCount++

				if retryDelay > time.Minute*3 {
					retryDelay = time.Minute * 3
				}

				b.Logger.Errorf("message is in retry(%d), sleeping for %s %s", retryCount, retryDelay, string(msg.Value))
				time.Sleep(retryDelay)
			}

			if err = envelope.Validate(); err != nil {
				envelope.Error = err
			}
			sub <- envelope
			if ackDeadline == nil {
				_, _ = b.Consumer.CommitMessage(msg)
				continue
			}
			select {
			case <-time.After(*ackDeadline):
				envelope.Header.Set(RetryCountKey, strconv.Itoa(retryCount))
				envelope.Header.Set(RetryDelayKey, "60s")
				_ = b.Publish(topic, envelope)
			case err = <-envelope.Wait():
				if err != nil {
					envelope.Header.Set(RetryCountKey, strconv.Itoa(retryCount))
					envelope.Header.Set(RetryDelayKey, "60s")
					_ = b.Publish(topic, envelope)
				}
			}
			_, _ = b.Consumer.CommitMessage(msg)
		}
	}()
	return sub, nil
}

func (b Broker) Close() error {
	b.Producer.Flush(int(time.Second * 15))
	b.Producer.Close()
	_ = b.Consumer.Close()
	return nil
}
