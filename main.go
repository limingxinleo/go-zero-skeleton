package main

import (
	"fmt"
	"github.com/zeromicro/go-zero/rest"
	"main/app"
	"main/app/config"
	"main/app/controller"
	"main/app/kernel"
	"net/http"
)

func main() {
	server := rest.MustNewServer(
		config.Conf.RestConf,
		rest.WithCustomCors(
			func(header http.Header) {
				header.Set("Access-Control-Allow-Headers", "DNT,Keep-Alive,User-Agent,Cache-Control,Content-Type,Authorization")
			},
			nil,
			"*",
		),
	)
	defer server.Stop()

	server.Use(kernel.ServerMiddleware)

	controller.RegisterHandlers(server, app.ServiceContext)

	fmt.Printf("Starting server at %s:%d...\n", config.Conf.Host, config.Conf.Port)
	server.Start()
}
