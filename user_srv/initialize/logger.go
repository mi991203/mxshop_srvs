package initialize

import (
	"fmt"
	"go.uber.org/zap"
)

func InitLogger() {
	config := &zap.Config{
		OutputPaths:   []string{"./mxshop_srvs_user_srv.log", "stderr"},
		Level:         zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:   true,
		Encoding:      "console",
		EncoderConfig: zap.NewDevelopmentEncoderConfig(),
	}
	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("zap 日志组件初始化失败:%s", err.Error()))
	}
	zap.ReplaceGlobals(logger)
}
