package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	nats "github.com/nats-io/nats.go"
)

//Sub 1 leave in queue group, and sub2 will take over all of subscription messages
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
			log.Println("send:", fmt.Sprintf("hello_Steven_%d", cnt))
			nc.Publish("testTopic.Steven", []byte(fmt.Sprintf("hello_Steven_%d", cnt)))
			cnt++
			timer.Reset(1 * time.Second)
		}
	}()

	mcbSteven1 := func(msg *nats.Msg) {
		log.Println("receive Steven sub1:", string(msg.Data))
	}
	mcbSteven2 := func(msg *nats.Msg) {
		log.Println("receive Steven sub2:", string(msg.Data))
	}
	var Sub1Cb *nats.Subscription
	var Sub2Cb *nats.Subscription
	go func() {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		Sub1Cb, err = nc.QueueSubscribe("testTopic.Steven", "queue", mcbSteven1)
		if err != nil {
			log.Println("queue subscribe testTopic.Steven:", err)
		}
		<-ctx.Done()
		log.Println("Sub1Cb leave")
		Sub1Cb.Unsubscribe()
	}()

	go func() {
		Sub2Cb, err = nc.QueueSubscribe("testTopic.Steven", "queue", mcbSteven2)
		if err != nil {
			log.Println("queue subscribe testTopic.Steven:", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
	Sub1Cb.Unsubscribe()
	Sub2Cb.Unsubscribe()
}

/*
2020/09/29 23:39:45 send: hello_Steven_0
2020/09/29 23:39:45 Any: hello_Steven_0
2020/09/29 23:39:45 Steven: hello_Steven_0
2020/09/29 23:39:47 send: hello_any_0
2020/09/29 23:39:47 Any: hello_any_0
2020/09/29 23:39:48 send: hello_Steven_1
2020/09/29 23:39:48 Steven: hello_Steven_1
2020/09/29 23:39:48 Any: hello_Steven_1
*/
