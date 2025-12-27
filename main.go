package main

import (
	"fmt"
	"net/http"

	"github.com/limingxinleo/go-zero-skeleton/app"
	"github.com/limingxinleo/go-zero-skeleton/app/config"
	"github.com/limingxinleo/go-zero-skeleton/app/controller"
	"github.com/limingxinleo/go-zero-skeleton/app/kernel"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

func main() {
	logx.MustSetup(logx.LogConf{
		ServiceName: config.Conf.Name,
		Level:       "info",
		TimeFormat:  "2006-01-02 15:04:05.000",
	})

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
