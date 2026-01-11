package handler

import (
	"context"
	"mx-shop-srvs/goods_srv/global"
	"mx-shop-srvs/goods_srv/model"
	"mx-shop-srvs/goods_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// 品牌分类
func (s *GoodsServer) CategoryBrandList(ctx context.Context, req *proto.CategoryBrandFilterRequest) (*proto.CategoryBrandListResponse, error) {
	var resp proto.CategoryBrandListResponse
	var total int64
	global.DB.Model(&model.GoodsCategoryBrand{}).Count(&total)
	resp.Total = int32(total)

	var categoryBrands []model.GoodsCategoryBrand
	// GORM 默认不会自动加载关联，需要显式 Preload，否则 Category/Brands 是零值
	result := global.DB.
		Preload("Category").
		Preload("Brands").
		Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).
		Find(&categoryBrands)
	if result.RowsAffected == 0 {
		return nil, result.Error
	}

	var categoryBrandsResp []*proto.CategoryBrandResponse
	for _, c := range categoryBrands {
		categoryBrandsResp = append(categoryBrandsResp, &proto.CategoryBrandResponse{
			Id: c.ID,
			Category: &proto.CategoryInfoResponse{
				Id:             c.Category.ID,
				Name:           c.Category.Name,
				Level:          c.Category.Level,
				IsTab:          c.Category.IsTab,
				ParentCategory: c.Category.ParentCategoryID,
			},
			Brand: &proto.BrandInfoResponse{
				Id:   c.Brands.ID,
				Name: c.Brands.Name,
				Logo: c.Brands.Logo,
			},
		})
	}
	resp.Data = categoryBrandsResp
	return &resp, nil
}

// 通过category获取brands
func (s *GoodsServer) GetCategoryBrandList(ctx context.Context, req *proto.CategoryInfoRequest) (*proto.BrandListResponse, error) {
	var categoryBrands []model.GoodsCategoryBrand

	result := global.DB.Preload("Brands").Where("category_id = ?", req.Id).Find(&categoryBrands)
	if result.RowsAffected == 0 {
		return nil, result.Error
	}
	var resp proto.BrandListResponse
	resp.Total = int32(result.RowsAffected)
	var categoryBrandList []*proto.BrandInfoResponse

	for _, c := range categoryBrands {
		categoryBrandList = append(categoryBrandList, &proto.BrandInfoResponse{
			Id:   c.Brands.ID,
			Name: c.Brands.Name,
			Logo: c.Brands.Logo,
		})
	}
	resp.Data = categoryBrandList
	return &resp, nil
}

func (s *GoodsServer) CreateCategoryBrand(ctx context.Context, req *proto.CategoryBrandRequest) (*proto.CategoryBrandResponse, error) {
	if result := global.DB.First(&model.Category{}, req.CategoryId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "分类不存在")
	}

	if result := global.DB.First(&model.Brands{}, req.BrandId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "品牌不存在")
	}

	var categoryBrand model.GoodsCategoryBrand
	categoryBrand.CategoryID = req.CategoryId
	categoryBrand.BrandsID = req.BrandId
	result := global.DB.Create(&categoryBrand)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.Internal, "创建失败")
	}

	return &proto.CategoryBrandResponse{
		Id: categoryBrand.ID,
	}, nil
}

func (s *GoodsServer) DeleteCategoryBrand(ctx context.Context, req *proto.CategoryBrandRequest) (*emptypb.Empty, error) {
	result := global.DB.Where("id = ?", req.Id).Delete(&model.GoodsCategoryBrand{})

	if result.RowsAffected >= 1 {
		return &emptypb.Empty{}, nil
	}

	return &emptypb.Empty{}, status.Errorf(codes.Internal, "删除失败")
}

func (s *GoodsServer) UpdateCategoryBrand(ctx context.Context, req *proto.CategoryBrandRequest) (*emptypb.Empty, error) {
	if result := global.DB.First(&model.Category{}, req.CategoryId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "分类不存在")
	}

	if result := global.DB.First(&model.Brands{}, req.BrandId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "品牌不存在")
	}

	var categoryBrand model.GoodsCategoryBrand
	result := global.DB.First(&categoryBrand, req.Id)
	if result.RowsAffected < 1 {
		return nil, status.Errorf(codes.Internal, "品牌分类不存在")
	}

	categoryBrand.CategoryID = req.CategoryId
	categoryBrand.BrandsID = req.BrandId

	result = global.DB.Save(&categoryBrand)
	if result.RowsAffected < 1 {
		return nil, status.Errorf(codes.Internal, "更新失败")
	}

	return &emptypb.Empty{}, nil
}
