package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	nats "github.com/nats-io/nats.go"
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

	go func() {
		var cnt = 0
		timer := time.NewTimer(1 * time.Second)
		for {
			<-timer.C
			message, err := nc.Request("testTopic.any", []byte(fmt.Sprintf("hello_any_%d", cnt)), 1*time.Second)
			if err != nil {
				log.Println("send:", fmt.Sprintf("hello_any_%d", cnt), " get error caused by reply timeout", err)
			} else {
				log.Println("send:", fmt.Sprintf("hello_any_%d", cnt), "reply receive:", string(message.Data))
			}
			cnt++
			timer.Reset(1 * time.Second)
		}
	}()

	mcbAny := func(msg *nats.Msg) {
		log.Println("receive Any:", string(msg.Data))
		msg.Respond([]byte("ACK: " + "receive Any:" + string(msg.Data) ))
	}
	var Sub1Cb *nats.Subscription
	go func() {
		time.Sleep(5 * time.Second)
		Sub1Cb, err = nc.Subscribe("testTopic.*", mcbAny)
		if err != nil {
			log.Println("queue subscribe testTopic.*:", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
	Sub1Cb.Unsubscribe()
}
