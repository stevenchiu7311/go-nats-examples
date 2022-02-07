package main

import (
	"encoding/json"
	"github.com/nats-io/go-nats-examples/examples/order/model"
	"github.com/nats-io/nats.go"
	"log"
	"strconv"
	"time"
)

const (
	streamName     = "ORDERS"
	streamSubjects = "ORDERS.*"
	subjectName    = "ORDERS.created"
)

func main() {
	// Connect to NATS
	nc, _ := nats.Connect(nats.DefaultURL)
	// Creates JetStreamContext
	js, err := nc.JetStream()
	js.PurgeStream(streamName)
	checkErr(err)
	// Creates stream
	err = createStream(js)
	checkErr(err)
	// Create orders by publishing messages
	err = createOrder(js)
	checkErr(err)
}

// createOrder publishes stream of events
// with subject "ORDERS.created"
func createOrder(js nats.JetStreamContext) error {
	var order model.Order
	for i := 1; i <= 10; i++ {
		order = model.Order{
			OrderID:    i,
			CustomerID: "Cust-" + strconv.Itoa(i),
			Status:     "created",
		}
		orderJSON, _ := json.Marshal(order)
		_, err := js.Publish(subjectName, orderJSON)
		if err != nil {
			return err
		}
		log.Printf("Order with OrderID:%d has been published\n", i)
	}
	return nil
}

// createStream creates a stream by using JetStreamContext
func createStream(js nats.JetStreamContext) error {
	// Check if the ORDERS stream already exists; if not, create it.
	stream, err := js.StreamInfo(streamName)
	if err != nil {
		log.Println(err)
	}
	if stream == nil {
		log.Printf("creating stream %q and subjects %q", streamName, streamSubjects)
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{streamSubjects},
			MaxAge:   1 * time.Second,
		})
		if err != nil {
			return err
		}
	} else {
		_, err = js.UpdateStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{streamSubjects},
			MaxAge:   1 * time.Second,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
