package app

import (
	"github.com/limingxinleo/go-zero-skeleton/app/config"
	"github.com/limingxinleo/go-zero-skeleton/app/kernel/conn"
	"github.com/limingxinleo/go-zero-skeleton/app/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"log"
	"os"
	"path/filepath"
)

var ServiceContext *svc.ServiceContext

func configPath() string {
	root := os.Getenv("ROOT_PATH")
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get working directory: %v", err)
		}

		root = wd
	}

	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "etc/main-api.yaml"
	}

	return filepath.Join(root, path)
}

func init() {
	// 从环境变量中读取配置文件路径
	path := configPath()

	config.Conf = &config.Config{}
	conf.MustLoad(path, config.Conf)

	// 设置链接相关数据
	conn.InitRedis()
	conn.InitMySQL()

	ServiceContext = svc.NewServiceContext(*config.Conf)
}

func TODO() {}
