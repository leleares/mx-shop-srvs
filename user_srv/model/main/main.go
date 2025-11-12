package main

import (
	"crypto/md5"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/anaskhan96/go-password-encoder"
)

// 生成MD5值
func genMD5(originStr string) string {
	MD5 := md5.New()
	_, _ = io.WriteString(MD5, originStr)

	return hex.EncodeToString(MD5.Sum(nil))
}

func main() {
	// dsn := "root:12345678@tcp(127.0.0.1:3306)/mxshop_user_srv?charset=utf8mb4&parseTime=True&loc=Local"
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
	// _ = db.AutoMigrate(&model.User{})

	options := &password.Options{16, 100, 32, sha512.New} // 盐值长度、 迭代次数、 key的长度、 加密算法使用sha512.New比md5更安全
	salt, encodedPwd := password.Encode("generic password", options)
	// 既然每次生成的盐值不同，那如何知道当时为用户生成了什么盐值，做法：将加密算法、盐值、加密后的密码整合成一个字符串存入库中。
	mutiPassword := fmt.Sprintf("$pbkdf2$%s$%s", salt, encodedPwd) // 使用$分割
	mutiPasswordData := strings.Split(mutiPassword, "$")

	check := password.Verify("generic password", mutiPasswordData[2], mutiPasswordData[3], options)
	fmt.Println(check) // true
}
