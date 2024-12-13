package app

import (
	"flag"
	"github.com/zeromicro/go-zero/core/conf"
	"main/app/config"
	"main/app/svc"
)

var configFile = flag.String("f", "etc/main-api.yaml", "the config file")

var ServiceContext *svc.ServiceContext

func BootApplication() {
	flag.Parse()

	config.Conf = &config.Config{}
	conf.MustLoad(*configFile, config.Conf)

	ServiceContext = svc.NewServiceContext(*config.Conf)
}
