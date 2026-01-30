package global

import (
	"mx-shop-srvs/inventory_srv/config"

	"gorm.io/gorm"
)

var (
	ServerConfig *config.ServerConfig = &config.ServerConfig{} // 全局配置文件
	NacosConfig  *config.NacosConfig  = &config.NacosConfig{}
	DB           *gorm.DB             // 全局DB连接
)
