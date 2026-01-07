package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type GormList []string

type BaseModel struct {
	ID        int32     `gorm:"primaryKey;type:int"` // type：int是告诉数据库使用int类型
	CreatedAt time.Time `gorm:"column:add_time"`
	UpdatedAt time.Time `gorm:"column:update_time"`
	DeletedAt gorm.DeletedAt
	IsDeleted bool
}

// BeforeDelete 在调用 GORM Delete()（包括软删除）时触发。
// GORM 软删除默认只会设置 DeletedAt，不会自动更新自定义的 IsDeleted 字段；
// 这里将 IsDeleted 同步为 true，方便业务侧按该字段做额外判断或兼容旧逻辑。
func (m *BaseModel) BeforeDelete(tx *gorm.DB) error {
	tx.Statement.SetColumn("IsDeleted", true)
	return nil
}

func (g *GormList) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("在将GormList进行转换时失败了:", value))
	}

	return json.Unmarshal(bytes, &g)
}

func (g GormList) Value() (driver.Value, error) {
	if len(g) == 0 {
		return nil, nil
	}
	return json.Marshal(g)
}
