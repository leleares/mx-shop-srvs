package main

import (
	"context"
	"mx-shop-srvs/inventory_srv/model"
	"mx-shop-srvs/inventory_srv/proto"
	"sync"

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
	var wg sync.WaitGroup

	wg.Add(50)
	for i := 0; i < 50; i++ {
		// go TestSell(&wg)
		go TestRollback(&wg)
	}
	wg.Wait()

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

func TestSell(wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := inventoryClient.Sell(context.Background(), &proto.SellInfo{
		GoodsInfo: []*proto.GoodInvInfo{
			&proto.GoodInvInfo{GoodId: 421, Num: 1},
		},
	})
	if err != nil {
		panic(err)
	}
	model.ToStringLog(resp)
}

func TestRollback(wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := inventoryClient.Reback(context.Background(), &proto.SellInfo{
		GoodsInfo: []*proto.GoodInvInfo{
			&proto.GoodInvInfo{GoodId: 421, Num: 1},
		},
	})
	if err != nil {
		panic(err)
	}
	model.ToStringLog(resp)
}
