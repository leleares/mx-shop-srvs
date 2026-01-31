package main

import (
	"context"
	"mx-shop-srvs/inventory_srv/model"
	"mx-shop-srvs/inventory_srv/proto"

	"google.golang.org/grpc"
)

var (
	inventoryClient proto.InventoryClient
	conn            *grpc.ClientConn
)

func init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:50052", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	inventoryClient = proto.NewInventoryClient(conn)
}

func main() {
	conn.Close()
}

func TestCreateUpdateInventory(id int32, num int32) {
	_, err := inventoryClient.SetInv(context.Background(), &proto.GoodInvInfo{
		GoodId: id,
		Num:    num,
	})
	if err != nil {
		panic(err)
	}
}

func TestInvDetail() {
	resp, err := inventoryClient.InvDetail(context.Background(), &proto.GoodInvInfo{
		GoodId: 433,
	})
	if err != nil {
		panic(err)
	}
	model.ToStringLog(resp)
}

func TestSell() {
	resp, err := inventoryClient.Sell(context.Background(), &proto.SellInfo{
		GoodsInfo: []*proto.GoodInvInfo{
			&proto.GoodInvInfo{GoodId: 430, Num: 1},
			&proto.GoodInvInfo{GoodId: 431, Num: 2},
		},
	})
	if err != nil {
		panic(err)
	}
	model.ToStringLog(resp)
}

func TestRollback() {
	resp, err := inventoryClient.Reback(context.Background(), &proto.SellInfo{
		GoodsInfo: []*proto.GoodInvInfo{
			&proto.GoodInvInfo{GoodId: 431, Num: 2},
			&proto.GoodInvInfo{GoodId: 430, Num: 1},
		},
	})
	if err != nil {
		panic(err)
	}
	model.ToStringLog(resp)
}
