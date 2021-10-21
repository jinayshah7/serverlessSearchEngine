package main

import (
	"fmt"
	"log"

	"github.com/Shopify/sarama"
)

type link struct {
	URL string
}

func main() {
	brokers := []string{"localhost:29092", "localhost:39092"}

	topic := "linksFound"
	producer, err := newProducer()
	if err != nil {
		log.Fatalln("Failed to set up Kafka connection", err)
	}
	message := prepareMessage(topic, "first message, it works")
	producer.SendMessage(message)

	consumer, err := sarama.NewConsumer(brokers, nil)
	if err != nil {
		fmt.Println("Could not create consumer: ", err)
	}
	partitionList, err := consumer.Partitions(topic)
	if err != nil {
		fmt.Println("Error retrieving partitionList ", err)
	}
	initialOffset := sarama.OffsetOldest
	for _, partition := range partitionList {
		pc, _ := consumer.ConsumePartition(topic, partition, initialOffset)
		for message := range pc.Messages() {
			fmt.Println(string(message.Value))
		}
	}
}

func prepareMessage(topic, message string) *sarama.ProducerMessage {
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: -1,
		Value:     sarama.StringEncoder(message),
	}

	return msg
}
