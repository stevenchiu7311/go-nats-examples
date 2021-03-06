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

	var URL string = "nats://127.0.0.1:4222,nats://127.0.0.1:14222"
	nc, err := nats.Connect(URL, opts...)

	if err != nil {
		log.Println("nats connect :", err)
	}
	defer nc.Close()

	sc, err := stan.Connect("test-cluster", "at-least-once-stan-pub", stan.NatsConn(nc),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Printf("Connection lost, reason: %v\n", reason)

		}))
	if err != nil {
		log.Println("Can't connect:", err)
		fmt.Printf("CMake sure a NATS Streaming Server is running at: %s", URL)

	}

	go func() {
		var cnt = 0
		timer := time.NewTimer(1 * time.Second)
		for {
			msgStr := fmt.Sprintf("hello_%d", cnt)
			sc.Publish("topic", []byte(msgStr))
			log.Println("Pub:", msgStr)
			cnt++
			<-timer.C
			timer.Reset(1 * time.Second)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
	sc.Close()
}

/*
2020/09/30 00:17:57 Sub2: hello_0
2020/09/30 00:17:58 Sub2: hello_1
2020/09/30 00:17:59 Sub1: hello_2
2020/09/30 00:18:00 Sub2: hello_3
2020/09/30 00:18:01 Sub1: hello_4
2020/09/30 00:18:02 Sub2: hello_5
2020/09/30 00:18:03 Sub1: hello_6
*/
