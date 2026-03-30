package main

import (
	"bytes"
	"encoding/json"
	"log"
	"mx-shop-srvs/goods_srv/model"
	"os"
	"strconv"
	"time"

	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func main() {
	// dsn := "root:12345678@tcp(127.0.0.1:3306)/mxshop_goods_srv?charset=utf8mb4&parseTime=True&loc=Local"
	// // 设置全局logger，作用是打印每个执行的sql语句
	// newLogger := logger.New(
	// 	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
	// 	logger.Config{
	// 		SlowThreshold: time.Second, // Slow SQL threshold
	// 		LogLevel:      logger.Info, // Log level
	// 		Colorful:      true,        // Disable color
	// 	},
	// )

	// db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
	// 	NamingStrategy: schema.NamingStrategy{
	// 		SingularTable: true, // 生成表时，不要加s后缀
	// 	},
	// 	Logger: newLogger,
	// })

	// if err != nil {
	// 	panic("failed to connect database")
	// }

	// // 在库里生成表
	// _ = db.AutoMigrate(&model.Category{}, &model.Brands{}, &model.GoodsCategoryBrand{}, &model.Banner{}, &model.Goods{})
	Mysql2Es()
}

func Mysql2Es() {
	dsn := "root:12345678@tcp(127.0.0.1:3306)/mxshop_goods_srv?charset=utf8mb4&parseTime=True&loc=Local"
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
		panic(err)
	}

	address := "http://localhost:9200"
	cfg := elasticsearch.Config{
		Addresses: []string{address},
		Logger: &elastictransport.ColorLogger{
			Output:             os.Stdout,
			EnableRequestBody:  true,
			EnableResponseBody: true,
		},
	}

	es, err := elasticsearch.NewClient(cfg)

	var goodsList []model.Goods
	db.Find(&goodsList)
	for _, g := range goodsList {
		esModel := model.EsGoods{
			ID:          g.ID,
			CategoryID:  g.CategoryID,
			BrandsID:    g.BrandsID,
			OnSale:      g.OnSale,
			ShipFree:    g.ShipFree,
			IsNew:       g.IsNew,
			IsHot:       g.IsHot,
			Name:        g.Name,
			ClickNum:    g.ClickNum,
			SoldNum:     g.SoldNum,
			FavNum:      g.FavNum,
			MarketPrice: g.MarketPrice,
			GoodsBrief:  g.GoodsBrief,
			ShopPrice:   g.ShopPrice,
		}

		data, err := json.Marshal(esModel)
		if err != nil {
			log.Fatal(err)
		}

		_, err = es.Index(
			esModel.GetIndexName(),
			bytes.NewReader(data),
			es.Index.WithDocumentID(strconv.Itoa(int(esModel.ID))),
		)
	}
}
