package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mx-shop-srvs/inventory_srv/model"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// 生成MD5值
func genMD5(originStr string) string {
	MD5 := md5.New()
	_, _ = io.WriteString(MD5, originStr)

	return hex.EncodeToString(MD5.Sum(nil))
}

func main() {
	dsn := "root:12345678@tcp(127.0.0.1:3306)/mxshop_inventory_srv?charset=utf8mb4&parseTime=True&loc=Local"
	// 设置全局logger，作用是打印每个执行的sql语句
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // Disable color
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 生成表时，不要加s后缀
		},
		Logger: newLogger,
	})

	if err != nil {
		panic("failed to connect database")
	}

	// 在库里生成表
	// _ = db.AutoMigrate(&model.StockSellDetail{})
	// stockSellDetail := model.StockSellDetail{
	// 	OrderSn: "wanglele",
	// 	Status:  1,
	// 	Detail:  []model.GoodsDetail{{1, 2}, {2, 3}},
	// }

	// db.Create(stockSellDetail)

	var detail model.StockSellDetail
	_ = db.Model(&model.StockSellDetail{}).Where("order_sn = ?", "wanglele").Find(&detail)
	fmt.Println(detail)
}
