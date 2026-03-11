package initialize

import (
	"fmt"
	"mx-shop-srvs/order_srv/global"
	"mx-shop-srvs/order_srv/proto"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func InitSrv() {
	s := zap.S()
	// 初始化商品服务👇
	consulInfo := global.ServerConfig.ConsulInfo
	goodsConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", consulInfo.Host, consulInfo.Port, global.ServerConfig.GoodsSrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		s.Fatal("【InitSrvConn】商品服务连接失败")
	}

	goodsClient := proto.NewGoodsClient(goodsConn)
	global.GoodSrvClient = goodsClient

	// 初始化库存服务👇
	inventoryConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", consulInfo.Host, consulInfo.Port, global.ServerConfig.InventorySrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		s.Fatal("【InitSrvConn】库存服务连接失败")
	}
	inventoryClient := proto.NewInventoryClient(inventoryConn)
	global.InventorySrvClient = inventoryClient
}
