package main

import (
	"context"
	"fmt"
	"mx-shop-srvs/goods_srv/proto"

	"google.golang.org/grpc"
)

var (
	goodsClient proto.GoodsClient
	conn        *grpc.ClientConn
)

func init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	goodsClient = proto.NewGoodsClient(conn)
}

func main() {
	TestGetBrandsList()
	conn.Close()
}

func TestGetBrandsList() {
	rsp, _ := goodsClient.BrandList(context.Background(), &proto.BrandFilterRequest{
		Pages:       2,
		PagePerNums: 5,
	})

	fmt.Println(rsp)
}
