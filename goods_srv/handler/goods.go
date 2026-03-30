package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func CreateGoodReqToModel(req *proto.CreateGoodsInfo, m *model.Goods) *model.Goods {
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

func UpdateGoodReqToModel(req *proto.CreateGoodsInfo, m *model.Goods) *model.Goods {
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
	goodsListResponse := &proto.GoodsListResponse{}
	mustClauses := make([]map[string]interface{}, 0)
	filterClauses := make([]map[string]interface{}, 0)

	if req.KeyWords != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  req.KeyWords,
				"fields": []string{"name", "goods_brief"},
			},
		})
	}
	if req.IsHot {
		filterClauses = append(filterClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"is_hot": req.IsHot,
			},
		})
	}
	if req.IsNew {
		filterClauses = append(filterClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"is_new": req.IsNew,
			},
		})
	}
	if req.PriceMin > 0 {
		filterClauses = append(filterClauses, map[string]interface{}{
			"range": map[string]interface{}{
				"shop_price": map[string]interface{}{
					"gte": req.PriceMin,
				},
			},
		})
	}
	if req.PriceMax > 0 {
		filterClauses = append(filterClauses, map[string]interface{}{
			"range": map[string]interface{}{
				"shop_price": map[string]interface{}{
					"lte": req.PriceMax,
				},
			},
		})
	}
	if req.Brand > 0 {
		filterClauses = append(filterClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"brands_id": req.Brand,
			},
		})
	}

	if req.TopCategory > 0 {
		var category model.Category
		if result := global.DB.First(&category, req.TopCategory); result.RowsAffected == 0 {
			return nil, status.Errorf(codes.NotFound, "商品分类不存在")
		}

		var subQuery string
		switch category.Level {
		case 1:
			subQuery = fmt.Sprintf("select id from category where parent_category_id in (select id from category where parent_category_id=%d)", req.TopCategory)
		case 2:
			subQuery = fmt.Sprintf("select id from category where parent_category_id=%d", req.TopCategory)
		case 3:
			subQuery = fmt.Sprintf("select id from category where id=%d", req.TopCategory)
		}

		type result struct {
			ID int32
		}
		var results []result
		global.DB.Model(model.Category{}).Raw(subQuery).Scan(&results)

		categoryIDs := make([]int32, 0, len(results))
		for _, item := range results {
			categoryIDs = append(categoryIDs, item.ID)
		}

		if len(categoryIDs) == 0 {
			return goodsListResponse, nil
		}

		filterClauses = append(filterClauses, map[string]interface{}{
			"terms": map[string]interface{}{
				"category_id": categoryIDs,
			},
		})
	}

	if req.Pages == 0 {
		req.Pages = 1
	}
	switch {
	case req.PagePerNums > 100:
		req.PagePerNums = 100
	case req.PagePerNums <= 0:
		req.PagePerNums = 10
	}

	boolQuery := map[string]interface{}{}
	if len(mustClauses) > 0 {
		boolQuery["must"] = mustClauses
	}
	if len(filterClauses) > 0 {
		boolQuery["filter"] = filterClauses
	}

	searchBody := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": boolQuery,
		},
		"from":    (req.Pages - 1) * req.PagePerNums,
		"size":    req.PagePerNums,
		"_source": []string{"id"},
	}

	body, err := json.Marshal(searchBody)
	if err != nil {
		return nil, err
	}

	res, err := global.ES.Search(
		global.ES.Search.WithContext(ctx),
		global.ES.Search.WithIndex(model.EsGoods{}.GetIndexName()),
		global.ES.Search.WithBody(bytes.NewReader(body)),
		global.ES.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		respBody, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("es search failed: %s", string(respBody))
	}

	var searchResp struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source model.EsGoods `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err = json.NewDecoder(res.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	goodsIDs := make([]int32, 0, len(searchResp.Hits.Hits))
	goodsListResponse.Total = int32(searchResp.Hits.Total.Value)
	for _, hit := range searchResp.Hits.Hits {
		goodsIDs = append(goodsIDs, hit.Source.ID)
	}
	if len(goodsIDs) == 0 {
		return goodsListResponse, nil
	}

	var goodsList []model.Goods
	result := global.DB.Where("id IN ?", goodsIDs).Preload("Category").Preload("Brands").Find(&goodsList)
	if result.Error != nil {
		return nil, result.Error
	}

	goodsMap := make(map[int32]model.Goods, len(goodsList))
	for _, goods := range goodsList {
		goodsMap[goods.ID] = goods
	}
	for _, goodsID := range goodsIDs {
		if goods, ok := goodsMap[goodsID]; ok {
			goodsListResponse.Data = append(goodsListResponse.Data, GoodModelToResp(goods))
		}
	}

	return goodsListResponse, nil
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
	g := CreateGoodReqToModel(req, &goodInfo)
	g.Category = category
	g.Brands = brand

	tx := global.DB.Begin()
	result = tx.Save(&g) // 注意这里要判断是否执行成功了保存操作，如果没有那就回滚，否则将导致AfterCreate钩子执行从而导致数据不一致问题
	if result.Error != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.NotFound, "新建商品失败")
	}
	if err := tx.Commit().Error; err != nil {
		return nil, status.Errorf(codes.Internal, "提交商品事务失败")
	}
	resp := GoodModelToResp(*g)
	return resp, nil
}

func (s *GoodsServer) DeleteGoods(ctx context.Context, req *proto.DeleteGoodsInfo) (*emptypb.Empty, error) {
	var goods model.Goods
	result := global.DB.First(&goods, req.Id)
	if result.RowsAffected == 0 {
		return &emptypb.Empty{}, status.Errorf(codes.NotFound, "商品不存在")
	}
	if result.Error != nil {
		return &emptypb.Empty{}, result.Error
	}

	result = global.DB.Delete(&goods)
	if result.Error != nil {
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

	g := UpdateGoodReqToModel(req, &goodInfo)
	g.Category = category
	g.Brands = brand

	result = global.DB.Save(&g)
	if result.Error != nil {
		return &emptypb.Empty{}, status.Errorf(codes.NotFound, "更新商品失败")
	}

	return &emptypb.Empty{}, nil
}

func (s *GoodsServer) UpdateGoodsStatus(ctx context.Context, req *proto.UpdateGoodsStatusRequest) (*emptypb.Empty, error) {
	var good model.Goods
	result := global.DB.First(&good, req.Id)
	if result.RowsAffected == 0 {
		return &emptypb.Empty{}, result.Error
	}

	good.IsHot = req.IsHot
	good.IsNew = req.IsNew
	good.OnSale = req.OnSale

	result = global.DB.Save(&good)
	if result.RowsAffected == 0 {
		return &emptypb.Empty{}, result.Error
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
