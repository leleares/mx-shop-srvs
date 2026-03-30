package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mx-shop-srvs/goods_srv/global"
	"strconv"

	"gorm.io/gorm"
)

type Category struct {
	BaseModel
	Name             string      `gorm:"type:varchar(20);not null" json:"name"`
	ParentCategoryID int32       `json:"parent"`
	ParentCategory   *Category   `json:"-"`
	SubCategory      []*Category `gorm:"foreignKey:ParentCategoryID;references:ID" json:"sub_category"`
	Level            int32       `gorm:"type:int;not null;default:1" json:"level"`
	IsTab            bool        `gorm:"default:false;not null" json:"is_tab"`
}

type Brands struct {
	BaseModel
	Name string `gorm:"type:varchar(20);not null"`
	Logo string `gorm:"type:varchar(200);default:'';not null"`
}

type GoodsCategoryBrand struct {
	BaseModel
	CategoryID int32 `gorm:"type:int;index:idx_category_brand,unique"`
	Category   Category

	BrandsID int32 `gorm:"type:int;index:idx_category_brand,unique"`
	Brands   Brands
}

func (GoodsCategoryBrand) TableName() string {
	return "goodscategorybrand"
}

type Banner struct {
	BaseModel
	Image string `gorm:"type:varchar(200);not null"`
	Url   string `gorm:"type:varchar(200);not null"`
	Index int32  `gorm:"type:int;default:1;not null"`
}

type Goods struct {
	BaseModel
	CategoryID      int32 `gorm:"type:int;not null"`
	Category        Category
	BrandsID        int32 `gorm:"type:int;not null"`
	Brands          Brands
	OnSale          bool     `gorm:"default:false;not null"`
	ShipFree        bool     `gorm:"default:false;not null"`
	IsNew           bool     `gorm:"default:false;not null"`
	IsHot           bool     `gorm:"default:false;not null"`
	Name            string   `gorm:"type:varchar(50);not null"`
	GoodsSn         string   `gorm:"type:varchar(50);not null"`
	ClickNum        int32    `gorm:"type:int;default:0;not null"`
	SoldNum         int32    `gorm:"type:int;default:0;not null"`
	FavNum          int32    `gorm:"type:int;default:0;not null"`
	MarketPrice     float32  `gorm:"not null"`
	ShopPrice       float32  `gorm:"not null"`
	GoodsBrief      string   `gorm:"type:varchar(100);not null"`
	Images          GormList `gorm:"type:varchar(1000);not null"`
	DescImages      GormList `gorm:"type:varchar(1000);not null"`
	GoodsFrontImage string   `gorm:"type:varchar(200);not null"`
}

func (g *Goods) toEsModel() EsGoods {
	return EsGoods{
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
}

func (g *Goods) syncToES() error {
	esModel := g.toEsModel()
	data, err := json.Marshal(esModel)
	if err != nil {
		return err
	}

	res, err := global.ES.Index(
		esModel.GetIndexName(),
		bytes.NewReader(data),
		global.ES.Index.WithDocumentID(strconv.Itoa(int(esModel.ID))),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("sync goods to es failed: %s", string(respBody))
	}

	return nil
}

func (g *Goods) deleteFromES() error {
	res, err := global.ES.Delete(g.toEsModel().GetIndexName(), strconv.Itoa(int(g.ID)))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// 文档不存在也视为删除成功，避免重复删除导致业务失败。
	if res.StatusCode == 404 {
		return nil
	}
	if res.IsError() {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("delete goods from es failed: %s", string(respBody))
	}

	return nil
}

// 向goods表中插入一行数据时，自动调用的钩子函数
func (g *Goods) AfterCreate(tx *gorm.DB) (err error) {
	return g.syncToES()
}

func (g *Goods) AfterUpdate(tx *gorm.DB) (err error) {
	return g.syncToES()
}

func (g *Goods) AfterDelete(tx *gorm.DB) (err error) {
	return g.deleteFromES()
}
