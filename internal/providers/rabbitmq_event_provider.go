package providers

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/config"
	appErrors "github.com/Alessandro-Massarotti-Jr/go-api/internal/errors"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// eventBusExchange is the topic exchange shared by every service.
	// Routing keys follow services.<serviceName>.<topic>.
	eventBusExchange      = "event-bus-services"
	listenerPrefetchCount = 10
)

// RabbitMQEventProvider publishes to and consumes from the event-bus-services exchange.
// Published events go out as services.<APP_NAME>.<topic>; listeners receive the topic
// from any service (services.*.<topic>) through a durable queue named <APP_NAME>.<topic>.
type RabbitMQEventProvider struct {
	connection     *amqp.Connection
	publishChannel *amqp.Channel
	publishMutex   sync.Mutex
	serviceName    string
	handlersMutex  sync.Mutex
	handlers       map[string][]EventHandler // by topic
	logger         *logger.Logger
}

var _ EventProvider = (*RabbitMQEventProvider)(nil)

var (
	rabbitMQEventProviderInstance *RabbitMQEventProvider
	rabbitMQEventProviderErr      error
	rabbitMQEventProviderOnce     sync.Once
)

// GetRabbitMQEventProviderInstance returns the shared provider, connecting on the first call
// and ensuring the event bus exchange exists.
func GetRabbitMQEventProviderInstance() (*RabbitMQEventProvider, error) {
	rabbitMQEventProviderOnce.Do(func() {
		rabbitMQEventProviderInstance, rabbitMQEventProviderErr = newRabbitMQEventProvider()
	})
	return rabbitMQEventProviderInstance, rabbitMQEventProviderErr
}

func newRabbitMQEventProvider() (*RabbitMQEventProvider, error) {
	const location = "RabbitMQEventProvider.New"
	appLogger := logger.GetInstance()
	serviceName := config.GetConfig(config.AppName)

	connection, err := amqp.Dial(config.GetConfig(config.RabbitMQURI))
	if err != nil {
		return nil, appErrors.CreateEventProviderError(appErrors.CreateEventProviderErrorInput{
			Location:    location,
			Message:     "Failed to connect to RabbitMQ",
			OriginError: err,
		})
	}

	publishChannel, err := connection.Channel()
	if err != nil {
		_ = connection.Close()
		return nil, appErrors.CreateEventProviderError(appErrors.CreateEventProviderErrorInput{
			Location:    location,
			Message:     "Failed to open RabbitMQ channel",
			OriginError: err,
		})
	}

	if err := publishChannel.ExchangeDeclare(eventBusExchange, amqp.ExchangeTopic, true, false, false, false, nil); err != nil {
		_ = connection.Close()
		return nil, appErrors.CreateEventProviderError(appErrors.CreateEventProviderErrorInput{
			Location:    location,
			Message:     "Failed to declare event bus exchange",
			OriginError: err,
		})
	}

	appLogger.Info(logger.LogMessage{
		Action:  "rabbitmq.connect",
		Message: "Connected to RabbitMQ",
		Data:    map[string]any{"exchange": eventBusExchange, "serviceName": serviceName},
	})

	return &RabbitMQEventProvider{
		connection:     connection,
		publishChannel: publishChannel,
		serviceName:    serviceName,
		handlers:       map[string][]EventHandler{},
		logger:         appLogger,
	}, nil
}

func (rabbitMQEventProvider *RabbitMQEventProvider) PublishEvent(ctx context.Context, input PublishEventInput) error {
	const location = "RabbitMQEventProvider.PublishEvent"

	payload, err := json.Marshal(input.Payload)
	if err != nil {
		return appErrors.CreateEventProviderError(appErrors.CreateEventProviderErrorInput{
			Location:    location,
			Message:     "Failed to encode event payload",
			Topic:       input.Topic,
			OriginError: err,
		})
	}

	routingKey := "services." + rabbitMQEventProvider.serviceName + "." + input.Topic

	// amqp channels are not meant to be shared by concurrent publishers.
	rabbitMQEventProvider.publishMutex.Lock()
	defer rabbitMQEventProvider.publishMutex.Unlock()

	err = rabbitMQEventProvider.publishChannel.PublishWithContext(ctx, eventBusExchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now().UTC(),
		AppId:        rabbitMQEventProvider.serviceName,
		Body:         payload,
	})
	if err != nil {
		return appErrors.CreateEventProviderError(appErrors.CreateEventProviderErrorInput{
			Location:    location,
			Message:     "Failed to publish event",
			Topic:       input.Topic,
			OriginError: err,
		})
	}

	return nil
}

func (rabbitMQEventProvider *RabbitMQEventProvider) AddListener(input AddListenerInput) {
	rabbitMQEventProvider.handlersMutex.Lock()
	defer rabbitMQEventProvider.handlersMutex.Unlock()

	rabbitMQEventProvider.handlers[input.Topic] = append(rabbitMQEventProvider.handlers[input.Topic], input.Handler)
}

func (rabbitMQEventProvider *RabbitMQEventProvider) StartListening(ctx context.Context) error {
	const location = "RabbitMQEventProvider.StartListening"

	rabbitMQEventProvider.handlersMutex.Lock()
	handlers := make(map[string][]EventHandler, len(rabbitMQEventProvider.handlers))
	for topic, topicHandlers := range rabbitMQEventProvider.handlers {
		handlers[topic] = append([]EventHandler(nil), topicHandlers...)
	}
	rabbitMQEventProvider.handlersMutex.Unlock()

	consumeChannel, err := rabbitMQEventProvider.connection.Channel()
	if err != nil {
		return appErrors.CreateEventProviderError(appErrors.CreateEventProviderErrorInput{
			Location:    location,
			Message:     "Failed to open RabbitMQ channel",
			OriginError: err,
		})
	}
	defer consumeChannel.Close()

	if err := consumeChannel.Qos(listenerPrefetchCount, 0, false); err != nil {
		return appErrors.CreateEventProviderError(appErrors.CreateEventProviderErrorInput{
			Location:    location,
			Message:     "Failed to set RabbitMQ prefetch",
			OriginError: err,
		})
	}

	var consumers sync.WaitGroup
	consumerTags := make([]string, 0, len(handlers))
	for topic, topicHandlers := range handlers {
		queueName := rabbitMQEventProvider.serviceName + "." + topic
		bindingKey := "services.*." + topic

		if _, err := consumeChannel.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
			return appErrors.CreateEventProviderError(appErrors.CreateEventProviderErrorInput{
				Location:    location,
				Message:     "Failed to declare queue",
				Topic:       topic,
				OriginError: err,
			})
		}
		if err := consumeChannel.QueueBind(queueName, bindingKey, eventBusExchange, false, nil); err != nil {
			return appErrors.CreateEventProviderError(appErrors.CreateEventProviderErrorInput{
				Location:    location,
				Message:     "Failed to bind queue",
				Topic:       topic,
				OriginError: err,
			})
		}

		consumerTag := queueName
		deliveries, err := consumeChannel.Consume(queueName, consumerTag, false, false, false, false, nil)
		if err != nil {
			return appErrors.CreateEventProviderError(appErrors.CreateEventProviderErrorInput{
				Location:    location,
				Message:     "Failed to consume queue",
				Topic:       topic,
				OriginError: err,
			})
		}
		consumerTags = append(consumerTags, consumerTag)

		rabbitMQEventProvider.logger.Info(logger.LogMessage{
			Action:  "rabbitmq.listen",
			Message: "Listening to topic",
			Data:    map[string]any{"queue": queueName, "bindingKey": bindingKey, "handlers": len(topicHandlers)},
		})

		consumers.Go(func() {
			// In-flight events finish even when ctx is cancelled during shutdown.
			rabbitMQEventProvider.consume(context.WithoutCancel(ctx), consumeInput{
				Topic:      topic,
				Handlers:   topicHandlers,
				Deliveries: deliveries,
			})
		})
	}

	connectionClosed := rabbitMQEventProvider.connection.NotifyClose(make(chan *amqp.Error, 1))

	select {
	case <-ctx.Done():
		// Stop receiving new deliveries, then wait for the in-flight ones to be acked.
		for _, consumerTag := range consumerTags {
			_ = consumeChannel.Cancel(consumerTag, false)
		}
		consumers.Wait()
		return nil
	case closeErr := <-connectionClosed:
		consumers.Wait()
		return appErrors.CreateEventProviderError(appErrors.CreateEventProviderErrorInput{
			Location:    location,
			Message:     "RabbitMQ connection closed",
			OriginError: closeErr,
		})
	}
}

type consumeInput struct {
	Topic      string
	Handlers   []EventHandler
	Deliveries <-chan amqp.Delivery
}

// consume runs every handler of the topic for each delivery. A failed event is requeued once;
// if it fails again on redelivery it is discarded so a poison message does not loop forever.
func (rabbitMQEventProvider *RabbitMQEventProvider) consume(ctx context.Context, input consumeInput) {
	for delivery := range input.Deliveries {
		event := Event{Topic: input.Topic, Payload: delivery.Body}

		var handlerErr error
		for _, handler := range input.Handlers {
			if handlerErr = handler(ctx, event); handlerErr != nil {
				break
			}
		}

		if handlerErr == nil {
			_ = delivery.Ack(false)
			continue
		}

		requeue := !delivery.Redelivered
		rabbitMQEventProvider.logger.Error(logger.LogMessage{
			Action:  "rabbitmq.consume",
			Message: "Failed to handle event",
			Data: map[string]any{
				"topic":      input.Topic,
				"routingKey": delivery.RoutingKey,
				"requeued":   requeue,
				"error":      handlerErr.Error(),
			},
		})
		_ = delivery.Nack(false, requeue)
	}
}
