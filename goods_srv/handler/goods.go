package handler

import (
	"mx-shop-srvs/goods_srv/proto"
)

type GoodsServer struct {
	proto.UnimplementedGoodsServer // 新版grpc强制要求添加，无意义
}

// 商品接口
// GoodsList(context.Context, *GoodsFilterRequest) (*GoodsListResponse, error)
// // 用户提交订单有多个商品，需要批量查询商品的信息
// BatchGetGoods(context.Context, *BatchGoodsIdInfo) (*GoodsListResponse, error)
// CreateGoods(context.Context, *CreateGoodsInfo) (*GoodsInfoResponse, error)
// DeleteGoods(context.Context, *DeleteGoodsInfo) (*emptypb.Empty, error)
// UpdateGoods(context.Context, *CreateGoodsInfo) (*emptypb.Empty, error)
// GetGoodsDetail(context.Context, *GoodInfoRequest) (*GoodsInfoResponse, error)
