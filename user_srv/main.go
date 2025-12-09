package main

import (
	"flag"
	"fmt"
	"mx-shop-srvs/user_srv/global"
	"mx-shop-srvs/user_srv/handler"
	"mx-shop-srvs/user_srv/initialize"
	"mx-shop-srvs/user_srv/proto"
	"mx-shop-srvs/user_srv/utils"
	"net"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	s := zap.S()
	// flag处理的参数可在运行可执行文件时注入
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int("port", 0, "端口号")

	// initialize
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()

	flag.Parse()
	s.Infof("ip", *IP)
	s.Infof("port", *Port)
	if *Port == 0 {
		port, err := utils.GetFreeAddr()
		// err 为空，证明没报错
		if err == nil {
			*Port = port
		}
	}

	server := grpc.NewServer()
	proto.RegisterUserServer(server, &handler.UserServer{})
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

	var localHost string = "192.168.130.43"
	var localPort int = port

	check := api.AgentServiceCheck{
		GRPC:                           fmt.Sprintf("%s:%d", localHost, localPort),
		Timeout:                        "5s",
		Interval:                       "5s",
		DeregisterCriticalServiceAfter: "10s",
	}

	registration := new(api.AgentServiceRegistration)
	registration.Address = localHost
	registration.Port = localPort
	registration.ID = global.ServerConfig.Name
	registration.Name = global.ServerConfig.Name
	registration.Tags = []string{"lele", "user", "srv"}
	registration.Check = &check

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		panic(err)
	}
	return nil
}
