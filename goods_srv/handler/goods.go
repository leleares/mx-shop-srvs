package handler

import (
	"context"
	"fmt"
	"mx-shop-srvs/goods_srv/global"
	"mx-shop-srvs/goods_srv/model"
	"mx-shop-srvs/goods_srv/proto"
)

type GoodsServer struct {
	proto.UnimplementedGoodsServer // 新版grpc强制要求添加，无意义
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
			subSqlQuery = fmt.Sprintf("(select id from category where parent_category_id = %d))", req.TopCategory)
		} else {
			subSqlQuery = fmt.Sprintf("(select id from category where id = %d))", req.TopCategory)
		}
		localDB = localDB.Where(fmt.Sprintf("category_id in (%s)", subSqlQuery))
	}

	var count int64
	localDB.Model(&model.Goods{}).Count(&count)
	resp.Total = int32(count)

	result := localDB.Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Find(&goodsList)
	if result.Error != nil {
		return nil, result.Error
	}

	var goodInfoResp []*proto.GoodsInfoResponse
	for _, g := range goodsList {
		goodInfoResp = append(goodInfoResp, &proto.GoodsInfoResponse{
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
		})
	}
	resp.Data = goodInfoResp
	return &resp, nil
}

// // 用户提交订单有多个商品，需要批量查询商品的信息
// BatchGetGoods(context.Context, *BatchGoodsIdInfo) (*GoodsListResponse, error)
// CreateGoods(context.Context, *CreateGoodsInfo) (*GoodsInfoResponse, error)
// DeleteGoods(context.Context, *DeleteGoodsInfo) (*emptypb.Empty, error)
// UpdateGoods(context.Context, *CreateGoodsInfo) (*emptypb.Empty, error)
// GetGoodsDetail(context.Context, *GoodInfoRequest) (*GoodsInfoResponse, error)
