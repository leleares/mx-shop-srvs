package main

import (
	"context"
	"fmt"
	"mx-shop-srvs/goods_srv/model"
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
	TestGetGoodsList()
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

func TestGetCategoryList() {
	resp, err := goodsClient.GetAllCategorysList(context.Background(), &emptypb.Empty{})
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.JsonData)
}

func TestGetSubCategoryList() {
	resp, err := goodsClient.GetSubCategory(context.Background(), &proto.CategoryListRequest{
		Id: 130358,
	})
	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}

func TestCreateCategory() {
	resp, err := goodsClient.CreateCategory(context.Background(), &proto.CategoryInfoRequest{
		Name:  "测试分类",
		Level: 1,
		IsTab: false,
	})
	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}

func TestUpdateCategory() {
	resp, err := goodsClient.UpdateCategory(context.Background(), &proto.CategoryInfoRequest{
		Id:    238015,
		Name:  "测试分类11",
		Level: 1,
		IsTab: false,
	})
	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}

func TestDeleteCategory() {
	resp, err := goodsClient.DeleteCategory(context.Background(), &proto.DeleteCategoryRequest{
		Id: 238015,
	})
	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}

func TestCategoryBrandList() {
	resp, err := goodsClient.CategoryBrandList(context.Background(), &proto.CategoryBrandFilterRequest{
		Pages:       1,
		PagePerNums: 10,
	})
	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}

func TestGetCategoryBrandList() {
	resp, err := goodsClient.GetCategoryBrandList(context.Background(), &proto.CategoryInfoRequest{
		Id: 130366,
	})
	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}

func TestGetGoodsList() {
	resp, err := goodsClient.GoodsList(context.Background(), &proto.GoodsFilterRequest{
		TopCategory: 136982,
		// KeyWords:    "火龙果",
		// PriceMin:    20,
		// PriceMax:    30,
	})
	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}
