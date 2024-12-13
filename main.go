package main

import (
	"fmt"
	"github.com/zeromicro/go-zero/rest"
	"main/app"
	"main/app/config"
	"main/app/controller"
	"main/app/kernel"
)

func main() {
	app.BootApplication()

	server := rest.MustNewServer(config.Conf.RestConf)
	defer server.Stop()

	server.Use(kernel.ServerMiddleware)

	controller.RegisterHandlers(server, app.ServiceContext)

	fmt.Printf("Starting server at %s:%d...\n", config.Conf.Host, config.Conf.Port)
	server.Start()
}
