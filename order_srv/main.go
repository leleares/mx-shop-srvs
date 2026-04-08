package main

import (
	"flag"
	"fmt"
	"log"
	"mx-shop-srvs/order_srv/global"
	"mx-shop-srvs/order_srv/handler"
	"mx-shop-srvs/order_srv/initialize"
	"mx-shop-srvs/order_srv/proto"
	"mx-shop-srvs/order_srv/utils"
	"mx-shop-srvs/order_srv/utils/otgrpc"
	"mx-shop-srvs/order_srv/utils/register/consul"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/opentracing/opentracing-go"
	uuid "github.com/satori/go.uuid"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	// flag处理的参数可在运行可执行文件时注入
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int("port", 50053, "端口号")
	var localHost string = "127.0.0.1"

	// initialize
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()
	initialize.InitRedisClient()
	initialize.InitSrv()
	initialize.InitMQ()

	s := zap.S()
	flag.Parse()
	if *Port == 0 {
		port, err := utils.GetFreeAddr()
		// err 为空，证明没报错
		if err == nil {
			*Port = port
		}
	}
	s.Infof("ip=%s", *IP)
	s.Infof("port=%d", *Port)

	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: "127.0.0.1:6831",
		},
		ServiceName: "mxshop",
	}

	tracer, closer, err := cfg.NewTracer(jaegercfg.Logger(jaeger.StdLogger))
	if err != nil {
		log.Printf("Could not initialize jaeger tracer: %s", err.Error())
		return
	}
	opentracing.SetGlobalTracer(tracer)
	server := grpc.NewServer(grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(tracer)))
	proto.RegisterOrderServer(server, &handler.OrderServer{})
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic("failed to listen:" + err.Error())
	}

	// 注册服务健康检查
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	// 启动服务
	go func() {
		err = server.Serve(lis)
		if err != nil {
			panic("failed to start grpc!" + err.Error())
		}
	}()

	// 服务注册
	serviceId := uuid.NewV4()
	serviceIdStr := fmt.Sprintf("%s", serviceId)
	registerClient := consul.NewRegistryClient(global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port)
	err = registerClient.Register(localHost, *Port, serviceIdStr, global.ServerConfig.Name, global.ServerConfig.Tags)
	if err != nil {
		s.Errorf("注册健康检查时发生了错误")
	}

	// 监听订单超时topic
	c, _ := rocketmq.NewPushConsumer(
		consumer.WithGroupName("mxshop-order"),
		consumer.WithNameServer(global.RocketMQNameServer),
	)
	err = c.Subscribe("order_timeout", consumer.MessageSelector{}, handler.OrderTimeoutCb)
	if err != nil {
		fmt.Println(err.Error())
	}
	// Note: start after subscribe
	err = c.Start()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	defer c.Shutdown()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	err = registerClient.DeRegister(serviceIdStr)
	initialize.CloseMQ()
	closer.Close()
	if err != nil {
		s.Errorf("注销失败")
	} else {
		s.Info("注销成功")
	}
}
