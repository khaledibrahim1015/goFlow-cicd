package main

import (
	"fmt"

	"github.com/khaledibrahim1015/goFlow-cicd/internal/config"
	"github.com/khaledibrahim1015/goFlow-cicd/internal/handlers"
	"github.com/khaledibrahim1015/goFlow-cicd/internal/server"
	"github.com/khaledibrahim1015/goFlow-cicd/internal/status"
	"github.com/sirupsen/logrus"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	prdctrl := handlers.NewProductController()
	serv := server.NewHttpServer(":8080")
	serv.GET("/", prdctrl.GetAllProducts)
	serv.POST("/webhook", func(ctx *server.HttpContext) {
		handlers.WebHookHandlerWithConfig(ctx, cfg)
	})
	serv.GET("/status", status.StatusHandler)

	if err := serv.Start(); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
