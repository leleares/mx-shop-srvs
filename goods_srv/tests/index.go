package main

import (
	"context"
	"fmt"
	"mx-shop-srvs/goods_srv/proto"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
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
	TestDeleteBanner()
	defer conn.Close()
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

func TestCreateBanner() {
	banner := proto.BannerRequest{
		Index: 0,
		Image: "https://piccdn2.umiwi.com/fe-oss/default/MTc2Nzg1MjkxNzA2.png",
		Url:   "https://www.dedao.cn/",
	}

	resp, _ := goodsClient.CreateBanner(context.Background(), &banner)
	fmt.Println(resp)
}

func TestDeleteBanner() {
	resp, _ := goodsClient.DeleteBanner(context.Background(), &proto.BannerRequest{
		Id: 4,
	})

	fmt.Println(resp)
}

func TestUpdateBanner() {
	banner := proto.BannerRequest{
		Id:    5,
		Index: 2,
		Image: "https://piccdn2.umiwi.com/fe-oss/default/MTc2Nzg1MjkxNzA2.png",
		Url:   "https://www.dedao1.cn/",
	}

	resp, _ := goodsClient.UpdateBanner(context.Background(), &banner)
	fmt.Println(resp)
}

func TestGetBannerList() {
	resp, _ := goodsClient.BannerList(context.Background(), &emptypb.Empty{})
	fmt.Println(resp)
}
