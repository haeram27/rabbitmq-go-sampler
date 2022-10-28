# rabbitmq-go-sampler
run rabbitmq broker
$ docker run -d --rm \
-p 5672:5672 -p 15672:15672 \
--hostname rabbitmq \
--name rabbitmq \
rabbitmq:3.10-management


rabbitmq logs
$ docker logs -f rabbitmq


rabbitmq management GUI
http://localhost:15672/

$ go run ./topic_publisher.go
$ go run ./topic_subscriber.go

# tutorials
https://github.com/rabbitmq/rabbitmq-tutorials/tree/main/go
