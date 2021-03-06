package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"
)

// Stop ack at 4th message and still can accept 5 message in sub cb function with MaxInflight(5)
func main() {
	opts := []nats.Option{nats.Timeout(10 * 60 * time.Second),
		nats.MaxReconnects(50), nats.ReconnectWait(10 * time.Second), nats.ReconnectHandler(func(_ *nats.Conn) {
			log.Println("nats client reconnected")
		})}

	var URL string = "nats://127.0.0.1:4222,nats://127.0.0.1:14222"
	nc, err := nats.Connect(URL, opts...)

	if err != nil {
		log.Println("nats connect :", err)
	}
	defer nc.Close()

	sc, err := stan.Connect("test-cluster", "pub-sub-stan-maxinflight", stan.NatsConn(nc),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Printf("Connection lost, reason: %v\n", reason)

		}))
	if err != nil {
		log.Println("Can't connect:", err)
		fmt.Printf("CMake sure a NATS Streaming Server is running at: %s", URL)

	}
	startOpt := stan.StartAt(pb.StartPosition_NewOnly)
	subAck := stan.SetManualAckMode()
	ackWait := stan.AckWait(10 * time.Second)

	go func() {
		var cnt = 0
		timer := time.NewTimer(1 * time.Second)
		for {
			sc.Publish("topic", []byte(fmt.Sprintf("hello_%d", cnt)))
			cnt++
			<-timer.C
			timer.Reset(1 * time.Second)
		}
	}()

	mcbSub := func(msg *stan.Msg) {
		log.Println("Sub:", string(msg.Data))
		counterStr := strings.Replace(string(msg.Data), "hello_", "", -1)
		counter, _ := strconv.ParseInt(counterStr, 10, 0)
		if counter < 3 {
			defer msg.Ack()
		} else {
			log.Println("Sub no ack:", string(msg.Data))
		}
	}

	var Sub stan.Subscription

	go func() {
		Sub, err = sc.QueueSubscribe("topic", "g1", mcbSub, startOpt, stan.DurableName(""), stan.MaxInflight(5), subAck, ackWait)
		if err != nil {
			log.Println("queue subscribe testTopic:", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
	Sub.Unsubscribe()
	Sub.ClearMaxPending()
	sc.Close()
}
