package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func consumer(ctx context.Context, conn *amqp.Connection, cname string, exname string, qname string, rkey string) {
	// defer wg.Done()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exname,  // name
		"topic", // type
		false,   // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		qname, // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name, // queue name
		rkey,   // routing key
		exname, // exchange name
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue name
		cname,  // consumer name
		true,   // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	for d := range msgs {
		log.Printf(" [%s] %s", cname, d.Body)
	}
	<-ctx.Done()

	// var forever chan struct{}
	// for d := range msgs {
	// 	log.Printf(" [%s][%s] %s", cname, rkey, d.Body)
	// }
	// <-forever

	log.Printf(" [%s] out", cname)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")

	go consumer(ctx, conn, "consumer(1)", "exchg.test", "q.1", "a.*.*")
	go consumer(ctx, conn, "consumer(2)", "exchg.test", "q.2", "*.1.*")
	go consumer(ctx, conn, "consumer(3)", "exchg.test", "q.3", "*.*.@")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	cancel()
	fmt.Println("Bye")
}
