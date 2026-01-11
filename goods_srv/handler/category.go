package handler

import (
	"context"
	"encoding/json"
	"mx-shop-srvs/goods_srv/global"
	"mx-shop-srvs/goods_srv/model"
	"mx-shop-srvs/goods_srv/proto"

	"google.golang.org/protobuf/types/known/emptypb"
)

// 商品分类
func (s *GoodsServer) GetAllCategorysList(ctx context.Context, empty *emptypb.Empty) (*proto.CategoryListResponse, error) {
	var categories []model.Category
	result := global.DB.Where(&model.Category{Level: 1}).Preload("SubCategory.SubCategory").Find(&categories)
	if result.RowsAffected == 0 {
		return nil, result.Error
	}

	b, _ := json.Marshal(categories)

	resp := proto.CategoryListResponse{
		JsonData: string(b),
	}
	return &resp, nil
}

// 获取子分类
// func (s *GoodsServer) GetSubCategory(ctx context.Context, req *proto.CategoryListRequest) (*proto.SubCategoryListResponse, error) {

// }

// CreateCategory(context.Context, *CategoryInfoRequest) (*CategoryInfoResponse, error)
// DeleteCategory(context.Context, *DeleteCategoryRequest) (*emptypb.Empty, error)
// UpdateCategory(context.Context, *CategoryInfoRequest) (*emptypb.Empty, error)
