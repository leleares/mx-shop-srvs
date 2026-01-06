package handler

import (
	"context"
	"mx-shop-srvs/goods_srv/global"
	"mx-shop-srvs/goods_srv/model"
	"mx-shop-srvs/goods_srv/proto"
)

// 品牌
func (s *GoodsServer) BrandList(ctx context.Context, req *proto.BrandFilterRequest) (*proto.BrandListResponse, error) {
	var brands []model.Brands
	result := global.DB.Find(&brands)
	if result.Error != nil {
		return nil, result.Error
	}

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
		Total: int32(len(brands)),
		Data:  data,
	}

	return &resp, nil
}

// CreateBrand(context.Context, *BrandRequest) (*BrandInfoResponse, error)
// DeleteBrand(context.Context, *BrandRequest) (*emptypb.Empty, error)
// UpdateBrand(context.Context, *BrandRequest) (*emptypb.Empty, error)
