package main

import (
	"fmt"
	"log"

	"github.com/Shopify/sarama"
)

type MessageRepository interface {
	SendMessage(topic string, message []byte) error
	Messages(topic string) <-chan []byte
}

type messageRepository struct {
	producer sarama.AsyncProducer
	consumer sarama.Consumer
}

func (mr *messageRepository) SendMessage(topic string, message []byte) error {

	return nil
}

func (mr *messageRepository) Messages(topic string) <-chan []byte {

	return nil
}
func NewMessageRepository(brokers []string) MessageRepository {

	producer, err := newProducer()
	if err != nil {
		log.Fatalln("Failed to set up Kafka connection", err)
	}

	consumer, err := sarama.NewConsumer(brokers, nil)
	if err != nil {
		fmt.Println("Could not create consumer: ", err)
	}

	return &messageRepository{
		producer: producer,
		consumer: consumer,
	}
}

func newProducer() (sarama.AsyncProducer, error) {
	brokers := []string{"localhost:29092", "localhost:39092"}
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	producer, err := sarama.NewAsyncProducer(brokers, config)
	return producer, err
}
