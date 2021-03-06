package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"
)

// After receive function work 5s, mock all subs disconnection and cannot receive/ack about 7s
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

	sc, err := stan.Connect("test-cluster", "queue-group", stan.NatsConn(nc),
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

	mcbSub1 := func(msg *stan.Msg) {
		log.Println("Sub1:", string(msg.Data))
		defer msg.Ack()
	}
	mcbSub2 := func(msg *stan.Msg) {
		log.Println("Sub2:", string(msg.Data))
		defer msg.Ack()
	}
	var sub1 stan.Subscription
	var sub2 stan.Subscription
	go func() {
		sub1, err = sc.QueueSubscribe("topic", "g1", mcbSub1, startOpt, stan.DurableName("durable"), stan.MaxInflight(1), subAck, ackWait)
		if err != nil {
			log.Println("queue subscribe Topic:", err)
		}
	}()

	go func() {
		sub2, err = sc.QueueSubscribe("topic", "g1", mcbSub2, startOpt, stan.DurableName("durable"), stan.MaxInflight(10), subAck, ackWait)
		if err != nil {
			log.Println("queue subscribe testTopic:", err)
		}
	}()

	time.Sleep(time.Second * 5)

	sub1.Close()
	sub2.Close()
	log.Println("After receive function work 5s, mock all subs disconnection and cannot receive/ack about 7s")
	time.Sleep(time.Second * 7)
	log.Println("Resume resume the connection")

	go func() {
		sub1, err = sc.QueueSubscribe("topic", "g1", mcbSub1, startOpt, stan.DurableName("durable"), stan.MaxInflight(1), subAck, ackWait)
		if err != nil {
			log.Println("queue subscribe Topic:", err)
		}
	}()

	go func() {
		sub2, err = sc.QueueSubscribe("topic", "g1", mcbSub2, startOpt, stan.DurableName("durable"), stan.MaxInflight(10), subAck, ackWait)
		if err != nil {
			log.Println("queue subscribe testTopic:", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
	sub1.Close()
	sub2.Unsubscribe()
	sub1.Close()
	sub2.Unsubscribe()
	sc.Close()
}
