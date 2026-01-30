package model

// 库存表
type Inventory struct {
	BaseModel
	Good    int32 `gorm:"type:int;index"` // 关联商品id
	Stock   int32 `gorm:"type:int"`       // 库存数量
	Version int32 `gorm:"type:int"`       // 乐观锁专用
}
