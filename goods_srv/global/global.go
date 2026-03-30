package global

import (
	"mx-shop-srvs/goods_srv/config"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

var (
	ServerConfig *config.ServerConfig  = &config.ServerConfig{} // 全局配置文件
	NacosConfig  *config.NacosConfig   = &config.NacosConfig{}
	DB           *gorm.DB              // 全局DB
	ES           *elasticsearch.Client // 全局es
)
