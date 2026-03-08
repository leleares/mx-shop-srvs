package handler

import (
	"context"
	"mx-shop-srvs/order_srv/global"
	"mx-shop-srvs/order_srv/model"
	"mx-shop-srvs/order_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type OrderServer struct {
	proto.UnimplementedOrderServer // 新版grpc强制要求添加，无意义
}

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// 购物车
func (s *OrderServer) CartItemList(ctx context.Context, req *proto.UserInfo) (*proto.CartItemListResponse, error) {
	var cartItemList []model.ShoppingCart

	result := global.DB.Where("user = ?", req.Id).Find(&cartItemList)
	if result.Error != nil {
		return nil, result.Error
	}

	var resp proto.CartItemListResponse
	resp.Total = int32(result.RowsAffected)
	for _, cartItem := range cartItemList {
		resp.Data = append(resp.Data, &proto.ShoppingCartInfoResponse{
			Id:      cartItem.ID,
			UserId:  cartItem.User,
			GoodsId: cartItem.Goods,
			Nums:    cartItem.Nums,
			Checked: cartItem.Checked,
		})
	}

	return &resp, nil
}

func (s *OrderServer) CreateCartItem(ctx context.Context, req *proto.CartItemRequest) (*proto.ShoppingCartInfoResponse, error) {
	// 分两种情况
	// 第一种情况是该商品不存在，创建商品则直接创建一条记录即可
	// 第二张情况是该商品已经存在了，那就需要直接修改nums字段，无需创建一条新记录
	var cartItem model.ShoppingCart

	result := global.DB.Where("goods = ? and user = ?", req.GoodsId, req.UserId).First(&cartItem)
	if result.RowsAffected == 0 {
		cartItem = model.ShoppingCart{
			User:    req.UserId,
			Goods:   req.GoodsId,
			Nums:    req.Nums,
			Checked: req.Checked,
		}
	} else {
		cartItem.Nums = cartItem.Nums + req.Nums
	}

	result = global.DB.Save(&cartItem)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.Internal, "创建购物车商品数量失败")
	}

	resp := &proto.ShoppingCartInfoResponse{
		Id: cartItem.ID,
	}

	return resp, nil
}

func (s *OrderServer) UpdateCartItem(ctx context.Context, req *proto.CartItemRequest) (*emptypb.Empty, error) {
	var cartItem model.ShoppingCart

	result := global.DB.Where("id = ?", req.Id).First(&cartItem)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "该商品不存在")
	}

	cartItem.Checked = req.Checked
	if req.Nums > 0 {
		cartItem.Nums = req.Nums
	}

	global.DB.Save(&cartItem)

	return &emptypb.Empty{}, nil
}

func (s *OrderServer) DeleteCartItem(ctx context.Context, req *proto.CartItemRequest) (*emptypb.Empty, error) {
	result := global.DB.Where("id = ?", req.Id).Delete(&model.ShoppingCart{})
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品不存在，删除商品失败")
	}

	return &emptypb.Empty{}, nil
}

// 订单
// CreateOrder(context.Context, *OrderRequest) (*OrderInfoResponse, error)
// OrderList(context.Context, *OrderFilterRequest) (*OrderListResponse, error)
// OrderDetail(context.Context, *OrderRequest) (*OrderInfoDetailResponse, error)
// UpdateOrderStatus(context.Context, *OrderStatus) (*emptypb.Empty, error)
