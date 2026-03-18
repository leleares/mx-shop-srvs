package model

const (
	LEAVING_MESSAGES = iota + 1
	COMPLAINT
	INQUIRY
	POST_SALE
	WANT_TO_BUY
)

// 用户留言表
type LeavingMessages struct {
	BaseModel

	User        int32  `gorm:"type:int;index"`
	MessageType int32  `gorm:"type:int comment '留言类型: 1(留言),2(投诉),3(询问),4(售后),5(求购)'"`
	Subject     string `gorm:"type:varchar(100)"`

	Message string // 不加gorm标签，默认是text类型
	File    string `gorm:"type:varchar(200)"`
}

func (LeavingMessages) TableName() string {
	return "leavingmessages"
}

// 用户地址表
type Address struct {
	BaseModel

	User         int32  `gorm:"type:int;index"`
	Province     string `gorm:"type:varchar(10)"`
	City         string `gorm:"type:varchar(10)"`
	District     string `gorm:"type:varchar(20)"`
	Address      string `gorm:"type:varchar(100)"`
	SignerName   string `gorm:"type:varchar(20)"`
	SignerMobile string `gorm:"type:varchar(11)"`
}

// 用户收藏商品表
type UserFav struct {
	BaseModel

	User  int32 `gorm:"type:int;index:idx_user_goods,unique"`
	Goods int32 `gorm:"type:int;index:idx_user_goods,unique"`
}

func (UserFav) TableName() string {
	return "userfav"
}
