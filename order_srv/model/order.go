package model

import "time"

// 购物车表
type ShoppingCart struct {
	BaseModel
	User    int32 `gorm:"type:int;index"` // 关联用户
	Goods   int32 `gorm:"type:int;index"` // 关联商品
	Nums    int32 `gorm:"type:int"`       // 购物车中该商品有几件
	Checked bool  // 用户在购物车中是否勾选了该商品
}

func (ShoppingCart) TableName() string {
	return "shoppingcart"
}

// 订单基础信息表结构
type OrderInfo struct {
	BaseModel
	User         int32  `gorm:"type:int;index"`
	OrderSn      string `gorm:"type:varchar(30);index"` //订单号，我们平台自己生成的订单号
	PayType      string `gorm:"type:varchar(20) comment 'alipay(支付宝)， wechat(微信)'"`
	Status       string `gorm:"type:varchar(20)  comment 'PAYING(待支付), TRADE_SUCCESS(成功)， TRADE_CLOSED(超时关闭), WAIT_BUYER_PAY(交易创建), TRADE_FINISHED(交易结束)'"`
	TradeNo      string `gorm:"type:varchar(100) comment '交易号'"` //交易号就是支付宝的订单号 查账
	OrderMount   float32
	PayTime      *time.Time `gorm:"type:datetime"`
	Address      string     `gorm:"type:varchar(100)"`
	SignerName   string     `gorm:"type:varchar(20)"`
	SingerMobile string     `gorm:"type:varchar(11)"`
	Post         string     `gorm:"type:varchar(20)"`
}

// 订单商品信息表结构
func (OrderInfo) TableName() string {
	return "orderinfo"
}

type OrderGoods struct {
	BaseModel
	Order int32 `gorm:"type:int;index"`
	Goods int32 `gorm:"type:int;index"`
	//把商品的信息保存下来了，这属于字段冗余，不符合sql设计三范式，但在高并发系统中我们一般都不会遵循三范式，主要出于性能考虑与业务便捷性
	GoodsName  string `gorm:"type:varchar(100);index"`
	GoodsImage string `gorm:"type:varchar(200)"`
	GoodsPrice float32
	Nums       int32 `gorm:"type:int"`
}

func (OrderGoods) TableName() string {
	return "ordergoods"
}
