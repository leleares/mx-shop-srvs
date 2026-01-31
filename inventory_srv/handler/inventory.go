package handler

import (
	"context"
	"mx-shop-srvs/inventory_srv/global"
	"mx-shop-srvs/inventory_srv/model"
	"mx-shop-srvs/inventory_srv/proto"

	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
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

// func (s *InventoryServer) InvDetail(ctx context.Context, req *proto.GoodInvInfo) (*proto.GoodInvInfo, error) {

// }

// func (s *InventoryServer) Sell(ctx context.Context, req *proto.SellInfo) (*proto.MsgTips, error) {

// }

// func (s *InventoryServer) Reback(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {

// }
