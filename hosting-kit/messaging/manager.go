package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"hosting-kit/otel"
	"log"
	"sync"
	"time"

	"github.com/wagslane/go-rabbitmq"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type MessageHandler func(ctx context.Context, body []byte, routingKey string) error

type MessageManager struct {
	conn           *rabbitmq.Conn
	consumers      []*rabbitmq.Consumer
	publisher      *rabbitmq.Publisher
	wg             sync.WaitGroup
	handlerTimeout time.Duration
	tracer         trace.Tracer
}

func NewMessageManager(url string, exchanges []ExchangeConfig, handlerTimeout time.Duration, tracer trace.Tracer) (*MessageManager, error) {
	conn, err := rabbitmq.NewConn(url, rabbitmq.WithConnectionOptionsLogging)
	if err != nil {
		return nil, err
	}

	for _, ex := range exchanges {
		declarer, err := rabbitmq.NewPublisher(
			conn,
			rabbitmq.WithPublisherOptionsExchangeName(ex.Name),
			rabbitmq.WithPublisherOptionsExchangeKind(string(ex.Type)),
			rabbitmq.WithPublisherOptionsExchangeDeclare,
			rabbitmq.WithPublisherOptionsExchangeDurable,
		)
		if err != nil {
			conn.Close()
			return nil, err
		}
		declarer.Close()
	}

	publisher, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsConfirm,
	)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &MessageManager{
		conn:           conn,
		consumers:      []*rabbitmq.Consumer{},
		publisher:      publisher,
		handlerTimeout: handlerTimeout,
		tracer:         tracer,
	}, nil
}

func (m *MessageManager) Subscribe(queueName, routingKey, exchangeName string, handler MessageHandler, dlq *DLQConfig) error {
	opts := []func(*rabbitmq.ConsumerOptions){
		rabbitmq.WithConsumerOptionsRoutingKey(routingKey),
		rabbitmq.WithConsumerOptionsExchangeName(exchangeName),
		rabbitmq.WithConsumerOptionsQueueDurable,
	}

	if dlq != nil {
		opts = append(opts, rabbitmq.WithConsumerOptionsQueueArgs(map[string]interface{}{
			"x-dead-letter-exchange":    dlq.ExchangeName,
			"x-dead-letter-routing-key": dlq.RoutingKey,
		}))
	}

	consumer, err := rabbitmq.NewConsumer(
		m.conn,
		queueName,
		opts...,
	)
	if err != nil {
		return err
	}

	m.consumers = append(m.consumers, consumer)

	rabbitHandler := func(d rabbitmq.Delivery) rabbitmq.Action {
		ctx := context.Background()

		ctx = ExtractTraceHeaders(ctx, d.Headers)

		ctx, span := m.tracer.Start(ctx, "rabbitmq.consume")
		defer span.End()

		ctx = otel.InjectTracing(ctx, m.tracer)

		span.SetAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.operation", "consume"),
			attribute.String("messaging.destination", exchangeName),
			attribute.String("messaging.routing_key", routingKey),
			attribute.String("messaging.rabbitmq.queue", queueName),
			attribute.Int("messaging.message.payload_size_bytes", len(d.Body)),
		)

		ctx, cancel := context.WithTimeout(ctx, m.handlerTimeout)
		defer cancel()

		err := handler(ctx, d.Body, d.RoutingKey)

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		if err == nil {
			span.SetAttributes(attribute.String("messaging.result", "ack"))
			return rabbitmq.Ack
		}

		if errors.Is(err, ErrPermanentFailure) {
			span.SetAttributes(attribute.String("messaging.result", "nack_discard"))
			return rabbitmq.NackDiscard
		}

		span.SetAttributes(attribute.String("messaging.result", "nack_requeue"))
		return rabbitmq.NackRequeue
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := consumer.Run(rabbitHandler); err != nil {
			log.Printf("Consumer stopped with error: %v", err)
		}
	}()

	return nil
}

func (m *MessageManager) Publish(ctx context.Context, exchangeName, routingKey string, data interface{}) error {
	eventBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	headers := InjectTraceHeaders(ctx)

	return m.publisher.Publish(
		eventBytes,
		[]string{routingKey},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsExchange(exchangeName),
		rabbitmq.WithPublishOptionsPersistentDelivery,
		rabbitmq.WithPublishOptionsHeaders(headers),
	)
}

func (m *MessageManager) Stop(ctx context.Context) error {
	for _, consumer := range m.consumers {
		consumer.CloseWithContext(ctx)
	}

	m.wg.Wait()

	m.publisher.Close()
	if err := m.conn.Close(); err != nil {
		return err
	}
	return nil
}
