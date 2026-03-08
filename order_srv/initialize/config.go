package initialize

import (
	"encoding/json"
	"fmt"
	"mx-shop-srvs/order_srv/global"
	"os"
	"path/filepath"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
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

	err = v.UnmarshalKey("nacos", global.NacosConfig) // 改变 global.NacosConfig 中的值
	if err != nil {
		panic(err)
	}

	nacosConfigData := global.NacosConfig
	// 创建clientConfig
	clientConfig := constant.ClientConfig{
		NamespaceId:         nacosConfigData.Namespace, // 如果需要支持多namespace，我们可以创建多个client,它们有不同的NamespaceId。当namespace是public时，此处填空字符串。
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "tmp/nacos/log",
		CacheDir:            "tmp/nacos/cache",
		LogLevel:            "debug",
	}

	// 创建serverConfigs
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: nacosConfigData.Host,
			Port:   nacosConfigData.Port,
		},
	}

	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		panic(err.Error())
	}

	// 监听nacos配置中心文件变动
	err = configClient.ListenConfig(vo.ConfigParam{
		DataId: nacosConfigData.Dataid,
		Group:  nacosConfigData.Group,
		OnChange: func(namespace, group, dataId, data string) {
			// fmt.Println("group:" + group + ", dataId:" + dataId + ", data:" + data)
		},
	})

	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: nacosConfigData.Dataid,
		Group:  nacosConfigData.Group})

	fmt.Println(content)

	err = json.Unmarshal([]byte(content), global.ServerConfig)
	if err != nil {
		s.Panicf("读取nacos配置中心异常", err.Error())
	}

	s.Infof("全局配置文件：%v", global.ServerConfig)
}
