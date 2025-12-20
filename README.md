mx-shop-srvs/  
├── user_srv/                 # 用户相关微服务  
├── model/                    # 定义表结构
├──────── main/                
├──────────────main.go        # 定义连接数据库，生成表结构相关内容


# 运行user_srv前，请先检查main.go文件中的ip地址是否是本机ip地址，否则consul会失去健康检查。