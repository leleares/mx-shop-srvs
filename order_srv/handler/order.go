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
// func (s *OrderServer) CreateOrder(ctx context.Context, req *proto.OrderRequest) (*proto.OrderInfoResponse, error) {

// }

func (s *OrderServer) OrderList(ctx context.Context, req *proto.OrderFilterRequest) (*proto.OrderListResponse, error) {
	var orderList []model.OrderInfo
	var resp proto.OrderListResponse
	var total int64
	global.DB.Where(&model.OrderInfo{User: req.UserId}).Count(&total) // 注意，当req.UserId为空时，会自动去除掉where条件
	resp.Total = int32(total)
	result := global.DB.Where(&model.OrderInfo{User: req.UserId}).Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Find(&orderList)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, order := range orderList {
		resp.Data = append(resp.Data, &proto.OrderInfoResponse{
			Id:      order.ID,
			UserId:  order.User,
			OrderSn: order.OrderSn,
			PayType: order.PayType,
			Status:  order.Status,
			Post:    order.Post,
			Address: order.Address,
			Name:    order.SignerName,
			Mobile:  order.SingerMobile,
			Total:   order.OrderMount,
			AddTime: order.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &resp, nil
}

func (s *OrderServer) OrderDetail(ctx context.Context, req *proto.OrderRequest) (*proto.OrderInfoDetailResponse, error) {
	var order model.OrderInfo

	// 注意这里要判断该订单是否是当前用户的
	result := global.DB.Where(&model.OrderInfo{BaseModel: model.BaseModel{ID: req.Id}, User: req.UserId}).First(&order)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "未找到订单信息")
	}

	var resp proto.OrderInfoDetailResponse
	resp.OrderInfo = &proto.OrderInfoResponse{
		Id:      order.ID,
		UserId:  order.User,
		OrderSn: order.OrderSn,
		PayType: order.PayType,
		Status:  order.Status,
		Post:    order.Post,
		Address: order.Address,
		Name:    order.SignerName,
		Mobile:  order.SingerMobile,
		Total:   order.OrderMount,
		AddTime: order.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	var orderGoods []model.OrderGoods
	result = global.DB.Where("order = ?", order.ID).Find(&orderGoods)
	for _, orderGood := range orderGoods {
		resp.Goods = append(resp.Goods, &proto.OrderItemResponse{
			Id:         orderGood.ID,
			OrderId:    orderGood.Order,
			GoodsId:    orderGood.Goods,
			GoodsName:  orderGood.GoodsName,
			GoodsImage: orderGood.GoodsImage,
			GoodsPrice: orderGood.GoodsPrice,
			Nums:       orderGood.Nums,
		})
	}

	return &resp, nil
}

// func (s *OrderServer) UpdateOrderStatus(ctx context.Context, req *proto.OrderStatus) (*emptypb.Empty, error) {
// }
