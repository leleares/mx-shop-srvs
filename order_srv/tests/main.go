package main

import (
	"context"
	"fmt"
	"mx-shop-srvs/order_srv/model"
	"mx-shop-srvs/order_srv/proto"

	"google.golang.org/grpc"
)

var (
	orderClient proto.OrderClient
	conn        *grpc.ClientConn
)

func init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:50053", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	orderClient = proto.NewOrderClient(conn)
}

func main() {
	TestCartItemList()
}

func TestCartItemList() {
	resp, err := orderClient.CartItemList(context.Background(), &proto.UserInfo{
		Id: 123456,
	})

	if err != nil {
		fmt.Printf("发生错误", err.Error())
		model.ToStringLog(err)
		return
	}

	model.ToStringLog(resp)
}

func TestCreateCartItem() {
	resp, err := orderClient.CreateCartItem(context.Background(), &proto.CartItemRequest{
		UserId:  123456,
		GoodsId: 654321,
		Nums:    1,
		Checked: true,
	})
	if err != nil {
		model.ToStringLog(err)
		return
	}

	model.ToStringLog(resp)
}

func TestUpdateCartItem() {
	resp, err := orderClient.UpdateCartItem(context.Background(), &proto.CartItemRequest{
		UserId:  123456,
		GoodsId: 654321,
		Nums:    4,
		Checked: true,
	})
	if err != nil {
		fmt.Printf("发生错误", err.Error())
		model.ToStringLog(err)
		return
	}

	model.ToStringLog(resp)
}

func TestDeleteCartItem() {
	resp, err := orderClient.DeleteCartItem(context.Background(), &proto.CartItemRequest{
		UserId:  123456,
		GoodsId: 654321,
	})
	if err != nil {
		fmt.Printf("发生错误", err.Error())
		model.ToStringLog(err)
		return
	}

	model.ToStringLog(resp)
}
