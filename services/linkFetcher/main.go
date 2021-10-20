package main

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Shopify/sarama"
)

func main() {
	//brokers := os.Getenv("KAFKA_BROKERS")

	err := connectKafka()
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
	}
}

func connectKafka() error {

	brokers := "localhost:29092,localhost:39092"
	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)

	brokerList := strings.Split(brokers, ",")
	log.Printf("Kafka brokers: %s", strings.Join(brokerList, ", "))

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	producer, err := sarama.NewAsyncProducer(brokerList, config)

	if err != nil {
		return errors.New("could not start Sarama client")
	}

	go func() {
		for err := range producer.Errors() {
			log.Println("Failed to write access log entry:", err)
		}
	}()
}
