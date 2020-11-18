package services

import (
	kafka "github.com/segmentio/kafka-go"

	"context"
	"os"
	"strings"
)

type KafkaService interface {
	Send(string, string) error
	Read() (string, error)
}

type kafkaService struct {
	writer *kafka.Writer
	reader *kafka.Reader
}

func NewKafkaWriterService() KafkaService {
	kafkaURL := os.Getenv("KAFKA_URL")
	topic := os.Getenv("KAFKA_TOPIC")
	kafka_writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{kafkaURL},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	return &kafkaService{
		writer: kafka_writer,
		reader: nil,
	}
}

func NewKafkaReaderService() KafkaService {
	kafkaURL := os.Getenv("KAFKA_URL")
	topic := os.Getenv("KAFKA_TOPIC")
	group_id := os.Getenv("GROUP_ID")
	brokers := strings.Split(kafkaURL, ",")
	kafka_reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  group_id,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	return &kafkaService{
		writer: nil,
		reader: kafka_reader,
	}
}

func (kafkaClient *kafkaService) Send(key, message string) error {

	msg := kafka.Message{
		Key:   []byte(key),
		Value: []byte(message),
	}
	err := kafkaClient.writer.WriteMessages(context.Background(), msg)

	if err != nil {
		return err
	} else {
		return nil
	}

}

func (kafkaClient *kafkaService) Read() (string, error) {

	m, err := kafkaClient.reader.ReadMessage(context.Background())
	if err != nil {
		return "", err
	}

	return string(m.Value), nil

}
