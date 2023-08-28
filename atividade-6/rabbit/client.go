package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	queueName := "tire_quality"
	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	quality := rand.Float64() * 100.0

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
		log.Fatalf("Failed to register a consumer: %v", err)
	}
	rttTimes := make([]time.Duration, 10001)
	for i := 0; i < 10000; i++ {
		sendTime := time.Now()

		message := fmt.Sprintf("%.2f", quality)
		err := ch.PublishWithContext(
			ctx,
			"",
			queueName,
			false,
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(message),
			})
		if err != nil {
			log.Fatalf("Failed to publish a message: %v", err)
		}
		//fmt.Printf("Sent tire quality: %s\n", message)

		msg := <-msgs // Receive the calculated quality from the server

		calculatedQuality, err := strconv.ParseFloat(string(msg.Body), 2)
		if err != nil {
			fmt.Println("Error converting calculated quality:", err)
			continue
		}

		if calculatedQuality < 20 {
			quality = 100.0
		} else {
			quality = float64(calculatedQuality)
		}

		receiveTime := time.Now()
		rtt := receiveTime.Sub(sendTime)
		rttTimes[i] = rtt

		//fmt.Printf("Received calculated quality from server: %.2f\n", quality)
	}
	var totalRTT time.Duration
	for _, rtt := range rttTimes {
		totalRTT += rtt
	}
	averageRTT := totalRTT / time.Duration(10001)

	// Imprime a média dos tempos de RTT
	fmt.Println("Tempo médio de RTT:", averageRTT)
	// Simulate some time passing
	//time.Sleep(time.Second)
}
