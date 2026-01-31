package main

import (
	"context"
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
	TestCreateUpdateInventory()
	conn.Close()
}

func TestCreateUpdateInventory() {
	_, err := inventoryClient.SetInv(context.Background(), &proto.GoodInvInfo{
		GoodId: 432,
		Num:    6,
	})
	if err != nil {
		panic(err)
	}
}
