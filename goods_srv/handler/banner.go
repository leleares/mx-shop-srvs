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

// 轮播图
func (s *GoodsServer) CreateBanner(ctx context.Context, req *proto.BannerRequest) (*proto.BannerResponse, error) {
	banner := model.Banner{
		Image: req.Image,
		Index: req.Index,
		Url:   req.Url,
	}
	result := global.DB.Create(&banner)
	if result.RowsAffected >= 1 {
		resp := proto.BannerResponse{
			Id:    banner.ID,
			Image: banner.Image,
			Index: banner.Index,
			Url:   banner.Url,
		}
		return &resp, nil
	}

	return nil, result.Error
}

func (s *GoodsServer) DeleteBanner(ctx context.Context, req *proto.BannerRequest) (*emptypb.Empty, error) {
	result := global.DB.Where("id = ?", req.Id).Delete(&model.Banner{})
	if result.RowsAffected >= 1 {
		return nil, nil
	}
	return nil, result.Error
}
func (s *GoodsServer) UpdateBanner(ctx context.Context, req *proto.BannerRequest) (*emptypb.Empty, error) {
	var banner model.Banner
	banner.ID = req.Id
	result := global.DB.First(&banner)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "未找到该banner")
	}

	result = global.DB.Save(&banner)
	if result.RowsAffected >= 1 {
		return nil, nil
	}

	return nil, result.Error
}
func (s *GoodsServer) BannerList(ctx context.Context, req *emptypb.Empty) (*proto.BannerListResponse, error) {
	var bannerList []model.Banner
	result := global.DB.Find(&bannerList)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "未查找到banner")
	}

	var data []*proto.BannerResponse
	for _, banner := range bannerList {
		bannerResp := &proto.BannerResponse{
			Id:    banner.ID,
			Index: banner.Index,
			Url:   banner.Url,
		}

		data = append(data, bannerResp)
	}

	resp := proto.BannerListResponse{
		Total: int32(len(bannerList)),
		Data:  data,
	}
	return &resp, nil
}
