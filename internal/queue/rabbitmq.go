package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// RabbitMQClient клиент для RabbitMQ
type RabbitMQClient struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	config  *RabbitMQConfig
}

// RabbitMQConfig конфигурация RabbitMQ
type RabbitMQConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	VHost    string
}

// QueueNames константы для имен очередей
const (
	QueueFramesToDetect    = "frames.to.detect"
	QueueFramesDLQ         = "frames.to.detect.dlq"
	QueueEventsNotify      = "events.notifications"
	QueueThumbnails        = "thumbnails.to.generate"
	
	ExchangeFrames         = "frames"
	ExchangeEvents         = "events"
	
	RoutingKeyDetect       = "detect"
	RoutingKeyNotify       = "notify"
)

// NewRabbitMQClient создает новый RabbitMQ клиент
func NewRabbitMQClient(cfg *RabbitMQConfig) (*RabbitMQClient, error) {
	// Формируем URL подключения
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.VHost,
	)

	log.Printf("🐰 Connecting to RabbitMQ at %s:%d...", cfg.Host, cfg.Port)

	// Подключаемся с retry
	var conn *amqp091.Connection
	var err error
	
	for i := 0; i < 5; i++ {
		conn, err = amqp091.Dial(url)
		if err == nil {
			break
		}
		
		log.Printf("⚠️  RabbitMQ connection attempt %d/5 failed: %v", i+1, err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after 5 attempts: %w", err)
	}

	// Открываем канал
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Устанавливаем QoS (Quality of Service)
	err = channel.Qos(
		10,    // prefetch count - максимум 10 сообщений на worker
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	client := &RabbitMQClient{
		conn:    conn,
		channel: channel,
		config:  cfg,
	}

	// Создаем exchanges и queues
	if err := client.setupTopology(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to setup topology: %w", err)
	}

	log.Println("✅ Connected to RabbitMQ successfully")
	return client, nil
}

// setupTopology создает exchanges, queues и bindings
func (r *RabbitMQClient) setupTopology() error {
	log.Println("🔧 Setting up RabbitMQ topology...")

	// 1. Создаем Exchanges
	exchanges := []struct {
		name string
		kind string
	}{
		{ExchangeFrames, "topic"},
		{ExchangeEvents, "topic"},
	}

	for _, ex := range exchanges {
		err := r.channel.ExchangeDeclare(
			ex.name, // name
			ex.kind, // type
			true,    // durable
			false,   // auto-deleted
			false,   // internal
			false,   // no-wait
			nil,     // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", ex.name, err)
		}
		log.Printf("✅ Exchange declared: %s", ex.name)
	}

	// 2. Создаем Dead Letter Queue
	_, err := r.channel.QueueDeclare(
		QueueFramesDLQ, // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}
	log.Printf("✅ Queue declared: %s", QueueFramesDLQ)

	// 3. Создаем основную очередь с DLQ
	args := amqp091.Table{
		"x-dead-letter-exchange":    ExchangeFrames,
		"x-dead-letter-routing-key": "dlq",
		"x-message-ttl":             300000, // 5 минут
		"x-max-length":              10000,  // максимум 10k сообщений
	}

	_, err = r.channel.QueueDeclare(
		QueueFramesToDetect, // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		args,                // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	log.Printf("✅ Queue declared: %s", QueueFramesToDetect)

	// 4. Bind queue to exchange
	err = r.channel.QueueBind(
		QueueFramesToDetect, // queue name
		RoutingKeyDetect,    // routing key
		ExchangeFrames,      // exchange
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// 5. Bind DLQ
	err = r.channel.QueueBind(
		QueueFramesDLQ,  // queue name
		"dlq",           // routing key
		ExchangeFrames,  // exchange
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind DLQ: %w", err)
	}

	// 6. Создаем очередь уведомлений
	_, err = r.channel.QueueDeclare(
		QueueEventsNotify, // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare events queue: %w", err)
	}
	log.Printf("✅ Queue declared: %s", QueueEventsNotify)

	err = r.channel.QueueBind(
		QueueEventsNotify, // queue name
		RoutingKeyNotify,  // routing key
		ExchangeEvents,    // exchange
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind events queue: %w", err)
	}

	log.Println("✅ RabbitMQ topology setup complete")
	return nil
}

// Publish публикует сообщение в очередь
func (r *RabbitMQClient) Publish(exchange, routingKey string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = r.channel.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent, // сообщение переживет перезапуск RabbitMQ
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Consume подписывается на очередь и обрабатывает сообщения
func (r *RabbitMQClient) Consume(queueName string, handler func([]byte) error) error {
	msgs, err := r.channel.Consume(
		queueName, // queue
		"",        // consumer tag
		false,     // auto-ack (делаем вручную)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("🐰 Started consuming from queue: %s", queueName)

	// Обрабатываем сообщения
	go func() {
		for msg := range msgs {
			err := handler(msg.Body)

			if err != nil {
				log.Printf("❌ Error processing message: %v", err)
				
				// Проверяем количество попыток
				retryCount := 0
				if msg.Headers != nil {
					if count, ok := msg.Headers["x-retry-count"].(int32); ok {
						retryCount = int(count)
					}
				}

				// Если меньше 3 попыток - отправляем обратно в очередь
				if retryCount < 3 {
					msg.Nack(false, true) // requeue
					log.Printf("⚠️  Message requeued (attempt %d/3)", retryCount+1)
				} else {
					// После 3 попыток - отправляем в DLQ
					msg.Nack(false, false) // не requeue, пойдет в DLQ
					log.Printf("💀 Message sent to DLQ after 3 failed attempts")
				}
			} else {
				msg.Ack(false) // успешно обработано
			}
		}
	}()

	return nil
}

// Close закрывает соединение
func (r *RabbitMQClient) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			log.Printf("Error closing channel: %v", err)
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}
	log.Println("🐰 RabbitMQ connection closed")
	return nil
}

// GetQueueStats возвращает статистику очереди
func (r *RabbitMQClient) GetQueueStats(queueName string) (int, error) {
	queue, err := r.channel.QueueInspect(queueName)
	if err != nil {
		return 0, err
	}
	return queue.Messages, nil
}