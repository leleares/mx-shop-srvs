package initialize

import "go.uber.org/zap"

func InitLogger() {
	// 需替换掉global才能后续使用zap.S()实例上的方法
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
}
