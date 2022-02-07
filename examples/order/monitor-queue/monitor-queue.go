package main

import (
	"encoding/json"
	"github.com/nats-io/go-nats-examples/examples/order/model"
	"log"
	"runtime"

	"github.com/nats-io/nats.go"
)

func main() {
	// Connect to NATS
	nc, _ := nats.Connect(nats.DefaultURL)
	js, err := nc.JetStream()
	if err != nil {
		log.Fatal(err)
	}

	queueSubscribe(js, "monitor3", "queue1")
	queueSubscribe(js, "monitor4", "queue1")
	runtime.Goexit()

}

func subscribe(js nats.JetStreamContext, consumerName string) {
	// Create durable consumer monitor
	js.Subscribe("ORDERS.*", func(msg *nats.Msg) {
		msg.Ack()
		var order model.Order
		err := json.Unmarshal(msg.Data, &order)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("durable[%s] service subscribes from subject:%s\n", consumerName, msg.Subject)
		log.Printf("durable[%s] OrderID:%d, CustomerID: %s, Status:%s\n", consumerName, order.OrderID, order.CustomerID, order.Status)
	}, nats.Durable(consumerName), nats.ManualAck())
}

func queueSubscribe(js nats.JetStreamContext, consumerName string, queueName string) {
	// Create durable consumer monitor
	js.QueueSubscribe("ORDERS.*", queueName, func(msg *nats.Msg) {
		msg.Ack()
		var order model.Order
		err := json.Unmarshal(msg.Data, &order)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("subscriber[%s] queue[%s] service subscribes from subject:%s\n", consumerName, queueName, msg.Subject)
		log.Printf("subscriber[%s] queue[%s] OrderID:%d, CustomerID: %s, Status:%s\n", consumerName, queueName, order.OrderID, order.CustomerID, order.Status)
	}, nats.ManualAck())
}