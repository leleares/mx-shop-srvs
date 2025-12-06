package global

import (
	"mx-shop-srvs/user_srv/config"

	"gorm.io/gorm"
)

var (
	ServerConfig *config.ServerConfig = &config.ServerConfig{} // 全局配置文件
	DB           *gorm.DB                                      // 全局DB连接
)
