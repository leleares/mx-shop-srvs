package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type GoodsDetail struct {
	Goods int32
	Num   int32
}

type GoodsListType []GoodsDetail

func (g *GoodsListType) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("在将GormList进行转换时失败了:", value))
	}

	return json.Unmarshal(bytes, &g)
}

func (g GoodsListType) Value() (driver.Value, error) {
	if len(g) == 0 {
		return nil, nil
	}
	return json.Marshal(g)
}

// 库存表
type Inventory struct {
	BaseModel
	Good    int32 `gorm:"type:int;index"` // 关联商品id
	Stock   int32 `gorm:"type:int"`       // 库存数量
	Version int32 `gorm:"type:int"`       // 乐观锁专用
}

// 记录订单中商品销售与归还情况。
type StockSellDetail struct {
	OrderSn string        `gorm:"type:varchar(200);index:unique"`
	Status  int32         `gorm:"type:int"` // 1表示售卖 2表示归还
	Detail  GoodsListType `gorm:"type:varchar(200)"`
}

func (StockSellDetail) TableName() string {
	return "stockselldetail"
}
