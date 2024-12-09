package main

import (
	"flag"
	"fmt"
	"main/app/config"
	"main/app/controller"
	"main/app/kernel"
	"main/app/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/main-api.yaml", "the config file")

func main() {
	flag.Parse()

	config.Conf = &config.Config{}
	conf.MustLoad(*configFile, config.Conf)

	server := rest.MustNewServer(config.Conf.RestConf)
	defer server.Stop()

	server.Use(kernel.ServerMiddleware)

	ctx := svc.NewServiceContext(*config.Conf)
	controller.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", config.Conf.Host, config.Conf.Port)
	server.Start()
}
