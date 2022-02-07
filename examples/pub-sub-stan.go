package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

func main() {
	opts := []nats.Option{nats.Timeout(10 * 60 * time.Second),
		nats.MaxReconnects(50), nats.ReconnectWait(10 * time.Second), nats.ReconnectHandler(func(_ *nats.Conn) {
			log.Println("nats client reconnected")
		})}

	var URL string = "nats://127.0.0.1:4222,nats://127.0.0.1:14223"
	nc, err := nats.Connect(URL, opts...)

	if err != nil {
		log.Println("nats connect :", err)
	}
	defer nc.Close()

	sc, err := stan.Connect("test-cluster", "nathan02", stan.NatsConn(nc),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Printf("Connection lost, reason: %v\n", reason)

		}))
	if err != nil {
		log.Println("Can't connect:", err)
		fmt.Printf("CMake sure a NATS Streaming Server is running at: %s", URL)

	}

	go func() {
		timer := time.NewTimer(1 * time.Second)
		for {

			sc.Publish("new_job_info", []byte(fmt.Sprintf("Software Engineer * %d", 1)))
			<-timer.C
			timer.Reset(1 * time.Second)
		}
	}()

	mcbPeopleA := func(msg *stan.Msg) {
		log.Println("PeopleA:", string(msg.Data))

	}
	mcbPeopleB := func(msg *stan.Msg) {
		log.Println("PeopleB:", string(msg.Data))

	}
	var subPeopleA stan.Subscription
	var subPeopleB stan.Subscription
	go func() {
		subPeopleA, err = sc.QueueSubscribe("new_job_info", "PeopleA", mcbPeopleA)
		if err != nil {
			log.Println("queue subscribe new_job_info:", err)
		}
	}()

	go func() {
		subPeopleB, err = sc.QueueSubscribe("new_job_info", "PeopleB", mcbPeopleB)
		if err != nil {
			log.Println("queue subscribe new_job_info:", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
	subPeopleA.Unsubscribe()
	subPeopleB.Unsubscribe()

	subPeopleA.Close()
	subPeopleB.Close()
	sc.Close()
}
/*
2020/09/30 00:12:30 PeopleA: Software Engineer * 1
2020/09/30 00:12:30 PeopleB: Software Engineer * 1
*/