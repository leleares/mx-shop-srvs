package initialize

import (
	"fmt"
	"log"
	"mx-shop-srvs/goods_srv/global"
	"mx-shop-srvs/goods_srv/model"
	"os"
	"strings"

	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8"
)

func InitEs() {

	address := fmt.Sprintf("http://%s:%d", global.ServerConfig.EsInfo.Host, global.ServerConfig.EsInfo.Port)
	cfg := elasticsearch.Config{
		Addresses: []string{address},
		Logger: &elastictransport.ColorLogger{
			Output:             os.Stdout,
			EnableRequestBody:  true,
			EnableResponseBody: true,
		},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating client: %s", err)
	}
	global.ES = es

	resp, err := global.ES.Indices.Exists([]string{model.EsGoods{}.GetIndexName()})
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		// 索引已存在
		return
	} else {
		// 创建索引
		res, err := global.ES.Indices.Create(model.EsGoods{}.GetIndexName(), es.Indices.Create.WithBody(strings.NewReader(model.EsGoods{}.GetMapping())))
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()
	}

}
