package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	// 連線到nats的伺服器
	conn, err := nats.Connect("nats://127.0.0.1:4222")
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

	// 初始化JetStream功能
	js, err := conn.JetStream()
	if err != nil {
		log.Panic(err)
	}

	// 判斷Stream是否存在，如果不存在，那麼需要建立這個Stream，否則會導致pub/sub失敗
	stream, err := js.StreamInfo(streamName)
	if err != nil {
		log.Println(err) // 如果不存在，這裡會有報錯
	}
	if stream == nil {
		log.Printf("creating stream %q and subject %q", streamName, subject)
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{subject},
			MaxAge:   3 * 24 * time.Hour,
		})
		if err != nil {
			log.Panicln(err)
		}
	}

	sub := subscribeResponse(js)
	defer func() {
		if err := sub.Unsubscribe(); err != nil {
			log.Panicln(err)
		}
	}()

	// 傳送訊息
	future, _ := js.PublishAsync(subject, []byte("Hello World! "+time.Now().Format(time.RFC3339)))

	select {
	case <-future.Ok():
		log.Println("Publish process ok")
	case <-time.After(3 * time.Second):
		log.Fatalln("Did not receive completion signal")
	}
	time.Sleep(2 * time.Second)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
}

func subscribeResponse(js nats.JetStreamContext) *nats.Subscription {
	log.Printf("subscribe")
	cbHandle := func(m *nats.Msg) {
		if len(m.Data) == 0 {
			return
		}
		log.Printf("m.Data: %s", m.Data)
		m.Respond([]byte(fmt.Sprintf("received %s", m.Data)))
	}

	sub, err := js.Subscribe(subject, cbHandle, nats.AckAll(), nats.DeliverLast())
	if err != nil {
		log.Panic(err)
	}
	return sub
}