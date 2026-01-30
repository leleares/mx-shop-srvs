package handler

import (
	"mx-shop-srvs/inventory_srv/proto"

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
