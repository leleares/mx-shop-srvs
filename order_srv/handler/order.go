package handler

import (
	"context"
	"fmt"
	"math/rand"
	"mx-shop-srvs/order_srv/global"
	"mx-shop-srvs/order_srv/model"
	"mx-shop-srvs/order_srv/proto"
	"time"

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

// 生成用户订单编号
func GenOrderSn(uid int32) string {
	// 订单编号生成规则：年月日时分秒+用户id+两位随机数
	now := time.Now()
	rand.Seed(time.Now().UnixNano())

	orderSn := fmt.Sprintf("%d%d%d%d%d%d%d%d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Nanosecond(), uid, rand.Intn(90)+10)
	return orderSn
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
func (s *OrderServer) CreateOrder(ctx context.Context, req *proto.OrderRequest) (*proto.OrderInfoResponse, error) {
	/*
		基本流程：
		1. 从购物车中查询用户选择了哪些商品（认为这些商品是用户想购买的）
		2. 调用商品服务，查询商品价格等信息
		4. 调用库存服务，扣减库存
		5. 创建订单，订单基础信息入库
		6. 订单商品基础信息入库
		7. 更新购物车状态（删除刚刚下单的商品等行为）
		事务保证：
		1. 对于订单的两张表的更新以及购物车状态更新，由于都在一个服务当中，因此可以使用本地事务来解决
		2. 对于跨服务的调用库存扣减服务，应使用分布式事务来进行保证原子性操作
	*/

	var goodIds []int32
	var shoppingCartSelectedItems []model.ShoppingCart
	goodNumsMap := make(map[int32]int32) // 该map用于承载购物车中商品id和数量的映射关系，便于后续根据商品id查数量
	result := global.DB.Where(&model.ShoppingCart{User: req.UserId, Checked: true}).Find(&shoppingCartSelectedItems)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "购物车中无选中商品")
	}

	for _, s := range shoppingCartSelectedItems {
		goodIds = append(goodIds, s.Goods)
		goodNumsMap[s.Goods] = s.Nums
	}

	// 调用商品服务
	resp, err := global.GoodSrvClient.BatchGetGoods(context.Background(), &proto.BatchGoodsIdInfo{
		Id: goodIds,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "批量获取商品数据失败")
	}

	// 用户应付金额
	var totalAmount float32
	// 用于向 OrderGoods 表中插入数据的 model 对象
	var orderGoods []model.OrderGoods
	for _, good := range resp.Data {
		amount := goodNumsMap[good.Id]
		totalAmount += good.ShopPrice * float32(amount) // 累计用户应支付金额
		orderGoods = append(orderGoods, model.OrderGoods{
			// Order: , // 这个空值后面会填补上
			Goods:      good.Id,
			GoodsName:  good.Name,
			GoodsImage: good.GoodsFrontImage,
			Nums:       amount,
		})
	}

	var goodsInfo []*proto.GoodInvInfo
	for _, goodId := range goodIds {
		goodsInfo = append(goodsInfo, &proto.GoodInvInfo{
			GoodId: goodId,
			Num:    goodNumsMap[goodId],
		})
	}

	// 调用库存服务
	_, err = global.InventorySrvClient.Sell(context.Background(), &proto.SellInfo{
		GoodsInfo: goodsInfo,
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "库存不足，扣减失败")
	}

	tx := global.DB.Begin()

	// 创建订单
	order := model.OrderInfo{
		User:    req.UserId,
		OrderSn: GenOrderSn(req.UserId), // 订单号
		// PayType: , // 支付方式
		// Status: , // 订单状态
		// TradeNo: , // 交易号，就是支付宝的订单号，用于查账
		OrderMount: totalAmount, // 总金额
		// PayTime: , // 支付时间
		Address:      req.Address,
		SignerName:   req.Name,
		SingerMobile: req.Mobile,
		Post:         req.Post,
	}

	result = tx.Save(&order)
	if result.Error != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "订单创建失败")
	}

	// 更新 orderGoods
	for index := range orderGoods {
		orderGoods[index].Order = order.ID
	}

	// 将订单商品表数据批量插入至订单商品表中
	result = tx.CreateInBatches(orderGoods, 100) // 将数据批量插入至表中，如果这批数据大于100个，那么grom想办法分批帮我们插入
	if result.Error != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "订单商品创建失败")
	}

	// 更新购物车
	result = tx.Where(&model.ShoppingCart{User: req.Id, Checked: true}).Delete(model.ShoppingCart{})
	if result.Error != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "更新购物车状态失败")
	}

	tx.Commit()

	return &proto.OrderInfoResponse{Id: order.ID, OrderSn: order.OrderSn, Total: order.OrderMount}, nil
}

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
