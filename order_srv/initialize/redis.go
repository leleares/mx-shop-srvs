package initialize

import (
	"fmt"
	"mx-shop-srvs/order_srv/global"

	goredislib "github.com/redis/go-redis/v9"
)

func InitRedisClient() {
	client := goredislib.NewClient(&goredislib.Options{
		Addr: fmt.Sprintf("%s:%d", global.ServerConfig.RedisInfo.Host, global.ServerConfig.RedisInfo.Port),
	})

	global.RedisClient = client
}
