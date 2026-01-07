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
	TestDeleteBrand()
	conn.Close()
}

func TestGetBrandsList() {
	rsp, _ := goodsClient.BrandList(context.Background(), &proto.BrandFilterRequest{
		Pages:       2,
		PagePerNums: 5,
	})

	fmt.Println(rsp)
}

func TestCreateBrand() {
	var brand proto.BrandRequest
	brand.Name = "得到"
	brand.Logo = "https://piccdn2.umiwi.com/fe-oss/default/MTc2MzYyOTAwMTQ1.png"
	rsp, _ := goodsClient.CreateBrand(context.Background(), &brand)

	fmt.Println(rsp)
}

func TestUpdateBrand() {
	var brand proto.BrandRequest
	brand.Id = 1113
	brand.Name = "得到1234"
	brand.Logo = "https://piccdn2.umiwi.com/fe-oss/default/MTc2MzYyOTAwMTQ1.png"
	rsp, _ := goodsClient.UpdateBrand(context.Background(), &brand)

	fmt.Println(rsp)
}

func TestDeleteBrand() {
	var brand proto.BrandRequest
	brand.Id = 1113
	rsp, _ := goodsClient.DeleteBrand(context.Background(), &brand)

	fmt.Println(rsp)
}
