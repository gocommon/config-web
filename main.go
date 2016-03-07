package main

import (
	config "github.com/micro/config-srv/proto/config"
	"github.com/micro/config-web/handler"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-web"
)

func main() {
	service := web.NewService(
		web.Name("go.micro.web.config"),
		web.Handler(handler.Router()),
	)

	service.Init()

	handler.Init(
		"templates",
		config.NewConfigClient("go.micro.srv.config", client.DefaultClient),
	)

	service.Run()
}
