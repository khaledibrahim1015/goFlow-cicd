package main

import (
	"fmt"

	"github.com/khaledibrahim1015/goFlow-cicd/internal/handlers"
	"github.com/khaledibrahim1015/goFlow-cicd/internal/server"
	"github.com/khaledibrahim1015/goFlow-cicd/internal/status"
)

func main() {

	// INtailize server
	serv := server.NewHttpServer(":8080")
	serv.POST("/webhook", handlers.WebHookHandler)
	serv.GET("/status", status.StatusHandler)

	if err := serv.Start(); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
