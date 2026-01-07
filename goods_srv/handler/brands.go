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

// 品牌
func (s *GoodsServer) BrandList(ctx context.Context, req *proto.BrandFilterRequest) (*proto.BrandListResponse, error) {
	var brands []model.Brands
	// 利用grom能力进行分页
	result := global.DB.Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Find(&brands)
	if result.Error != nil {
		return nil, result.Error
	}

	var total int64
	// 这里查询Brands表总共有多少条记录
	global.DB.Model(model.Brands{}).Count(&total)

	var data []*proto.BrandInfoResponse
	for _, brand := range brands {
		brandResp := proto.BrandInfoResponse{
			Id:   brand.ID,
			Name: brand.Name,
			Logo: brand.Logo,
		}
		data = append(data, &brandResp)
	}

	var resp proto.BrandListResponse
	resp = proto.BrandListResponse{
		Total: int32(total), // 表里总记录数
		Data:  data,         // 分页数据
	}

	return &resp, nil
}

func (s *GoodsServer) CreateBrand(ctx context.Context, req *proto.BrandRequest) (*proto.BrandInfoResponse, error) {
	var brand model.Brands
	// 校验入参合法性，例如不能有重名的品牌
	result := global.DB.Where("name = ?", req.Name).First(&brand)
	if result.RowsAffected >= 1 {
		return nil, status.Errorf(codes.InvalidArgument, "已存在该品牌")
	}

	brand.Name = req.Name
	brand.Logo = req.Logo

	result = global.DB.Create(&brand)
	if result.Error != nil {
		return nil, status.Error(codes.Internal, result.Error.Error())
	}

	if result.RowsAffected < 1 {
		return nil, status.Errorf(codes.Internal, "创建品牌失败")
	}

	var resp proto.BrandInfoResponse
	resp = proto.BrandInfoResponse{
		Id:   brand.ID,
		Name: brand.Name,
		Logo: brand.Logo,
	}
	return &resp, nil
}

func (s *GoodsServer) DeleteBrand(ctx context.Context, req *proto.BrandRequest) (*emptypb.Empty, error) {
	result := global.DB.Where("id = ?", req.Id).Delete(&model.Brands{})

	if result.RowsAffected >= 1 {
		return &emptypb.Empty{}, nil
	}

	return &emptypb.Empty{}, status.Errorf(codes.Internal, "删除失败")
}

func (s *GoodsServer) UpdateBrand(ctx context.Context, req *proto.BrandRequest) (*emptypb.Empty, error) {
	var brand model.Brands
	result := global.DB.First(&brand, req.Id)
	if result.RowsAffected < 1 {
		return nil, status.Errorf(codes.Internal, "品牌不存在")
	}

	brand.Name = req.Name
	brand.Logo = req.Logo

	result = global.DB.Save(&brand)
	if result.RowsAffected < 1 {
		return nil, status.Errorf(codes.Internal, "更新失败")
	}

	return &emptypb.Empty{}, nil
}
