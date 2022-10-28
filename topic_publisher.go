package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func bodyFrom(s string) string {
	if len(s) <= 0 {
		s = "hello"
	}
	return s
}

func severityFrom(s string) string {
	if len(s) <= 0 {
		s = "anonymous.info"
	}

	return s
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	exname := "exchg.test"

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	klv1 := []string{"a", "b", "c"}
	klv2 := []string{"1", "2", "3"}
	klv3 := []string{"!", "@", "#"}
	var count atomic.Uint64

	producer := func(name string) {
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

		for {
			for i := 0; i < len(klv1); i++ {
				for j := 0; j < len(klv2); j++ {
					for k := 0; k < len(klv3); k++ {
						select {
						case <-ctx.Done():
							log.Printf("[%s] out", name)
							return
						default:
						}

						rkey := klv1[i] + "." + klv2[j] + "." + klv3[k]
						msgNum := strconv.FormatUint(count.Add(1), 10)
						msg := msgNum + " -> " + rkey
						err := ch.PublishWithContext(ctx,
							exname,             // exchange
							severityFrom(rkey), // routing key
							false,              // mandatory
							false,              // immediate
							amqp.Publishing{
								ContentType: "text/plain",
								Body:        []byte(bodyFrom(msg)),
							})
						failOnError(err, "Failed to publish a message")
						log.Printf("[%s] %s -> %s", name, msgNum, rkey)
					}
				}
			}
		}
	}

	go producer("producer(1)")
	go producer("producer(2)")
	go producer("producer(3)")

	fmt.Println(" [*] Waiting for logs. To exit press CTRL+C")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	cancel()
	fmt.Println("Bye")
}
