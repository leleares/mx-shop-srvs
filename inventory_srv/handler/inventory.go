package handler

import (
	"context"
	"fmt"
	"mx-shop-srvs/inventory_srv/global"
	"mx-shop-srvs/inventory_srv/model"
	"mx-shop-srvs/inventory_srv/proto"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InventoryServer struct {
	proto.UnimplementedInventoryServer // 新版grpc强制要求添加，无意义
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

// 设置库存&更新库存
func (s *InventoryServer) SetInv(ctx context.Context, req *proto.GoodInvInfo) (*emptypb.Empty, error) {
	var inventory model.Inventory
	result := global.DB.Where("good = ?", req.GoodId).Find(&inventory)
	if result.Error != nil {
		return &emptypb.Empty{}, result.Error
	}
	inventory.Good = req.GoodId
	inventory.Stock = req.Num
	result = global.DB.Save(&inventory) // Save兼容Create和Update操作
	if result.Error != nil {
		return &emptypb.Empty{}, result.Error
	}
	return &emptypb.Empty{}, nil
}

func (s *InventoryServer) InvDetail(ctx context.Context, req *proto.GoodInvInfo) (*proto.GoodInvInfo, error) {
	var inventory model.Inventory
	var resp proto.GoodInvInfo
	result := global.DB.Where("good = ?", req.GoodId).Find(&inventory)
	if result.Error != nil {
		return nil, status.Errorf(codes.NotFound, "库存信息不存在")
	}
	resp.GoodId = inventory.Good
	resp.Num = inventory.Stock
	return &resp, nil
}

// 这里会有两个问题
/*
	1. 事务：如果用户要下单三个商品分别为：商品1购买1件，商品2购买两件，商品3购买3件，如果不做任何处理则会发生第一件商品库存扣减，第二个商品库存不足。
	正确逻辑应为：全部扣减成功或者全部扣减失败，使用事务来解决这个问题。
	2. 锁：不做处理高并发场景下会发生超卖情况，例如有两个请求同时对库存为1的商品进行扣减，这需要使用锁来解决。
*/
var m sync.Mutex // 应放在全局位置，保证所有请求协程共享一把锁

func (s *InventoryServer) Sell(ctx context.Context, req *proto.SellInfo) (*proto.MsgTips, error) {
	tx := global.DB.Begin()
	for _, goodInfo := range req.GoodsInfo {
		var inventory model.Inventory
		result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("good = ?", goodInfo.GoodId).Find(&inventory)
		if result.RowsAffected == 0 {
			tx.Rollback()
			return nil, status.Errorf(codes.NotFound, fmt.Sprintf("找不到%s商品库存信息", goodInfo.GoodId))
		}
		if inventory.Stock < goodInfo.Num {
			tx.Rollback()
			return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("%s商品库存信息不足", goodInfo.GoodId))
		}
		inventory.Stock -= goodInfo.Num
		tx.Save(&inventory)
	}
	tx.Commit() // Commit 后才真正执行更新数据库的操作
	return &proto.MsgTips{
		Msg: "success",
	}, nil
}

func (s *InventoryServer) Reback(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	tx := global.DB.Begin()
	for _, goodInfo := range req.GoodsInfo {
		var inventory model.Inventory
		// 悲观锁，本质上利用MySql的行锁能力
		result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("good = ?", goodInfo.GoodId).Find(&inventory)
		if result.RowsAffected == 0 {
			tx.Rollback()
			return nil, status.Errorf(codes.NotFound, fmt.Sprintf("找不到%s商品库存信息", goodInfo.GoodId))
		}
		inventory.Stock += goodInfo.Num
		tx.Save(&inventory)
	}
	tx.Commit()
	return &emptypb.Empty{}, nil
}
