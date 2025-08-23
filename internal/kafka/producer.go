package kafka

import (
	"context"
	"demoserv/internal/config"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
)

func NewProducer(cfg *config.Config)  {
	ctx := context.Background()

	writer := kafka.NewWriter(kafka.WriterConfig{
    Brokers: []string{cfg.Kafka.KAFKA_BROKER},
    Topic:   cfg.Kafka.KAFKA_TOPIC,
	})

	defer writer.Close()

	
	jsonData, err := os.ReadFile("test2.json")
	if err != nil {
		log.Fatal("Ошибка при чтении файла model.json:", err)
	}

	
	err = writer.WriteMessages(ctx, kafka.Message{
		Value: jsonData,
	})

	if err != nil {
		log.Fatal("Ошибка при отправке:", err)
	}
}
