package main

import (
	"fmt"
	"net/http"

	"github.com/limingxinleo/go-zero-skeleton/app"
	"github.com/limingxinleo/go-zero-skeleton/app/controller"
	"github.com/limingxinleo/go-zero-skeleton/app/kernel"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

func main() {
	logx.MustSetup(logx.LogConf{
		ServiceName: app.GetApplication().Config.Name,
		Level:       "info",
		TimeFormat:  "2006-01-02 15:04:05.000",
	})

	server := rest.MustNewServer(
		app.GetApplication().Config.RestConf,
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

	controller.RegisterHandlers(server, app.GetApplication().ServiceContext)

	fmt.Printf("Starting server at %s:%d...\n", app.GetApplication().Config.Host, app.GetApplication().Config.Port)
	server.Start()
}
