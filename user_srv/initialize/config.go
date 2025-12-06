package initialize

import (
	"mx-shop-srvs/user_srv/global"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func GetEnv(env string) string {
	viper.AutomaticEnv()
	return "dev"
	// return viper.GetString("")
}

func InitConfig() {
	s := zap.S()
	v := viper.New()

	// 获取当前工作目录绝对路径
	workDir, _ := os.Getwd()

	var configFilePath string
	env := GetEnv("env")
	if env == "dev" {
		configFilePath = "config-dev.yaml"
	} else {
		configFilePath = "config-prod.yaml"
	}

	v.SetConfigFile(filepath.Join(workDir, configFilePath))
	err := v.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = v.Unmarshal(global.ServerConfig) // 改变 gloabl.serverConfig 中的值
	if err != nil {
		panic(err)
	}

	s.Infof("全局配置文件：%v", global.ServerConfig)

	v.OnConfigChange(func(e fsnotify.Event) {
		s.Infof("file changed %s", e.Name)
		_ = v.ReadInConfig()
		_ = v.Unmarshal(global.ServerConfig)
		s.Infof("全局配置文件：%v", global.ServerConfig)
	})
}
