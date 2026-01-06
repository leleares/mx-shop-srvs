package main

import (
	"flag"
	"fmt"
	"mx-shop-srvs/goods_srv/global"
	"mx-shop-srvs/goods_srv/handler"
	"mx-shop-srvs/goods_srv/initialize"
	"mx-shop-srvs/goods_srv/proto"
	"mx-shop-srvs/goods_srv/utils"
	"net"

	"github.com/hashicorp/consul/api"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	// flag处理的参数可在运行可执行文件时注入
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int("port", 50051, "端口号")

	// initialize
	initialize.InitLogger()
	s := zap.S()
	initialize.InitConfig()
	initialize.InitDB()

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

	server := grpc.NewServer()
	proto.RegisterGoodsServer(server, &handler.GoodsServer{})
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic("failed to listen:" + err.Error())
	}

	// 注册服务健康检查
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())
	err = Register(*Port)
	if err != nil {
		s.Errorf("注册健康检查时发生了错误")
	}

	err = server.Serve(lis)
	if err != nil {
		panic("failed to start grpc!" + err.Error())
	}
}

// 将此服务注册到服务健康检查中心consul
func Register(port int) error {
	conf := api.DefaultConfig()
	conf.Address = fmt.Sprintf("%s:%d", global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port)

	client, err := api.NewClient(conf)
	if err != nil {
		panic(err)
	}

	var localHost string = "127.0.0.1"
	var localPort int = port

	// 注意，当应用负载均衡策略以后，将此srv服务注册到consul时将不能使用固定id的方式，原因在于该服务可能运行于多台服务器上
	// 后面运行的服务会覆盖在consul中的注册，导致consul中永远只有此srv的一个服务。这里使用uuid来为每次向consul中注册生成唯一id
	check := api.AgentServiceCheck{
		GRPC:                           fmt.Sprintf("%s:%d", localHost, localPort),
		Timeout:                        "5s",
		Interval:                       "5s",
		DeregisterCriticalServiceAfter: "10s",
	}

	registration := new(api.AgentServiceRegistration)
	registration.Address = localHost
	registration.Port = localPort
	// 每次启动服务，注册到consul的id将是唯一的。负载均衡需要
	serviceId := fmt.Sprintf("%s", uuid.NewV4())
	registration.ID = serviceId
	registration.Name = global.ServerConfig.Name
	registration.Tags = global.ServerConfig.Tags
	registration.Check = &check

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		panic(err)
	}
	return nil
}
