package model

import (
	"time"

	"gorm.io/gorm"
)

// 基础字段信息（基础表结构信息）
type BaseModel struct {
	ID        int32     `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"column:add_time"`
	UpdatedAt time.Time `gorm:"column:update_time"`
	DeletedAt gorm.DeletedAt
	IsDeleted bool
}

// 用户表
type User struct {
	BaseModel
	Mobile   string     `gorm:"type:varchar(11);index:idx_mobile;unique;not null"`
	Password string     `gorm:"type:varchar(100);not null;"`
	NickName string     `gorm:"type:varchar(20)"`
	Birthday *time.Time `gorm:"type:datetime"`
	Gender   string     `gorm:"column:gender;type:varchar(6);default:'male';comment:'female表示女，male表示男'"`
	Role     int        `gorm:"column:role;default:1;type:int;comment:'1表示普通用户，2表示管理员'"`
}
