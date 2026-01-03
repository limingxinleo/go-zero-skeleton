package app

import (
	"github.com/limingxinleo/go-zero-skeleton/app/svc"
)

var ServiceContext *svc.ServiceContext

func init() {
	// 从环境变量中读取配置文件路径
	app = NewApplication()

	app.Init()
}

func TODO() {}
