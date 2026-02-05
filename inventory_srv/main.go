package main

import (
	"flag"
	"fmt"
	"mx-shop-srvs/inventory_srv/global"
	"mx-shop-srvs/inventory_srv/handler"
	"mx-shop-srvs/inventory_srv/initialize"
	"mx-shop-srvs/inventory_srv/proto"
	"mx-shop-srvs/inventory_srv/utils"
	"mx-shop-srvs/inventory_srv/utils/register/consul"
	"net"
	"os"
	"os/signal"
	"syscall"

	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	// flag处理的参数可在运行可执行文件时注入
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int("port", 50052, "端口号")
	var localHost string = "127.0.0.1"

	// initialize
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()
	initialize.InitRedisClient()
	s := zap.S()
	flag.Parse()
	if *Port == 0 {
		port, err := utils.GetFreeAddr()
		// err 为空，证明没报错
		if err == nil {
			*Port = port
		}
	}
	s.Infof("ip", *IP)
	s.Infof("port", *Port)

	server := grpc.NewServer()
	proto.RegisterInventoryServer(server, &handler.InventoryServer{})
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic("failed to listen:" + err.Error())
	}

	// 注册服务健康检查
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())
	serviceId := uuid.NewV4()
	serviceIdStr := fmt.Sprintf("%s", serviceId)
	registerClient := consul.NewRegistryClient(global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port)
	err = registerClient.Register(localHost, *Port, serviceIdStr, global.ServerConfig.Name, global.ServerConfig.Tags)
	if err != nil {
		s.Errorf("注册健康检查时发生了错误")
	}

	go func() {
		err = server.Serve(lis)
		if err != nil {
			panic("failed to start grpc!" + err.Error())
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	err = registerClient.DeRegister(serviceIdStr)
	if err != nil {
		s.Errorf("注销失败")
	} else {
		s.Info("注销成功")
	}
}
