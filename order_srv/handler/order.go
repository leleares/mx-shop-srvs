package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"mx-shop-srvs/order_srv/global"
	"mx-shop-srvs/order_srv/model"
	"mx-shop-srvs/order_srv/proto"
	"sync"
	"time"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"go.uber.org/zap"
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

	result := global.DB.Where("id = ? and user = ?", req.Id, req.UserId).First(&cartItem)
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
	result := global.DB.Where("id = ? and user = ?", req.Id, req.UserId).Delete(&model.ShoppingCart{})
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品不存在，删除商品失败")
	}

	return &emptypb.Empty{}, nil
}

// 事务消息需要
type OrderListener struct{}

type orderTxResult struct {
	Code        codes.Code
	Msg         string
	ID          int32
	OrderAmount float32
}

var orderTxResults sync.Map
var orderTraceContexts sync.Map

func storeOrderTxResult(orderSn string, result orderTxResult) {
	orderTxResults.Store(orderSn, result)
}

func loadOrderTxResult(orderSn string) (orderTxResult, bool) {
	result, ok := orderTxResults.LoadAndDelete(orderSn)
	if !ok {
		return orderTxResult{}, false
	}

	txResult, ok := result.(orderTxResult)
	if !ok {
		return orderTxResult{}, false
	}

	return txResult, true
}

func storeOrderTraceContext(orderSn string, spanContext opentracing.SpanContext) {
	orderTraceContexts.Store(orderSn, spanContext)
}

func deleteOrderTraceContext(orderSn string) {
	orderTraceContexts.Delete(orderSn)
}

func startOrderStepSpan(orderSn, operation string) opentracing.Span {
	parentContext, ok := orderTraceContexts.Load(orderSn)
	if ok {
		if spanContext, ok := parentContext.(opentracing.SpanContext); ok {
			span := opentracing.StartSpan(operation, opentracing.ChildOf(spanContext))
			span.SetTag("order_sn", orderSn)
			return span
		}
	}

	span := opentracing.StartSpan(operation)
	span.SetTag("order_sn", orderSn)
	return span
}

func logSpanError(span opentracing.Span, err error, message string) {
	if span == nil || err == nil {
		return
	}

	span.SetTag("error", true)
	span.LogFields(
		otlog.String("event", "error"),
		otlog.String("message", message),
		otlog.String("detail", err.Error()),
	)
}

func (l *OrderListener) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {
	var orderInfo model.OrderInfo
	_ = json.Unmarshal(msg.Body, &orderInfo)
	txResult := orderTxResult{}
	var goodIds []int32
	var shoppingCartSelectedItems []model.ShoppingCart
	goodNumsMap := make(map[int32]int32) // 该map用于承载购物车中商品id和数量的映射关系，便于后续根据商品id查数量

	shopCartSpan := startOrderStepSpan(orderInfo.OrderSn, "order.select_shopcart")
	result := global.DB.Where(&model.ShoppingCart{User: orderInfo.User, Checked: true}).Find(&shoppingCartSelectedItems)
	if result.RowsAffected == 0 {
		shopCartSpan.SetTag("error", true)
		shopCartSpan.LogFields(otlog.String("message", "没有选中结算的商品"))
		shopCartSpan.Finish()
		txResult.Code = codes.InvalidArgument
		txResult.Msg = "没有选中结算的商品"
		storeOrderTxResult(orderInfo.OrderSn, txResult)
		return primitive.RollbackMessageState
	}
	shopCartSpan.Finish()

	for _, s := range shoppingCartSelectedItems {
		goodIds = append(goodIds, s.Goods)
		goodNumsMap[s.Goods] = s.Nums
	}

	// 调用商品服务
	queryGoodsSpan := startOrderStepSpan(orderInfo.OrderSn, "order.query_goods")
	resp, err := global.GoodSrvClient.BatchGetGoods(context.Background(), &proto.BatchGoodsIdInfo{
		Id: goodIds,
	})
	if err != nil {
		logSpanError(queryGoodsSpan, err, "批量查询商品信息失败")
		queryGoodsSpan.Finish()
		txResult.Code = codes.Internal
		txResult.Msg = "批量查询商品信息失败"
		storeOrderTxResult(orderInfo.OrderSn, txResult)
		return primitive.RollbackMessageState
	}
	queryGoodsSpan.Finish()

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
			GoodsPrice: good.ShopPrice,
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
	queryInventorySpan := startOrderStepSpan(orderInfo.OrderSn, "order.sell_inventory")
	_, err = global.InventorySrvClient.Sell(context.Background(), &proto.SellInfo{
		OrderSn:   orderInfo.OrderSn,
		GoodsInfo: goodsInfo,
	})
	if err != nil {
		logSpanError(queryInventorySpan, err, "扣减库存失败")
		queryInventorySpan.Finish()
		txResult.Code = codes.ResourceExhausted
		txResult.Msg = "扣减库存失败"
		storeOrderTxResult(orderInfo.OrderSn, txResult)
		return primitive.RollbackMessageState
	}
	queryInventorySpan.Finish()

	tx := global.DB.Begin()

	// 创建订单
	orderInfo.OrderMount = totalAmount
	orderInfo.Status = string(model.TradeStatusWaitBuyerPay)

	saveOrderSpan := startOrderStepSpan(orderInfo.OrderSn, "order.save_order")
	result = tx.Save(&orderInfo)
	if result.Error != nil {
		logSpanError(saveOrderSpan, result.Error, "创建订单失败")
		saveOrderSpan.Finish()
		txResult.Code = codes.Internal
		txResult.Msg = "创建订单失败"
		storeOrderTxResult(orderInfo.OrderSn, txResult)
		tx.Rollback()
		return primitive.CommitMessageState
	}
	saveOrderSpan.Finish()

	txResult.ID = orderInfo.ID
	txResult.OrderAmount = totalAmount

	// 更新 orderGoods
	for index := range orderGoods {
		orderGoods[index].Order = orderInfo.ID
	}

	// 将订单商品表数据批量插入至订单商品表中
	saveOrderGoodsSpan := startOrderStepSpan(orderInfo.OrderSn, "order.save_order_goods")
	result = tx.CreateInBatches(orderGoods, 100) // 将数据批量插入至表中，如果这批数据大于100个，那么grom想办法分批帮我们插入
	if result.Error != nil {
		logSpanError(saveOrderGoodsSpan, result.Error, "批量插入订单商品失败")
		saveOrderGoodsSpan.Finish()
		txResult.Code = codes.Internal
		txResult.Msg = "批量插入订单商品失败"
		storeOrderTxResult(orderInfo.OrderSn, txResult)
		tx.Rollback()
		return primitive.CommitMessageState
	}
	saveOrderGoodsSpan.Finish()

	// 更新购物车
	deleteCartSpan := startOrderStepSpan(orderInfo.OrderSn, "order.delete_shopcart")
	result = tx.Where(&model.ShoppingCart{User: orderInfo.User, Checked: true}).Delete(&model.ShoppingCart{})
	if result.Error != nil {
		logSpanError(deleteCartSpan, result.Error, "删除购物车记录失败")
		deleteCartSpan.Finish()
		txResult.Code = codes.Internal
		txResult.Msg = "删除购物车记录失败"
		storeOrderTxResult(orderInfo.OrderSn, txResult)
		tx.Rollback()
		return primitive.CommitMessageState
	}
	deleteCartSpan.Finish()

	// 处理订单超时问题，向mq中投递延时消息，订单30分钟后过期
	if global.OrderProducer == nil {
		zap.S().Error("全局普通producer未初始化")
		tx.Rollback()
		txResult.Code = codes.Internal
		txResult.Msg = "普通producer未初始化"
		storeOrderTxResult(orderInfo.OrderSn, txResult)
		return primitive.CommitMessageState
	}

	msg = &primitive.Message{
		Topic: "order_timeout",
		Body:  msg.Body,
	}
	msg.WithDelayTimeLevel(5) // 5为1分钟，16为30分钟 // 发送延时消息

	sendDelayMsgSpan := startOrderStepSpan(orderInfo.OrderSn, "order.send_timeout_message")
	_, err = global.OrderProducer.SendSync(context.Background(), msg)

	if err != nil {
		logSpanError(sendDelayMsgSpan, err, "发送延时消息失败")
		sendDelayMsgSpan.Finish()
		zap.S().Errorf("发送延时消息失败: %v\n", err)
		tx.Rollback()
		txResult.Code = codes.Internal
		txResult.Msg = "发送延时消息失败"
		storeOrderTxResult(orderInfo.OrderSn, txResult)
		return primitive.CommitMessageState
	}
	sendDelayMsgSpan.Finish()
	tx.Commit()
	txResult.Code = codes.OK
	storeOrderTxResult(orderInfo.OrderSn, txResult)
	return primitive.RollbackMessageState
}

// 消息回查
func (l *OrderListener) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {
	var orderInfo model.OrderInfo
	_ = json.Unmarshal(msg.Body, &orderInfo)

	// 查询订单表看看消息中的 orderSn 是否存在，存在就证明整个链路没有问题，只是当时由于不可抗拒力宕机了。
	result := global.DB.Model(&model.OrderInfo{}).Where("order_sn = ?", orderInfo.OrderSn).Find(&model.OrderInfo{})
	if result.RowsAffected == 1 {
		return primitive.RollbackMessageState
	} else {
		return primitive.CommitMessageState
	}
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

	/*
		1. 向rockemq发送库存预扣减操作的 prepare half 消息。
		2. 调用库存服务，进行扣除库存的操作。
		3. 如果库存不足，那么执行消息的rollback操作，如果库存扣除成功，那么执行本地事务。
		4. 如果本地事务执行失败，那应该给rocketmq发送commit消息（回滚库存）。否则给rocketmq发送更rollback消息。
	*/

	if global.OrderTxProducer == nil {
		return nil, status.Error(codes.Internal, "事务producer未初始化")
	}

	createOrderSpan, ctx := opentracing.StartSpanFromContext(ctx, "order.CreateOrder")
	createOrderSpan.SetTag("user.id", req.UserId)
	defer createOrderSpan.Finish()

	// 创建订单
	order := model.OrderInfo{
		User:         req.UserId,
		OrderSn:      GenOrderSn(req.UserId), // 订单号
		Address:      req.Address,
		SignerName:   req.Name,
		SingerMobile: req.Mobile,
		Post:         req.Post,
	}
	createOrderSpan.SetTag("order.sn", order.OrderSn)
	storeOrderTraceContext(order.OrderSn, createOrderSpan.Context())
	defer deleteOrderTraceContext(order.OrderSn)

	orderJsonStr, _ := json.Marshal(order)

	msg := &primitive.Message{
		Topic: "order_reback", // 库存归还
		Body:  orderJsonStr,
	}
	sendTxMsgSpan := opentracing.StartSpan("order.send_tx_message", opentracing.ChildOf(createOrderSpan.Context()))
	_, err := global.OrderTxProducer.SendMessageInTransaction(ctx, msg)
	if err != nil {
		logSpanError(sendTxMsgSpan, err, "发送事务消息失败")
		sendTxMsgSpan.Finish()
		fmt.Printf("send message error: %s\n", err)
		return nil, status.Error(codes.Internal, "发送消息失败")
	}
	sendTxMsgSpan.Finish()

	// 本地事务执行失败
	waitTxResultSpan := opentracing.StartSpan("order.wait_tx_result", opentracing.ChildOf(createOrderSpan.Context()))
	txResult, ok := loadOrderTxResult(order.OrderSn)
	if !ok {
		waitTxResultSpan.SetTag("error", true)
		waitTxResultSpan.LogFields(otlog.String("message", "未获取到事务执行结果"))
		waitTxResultSpan.Finish()
		return nil, status.Error(codes.Internal, "未获取到事务执行结果")
	}
	waitTxResultSpan.Finish()

	if txResult.Code != codes.OK {
		createOrderSpan.SetTag("error", true)
		createOrderSpan.LogFields(otlog.String("message", txResult.Msg))
		return nil, status.Error(txResult.Code, txResult.Msg)
	}

	return &proto.OrderInfoResponse{Id: txResult.ID, OrderSn: order.OrderSn, Total: txResult.OrderAmount}, nil
}

func (s *OrderServer) OrderList(ctx context.Context, req *proto.OrderFilterRequest) (*proto.OrderListResponse, error) {
	var orderList []model.OrderInfo
	var resp proto.OrderListResponse
	var total int64
	result := global.DB.Model(&model.OrderInfo{}).Where(&model.OrderInfo{User: req.UserId}).Count(&total) // 注意，当req.UserId为空时，会自动去除掉where条件
	if result.Error != nil {
		return nil, result.Error
	}
	resp.Total = int32(total)
	result = global.DB.Where(&model.OrderInfo{User: req.UserId}).Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Find(&orderList)
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
	result = global.DB.Where("`order` = ?", order.ID).Find(&orderGoods) // 注意order是mysql里的关键词，因为：order by😓，因此order要加上反引号
	model.ToStringLog(result)
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

func (s *OrderServer) UpdateOrderStatus(ctx context.Context, req *proto.OrderStatus) (*emptypb.Empty, error) {
	if result := global.DB.Model(&model.OrderInfo{}).Where("order_sn = ?", req.OrderSn).Update("status", req.Status); result.RowsAffected == 0 {
		return &emptypb.Empty{}, status.Errorf(codes.NotFound, "未找到订单信息")
	}
	return &emptypb.Empty{}, nil
}

// 处理订单超时具体逻辑
func OrderTimeoutCb(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	type OrderInfo struct {
		OrderSn string
	}
	for _, msg := range msgs {
		var orderInfo OrderInfo
		err := json.Unmarshal(msg.Body, &orderInfo)
		if err != nil {
			zap.S().Errorf("json解析失败:%s", msg)
			return consumer.ConsumeSuccess, err
		}

		tx := global.DB.Begin()
		// todo...
		// 1. 查询订单的支付状态，如果已支付则什么都不做，如果未支付则修改订单状态为closed
		// 2. 归还库存
		var orderInfoRecord model.OrderInfo
		result := tx.Model(&model.OrderInfo{}).Where("order_sn = ?", orderInfo.OrderSn).Find(&orderInfoRecord)
		if result.RowsAffected == 0 {
			zap.S().Errorf("未找到订单记录:%s", msg)
			tx.Rollback()
			return consumer.ConsumeRetryLater, err
		}
		if orderInfoRecord.Status == string(model.TradeStatusSuccess) {
			// pass 什么也不做
			return consumer.ConsumeSuccess, err
		} else {
			result := tx.Model(&model.OrderInfo{}).Where("order_sn = ?", orderInfo.OrderSn).Update("status", model.TradeStatusClosed)
			if result.RowsAffected == 0 {
				zap.S().Errorf("更新订单状态失败:%s", msg)
				tx.Rollback()
				return consumer.ConsumeRetryLater, err
			}

			// 归还库存逻辑，怎么归还库存？向mq中写入消息:order_reback
			if global.OrderProducer == nil {
				tx.Rollback()
				return consumer.ConsumeRetryLater, status.Error(codes.Internal, "普通producer未初始化")
			}
			order := model.OrderInfo{
				OrderSn: orderInfo.OrderSn, // 订单号
			}

			orderJsonStr, _ := json.Marshal(order)

			msg := &primitive.Message{
				Topic: "order_reback", // 库存归还
				Body:  orderJsonStr,
			}
			_, err = global.OrderProducer.SendSync(context.Background(), msg)
			if err != nil {
				tx.Rollback()
				return consumer.ConsumeRetryLater, err
			}
		}
		tx.Commit()
	}
	return consumer.ConsumeSuccess, nil
}
