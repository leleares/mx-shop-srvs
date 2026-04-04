package main

import (
	"context"
	"fmt"
	"os"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

type Listener struct{}

func (l *Listener) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {
	return primitive.CommitMessageState

}

func (l *Listener) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {
	return primitive.CommitMessageState
}

func main() {
	sig := make(chan os.Signal)
	p, err := rocketmq.NewTransactionProducer(&Listener{}, producer.WithNameServer([]string{"127.0.0.1:9876"}))

	if err != nil {
		panic("生成producer失败")
	}

	err = p.Start()
	if err != nil {
		fmt.Printf("start producer error: %s", err.Error())
		os.Exit(1)
	}
	topic := "transTopic"

	msg := &primitive.Message{
		Topic: topic,
		Body:  []byte("this is a tx msg"),
	}
	res, err := p.SendMessageInTransaction(context.Background(), msg)

	if err != nil {
		fmt.Printf("send message error: %s\n", err)
	} else {
		fmt.Printf("send message success: result=%s\n", res.String())
	}

	<-sig
	err = p.Shutdown()
	if err != nil {
		fmt.Printf("shutdown producer error: %s", err.Error())
	}
}
