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

// CreateBrand(context.Context, *BrandRequest) (*BrandInfoResponse, error)
// DeleteBrand(context.Context, *BrandRequest) (*emptypb.Empty, error)
// UpdateBrand(context.Context, *BrandRequest) (*emptypb.Empty, error)
