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
	TestDeleteGood()
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

func TestBatchGetGoodsList() {
	resp, err := goodsClient.BatchGetGoods(context.Background(), &proto.BatchGoodsIdInfo{
		Id: []int32{430, 436},
	})
	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}

func TestGetGoodDetail() {
	resp, err := goodsClient.GetGoodsDetail(context.Background(), &proto.GoodInfoRequest{
		Id: 427,
	})
	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}

func TestUpdateGoodDetail() {
	resp, err := goodsClient.UpdateGoods(context.Background(), &proto.CreateGoodsInfo{
		Id:              444,
		CategoryId:      135501,
		BrandId:         614,
		Name:            "四川攀枝花凯特芒果 2.5kg装 单果400g以上 新鲜水果test",
		GoodsSn:         "1",
		MarketPrice:     39.9,
		ShopPrice:       39.9,
		GoodsBrief:      "四川攀枝花凯特芒果 2.5kg装 单果400g以上 新鲜水果test",
		ShipFree:        true,
		Images:          []string{"https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/5e17643c07463615c30aa87aeffd164f", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/07d8bc1a084e6c4545123e30f46084da", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/04e4d9be71317ba9d6145a2077d2cacb", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/b01624d3aa0dc5414ffcb4d3bf4f93c5", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/cb34219fbbf90cf6c52c26bbe0aefcaa", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/c73c5e3b8124193b255dc9d3e1589db7"},
		DescImages:      []string{"https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/5e17643c07463615c30aa87aeffd164f", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/07d8bc1a084e6c4545123e30f46084da", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/04e4d9be71317ba9d6145a2077d2cacb", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/b01624d3aa0dc5414ffcb4d3bf4f93c5", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/cb34219fbbf90cf6c52c26bbe0aefcaa", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/c73c5e3b8124193b255dc9d3e1589db7"},
		GoodsFrontImage: "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/5e17643c07463615c30aa87aeffd164f",
		IsNew:           false,
		IsHot:           true,
		OnSale:          true,
	})
	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}

func TestCreateGood() {
	resp, err := goodsClient.CreateGoods(context.Background(), &proto.CreateGoodsInfo{
		CategoryId:      135501,
		BrandId:         614,
		Name:            "四川攀枝花凯特芒果 2.5kg装 单果400g以上 新鲜水果 v2",
		GoodsSn:         "1",
		MarketPrice:     39.99,
		ShopPrice:       38.9,
		GoodsBrief:      "四川攀枝花凯特芒果 2.5kg装 单果400g以上 新鲜水果 v2",
		ShipFree:        true,
		Images:          []string{"https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/5e17643c07463615c30aa87aeffd164f", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/07d8bc1a084e6c4545123e30f46084da", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/04e4d9be71317ba9d6145a2077d2cacb", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/b01624d3aa0dc5414ffcb4d3bf4f93c5", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/cb34219fbbf90cf6c52c26bbe0aefcaa", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/c73c5e3b8124193b255dc9d3e1589db7"},
		DescImages:      []string{"https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/5e17643c07463615c30aa87aeffd164f", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/07d8bc1a084e6c4545123e30f46084da", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/04e4d9be71317ba9d6145a2077d2cacb", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/b01624d3aa0dc5414ffcb4d3bf4f93c5", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/cb34219fbbf90cf6c52c26bbe0aefcaa", "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/c73c5e3b8124193b255dc9d3e1589db7"},
		GoodsFrontImage: "https://py-go.oss-cn-beijing.aliyuncs.com/goods_images/5e17643c07463615c30aa87aeffd164f",
		IsNew:           false,
		IsHot:           true,
		OnSale:          true,
	})
	if err != nil {
		panic(err)
	}

	model.ToStringLog(resp)
}

func TestDeleteGood() {
	resp, err := goodsClient.DeleteGoods(context.Background(), &proto.DeleteGoodsInfo{
		Id: 846,
	})

	if err != nil {
		panic(err)
	}
	model.ToStringLog(resp)
}
