package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	queueName := "tire_quality"
	_, err = conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		fmt.Println("\nInterrupt signal received. Stopping client...")
		os.Exit(0)
	}()

	startClient(conn, queueName)
}

func resolveClient(ch *amqp.Channel, msgs <-chan amqp.Delivery, queueName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for msg := range msgs {
		qualityStr := string(msg.Body)
		fmt.Printf("Received tire quality: %s\n", qualityStr)

		qualityFloat, err := strconv.ParseFloat(strings.TrimSpace(qualityStr), 64)
		if err != nil {
			fmt.Printf("Error converting tire quality: %v\n", err)
			continue
		}

		// Calculate the new tire quality with a -3% reduction
		newQuality := int(qualityFloat * 0.97)

		// Publish the new quality back to the client
		err = ch.PublishWithContext(
			ctx,
			"",
			queueName,
			false,
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(strconv.Itoa(newQuality)),
			})
		if err != nil {
			fmt.Printf("Failed to publish a message: %v\n", err)
		}
	}
}

func startClient(conn *amqp.Connection, queueName string) {
	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Failed to open a channel: %v", err)
		return
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed to register a consumer: %v", err)
		return
	}

	fmt.Printf("Waiting for clients...\n")

	resolveClient(ch, msgs, queueName)
}
