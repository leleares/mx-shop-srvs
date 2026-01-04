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
