package main

import (
	"flag"
	"fmt"
	"mx-shop-srvs/user_srv/handler"
	"mx-shop-srvs/user_srv/initialize"
	"mx-shop-srvs/user_srv/proto"
	"net"

	"google.golang.org/grpc"
)

func main() {
	// flag处理的参数可在运行可执行文件时注入
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int("port", 50051, "端口号")

	// initialize
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()

	flag.Parse()
	fmt.Println("ip", *IP)
	fmt.Println("port", *Port)

	server := grpc.NewServer()
	proto.RegisterUserServer(server, &handler.UserServer{})
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic("failed to listen:" + err.Error())
	}

	err = server.Serve(lis)
	if err != nil {
		panic("failed to start grpc!" + err.Error())
	}
}
