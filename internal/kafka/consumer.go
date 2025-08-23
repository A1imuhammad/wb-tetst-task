package kafka

import (
	"demoserv/internal/cache"
	"demoserv/internal/config"
	"demoserv/internal/models"
	"demoserv/internal/postgress"
	"demoserv/internal/validate"
	
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

func NewConsumer(сtx context.Context, cfg *config.Config, pool *pgxpool.Pool, cache *cache.Cache) {
	// Подключаемся к брокеру
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.Kafka.KAFKA_BROKER},
		Topic:   cfg.Kafka.KAFKA_TOPIC,
		GroupID: cfg.Kafka.KAFKA_GROUP,
	})

	defer reader.Close()


	log.Println("listening topic...")

	// Читаем очередь
	for {
		msg, err := reader.ReadMessage(сtx)
		if err != nil {
			log.Printf("unable to read message: %v", err)
			continue
		}
		fmt.Println("message received")

		// Анмаршалим сообщение
		var order models.Order
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			log.Printf("unable to unmarshal message: %v", err)
			continue
		}

		// Проверяем валидность каждого поля заказа
		if err := validate.ValidateOrder(order); err != nil {
			log.Printf("invalid order data: %v (order_uid: %s)", err, order.OrderUID)
			continue
		}

		// Вставка в базу
		if err := postgress.InsertOrder(сtx, pool, &order); err != nil {
			log.Printf("unable to insert order: %v", err)
			continue
		}

		cache.Add(order) // добавляем в кэш

		if err := reader.CommitMessages(сtx, msg); err != nil {
			log.Printf("unable to commit message: %v", err)
		}

		fmt.Printf("Processed order: %s\n", order.OrderUID)
	}
}
