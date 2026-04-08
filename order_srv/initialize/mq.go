package initialize

import (
	"mx-shop-srvs/order_srv/global"
	"mx-shop-srvs/order_srv/handler"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"go.uber.org/zap"
)

func InitMQ() {
	var err error
	s := zap.S()
	global.OrderProducer, err = rocketmq.NewProducer(
		producer.WithNameServer(global.RocketMQNameServer),
		producer.WithGroupName("order_timeout_producer"),
	)
	if err != nil {
		s.Fatal("【InitMQ】初始化RocketMQ producer失败")
	}

	err = global.OrderProducer.Start()
	if err != nil {
		s.Fatal("【InitMQ】启动RocketMQ producer失败")
	}

	global.OrderTxProducer, err = rocketmq.NewTransactionProducer(
		&handler.OrderListener{},
		producer.WithNameServer(global.RocketMQNameServer),
		producer.WithGroupName("order_tx_producer"),
	)
	if err != nil {
		_ = global.OrderProducer.Shutdown()
		global.OrderProducer = nil
		s.Fatal("【InitMQ】初始化RocketMQ transaction producer失败")
	}

	err = global.OrderTxProducer.Start()
	if err != nil {
		_ = global.OrderTxProducer.Shutdown()
		global.OrderTxProducer = nil
		_ = global.OrderProducer.Shutdown()
		global.OrderProducer = nil
		s.Fatal("【InitMQ】启动RocketMQ transaction producer失败")
	}
}

func CloseMQ() {
	if global.OrderTxProducer != nil {
		if err := global.OrderTxProducer.Shutdown(); err != nil {
			zap.S().Warnf("关闭事务producer失败: %v", err)
		}
		global.OrderTxProducer = nil
	}

	if global.OrderProducer != nil {
		if err := global.OrderProducer.Shutdown(); err != nil {
			zap.S().Warnf("关闭普通producer失败: %v", err)
		}
		global.OrderProducer = nil
	}
}
