package handler

import (
	"context"
	"fmt"
	"mx-shop-srvs/goods_srv/global"
	"mx-shop-srvs/goods_srv/model"
	"mx-shop-srvs/goods_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GoodsServer struct {
	proto.UnimplementedGoodsServer // 新版grpc强制要求添加，无意义
}

func GoodModelToResp(g model.Goods) *proto.GoodsInfoResponse {
	return &proto.GoodsInfoResponse{
		Id:              g.ID,
		CategoryId:      g.CategoryID,
		Name:            g.Name,
		GoodsSn:         g.GoodsSn,
		ClickNum:        g.ClickNum,
		SoldNum:         g.SoldNum,
		FavNum:          g.FavNum,
		MarketPrice:     g.MarketPrice,
		ShopPrice:       g.ShopPrice,
		GoodsBrief:      g.GoodsBrief,
		ShipFree:        g.ShipFree,
		Images:          g.Images,
		DescImages:      g.DescImages,
		GoodsFrontImage: g.GoodsFrontImage,
		IsNew:           g.IsNew,
		IsHot:           g.IsHot,
		OnSale:          g.OnSale,
		Category: &proto.CategoryBriefInfoResponse{
			Id:   g.Category.ID,
			Name: g.Category.Name,
		},
		Brand: &proto.BrandInfoResponse{
			Id:   g.Brands.ID,
			Name: g.Brands.Name,
			Logo: g.Brands.Logo,
		},
	}
}

func GoodReqToModel(req *proto.CreateGoodsInfo, m *model.Goods) *model.Goods {
	m.CategoryID = req.CategoryId
	m.BrandsID = req.BrandId
	m.Name = req.Name
	m.GoodsSn = req.GoodsSn
	m.MarketPrice = req.MarketPrice
	m.ShopPrice = req.ShopPrice
	m.GoodsBrief = req.GoodsBrief
	m.ShipFree = req.ShipFree
	m.Images = req.Images
	m.DescImages = req.DescImages
	m.GoodsFrontImage = req.GoodsFrontImage
	m.IsNew = req.IsNew
	m.IsHot = req.IsHot
	m.OnSale = req.OnSale
	return m
}

// 商品接口
/*
	考虑这些过滤条件：
	1. 关键词搜索
	2. 查询新品
	3. 查询热门商品
	4. 通过价格区间筛选
	5. 通过品牌筛选商品
	6. 通过商品分类筛选
*/
/*
	已知所有商品都是归属于三级分类的，但是前端可能传的是1级分类id耳机分类id或者三级分类id
	拿前端传的是1级分类id来说事，这涉及到子查询，首先可以根据查到二级分类，根据二级分类id又可以查出三级分类id，有了这些id则可以查询符合条件的goods
	select * from goods where category_id in
	(select id from category where parent_category_id in // 查出的是三级分类的id
	(select id from category where parent_category_id = 130358)) // 查出的是二级分类的id
*/
func (s *GoodsServer) GoodsList(ctx context.Context, req *proto.GoodsFilterRequest) (*proto.GoodsListResponse, error) {
	var goodsList []model.Goods
	var resp proto.GoodsListResponse
	localDB := global.DB.Model(&model.Goods{})
	if req.KeyWords != "" {
		localDB = localDB.Where("name LIKE ?", "%"+req.KeyWords+"%")
	}
	if req.IsNew == true {
		localDB = localDB.Where("is_new = 1")
	}
	if req.IsHot == true {
		localDB = localDB.Where("is_hot = 1")
	}
	if req.PriceMin > 0 {
		localDB = localDB.Where("shop_price >= ?", req.PriceMin)
	}
	if req.PriceMax > 0 {
		localDB = localDB.Where("shop_price <= ?", req.PriceMax)
	}
	if req.Brand > 0 {
		localDB = localDB.Where("brands_id = ?", req.Brand)
	}

	var subSqlQuery string
	if req.TopCategory > 0 {
		var category model.Category
		if result := global.DB.First(&category, req.TopCategory); result.RowsAffected == 0 {
			return nil, result.Error
		}

		if category.Level == 1 {
			subSqlQuery = fmt.Sprintf("(select id from category where parent_category_id in (select id from category where parent_category_id = %d))", req.TopCategory)
		} else if category.Level == 2 {
			subSqlQuery = fmt.Sprintf("(select id from category where parent_category_id = %d)", req.TopCategory)
		} else {
			subSqlQuery = fmt.Sprintf("select id from category where id = %d", req.TopCategory)
		}
		localDB = localDB.Where(fmt.Sprintf("category_id in (%s)", subSqlQuery))
	}

	var count int64
	localDB.Model(&model.Goods{}).Count(&count)
	resp.Total = int32(count)

	result := localDB.Preload("Category").Preload("Brands").Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Find(&goodsList)
	if result.Error != nil {
		return nil, result.Error
	}

	var goodInfoResp []*proto.GoodsInfoResponse
	for _, g := range goodsList {
		goodInfoResp = append(goodInfoResp, GoodModelToResp(g))
	}
	resp.Data = goodInfoResp
	return &resp, nil
}

// 用户提交订单有多个商品，需要批量查询商品的信息
func (s *GoodsServer) BatchGetGoods(ctx context.Context, req *proto.BatchGoodsIdInfo) (*proto.GoodsListResponse, error) {
	var goodList []model.Goods
	var resp proto.GoodsListResponse
	result := global.DB.Preload("Category").Preload("Brands").Find(&goodList, req.Id)
	if result.Error != nil {
		return nil, result.Error
	}

	var goodsInfoRespList []*proto.GoodsInfoResponse
	for _, g := range goodList {
		goodsInfoRespList = append(goodsInfoRespList, GoodModelToResp(g))
	}

	resp.Total = int32(result.RowsAffected)
	resp.Data = goodsInfoRespList
	return &resp, nil
}

func (s *GoodsServer) CreateGoods(ctx context.Context, req *proto.CreateGoodsInfo) (*proto.GoodsInfoResponse, error) {
	var category model.Category
	result := global.DB.First(&category, req.CategoryId)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品关联分类不存在")
	}

	var brand model.Brands
	result = global.DB.First(&brand, req.BrandId)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品关联品牌不存在")
	}

	var goodInfo model.Goods
	g := GoodReqToModel(req, &goodInfo)
	g.Category = category
	g.Brands = brand

	result = global.DB.Save(&g)
	resp := GoodModelToResp(*g)
	if result.Error != nil {
		return nil, status.Errorf(codes.NotFound, "新建商品失败")
	}

	return resp, nil
}

func (s *GoodsServer) DeleteGoods(ctx context.Context, req *proto.DeleteGoodsInfo) (*emptypb.Empty, error) {
	result := global.DB.First(&model.Goods{}, req.Id)
	if result.RowsAffected == 0 {
		return &emptypb.Empty{}, status.Errorf(codes.NotFound, "商品不存在")
	}

	result = global.DB.Where("id = ?", req.Id).Delete(&model.Goods{})
	if result.RowsAffected == 0 {
		return &emptypb.Empty{}, status.Errorf(codes.Internal, "删除失败")
	}

	return &emptypb.Empty{}, nil
}

func (s *GoodsServer) UpdateGoods(ctx context.Context, req *proto.CreateGoodsInfo) (*emptypb.Empty, error) {
	var category model.Category
	result := global.DB.First(&category, req.CategoryId)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品关联分类不存在")
	}

	var brand model.Brands
	result = global.DB.First(&brand, req.BrandId)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品关联品牌不存在")
	}

	var goodInfo model.Goods
	result = global.DB.First(&goodInfo, req.Id)
	if result.RowsAffected == 0 {
		return &emptypb.Empty{}, status.Errorf(codes.NotFound, "商品不存在")
	}

	g := GoodReqToModel(req, &goodInfo)
	g.Category = category
	g.Brands = brand

	result = global.DB.Save(&g)
	if result.Error != nil {
		return &emptypb.Empty{}, status.Errorf(codes.NotFound, "更新商品失败")
	}

	return &emptypb.Empty{}, nil
}

func (s *GoodsServer) GetGoodsDetail(ctx context.Context, req *proto.GoodInfoRequest) (*proto.GoodsInfoResponse, error) {

	var goodInfo model.Goods
	result := global.DB.Preload("Category").Preload("Brands").Find(&goodInfo, req.Id)
	if result.Error != nil {
		return nil, result.Error
	}

	resp := GoodModelToResp(goodInfo)
	return resp, nil
}
