package main

import (
	"fmt"

	"github.com/khaledibrahim1015/goFlow-cicd/internal/handlers"
	"github.com/khaledibrahim1015/goFlow-cicd/internal/server"
)

func main() {

	serv := server.NewHttpServer(":9099")
	// Initialize the ProductController
	pc := handlers.NewProductController()

	serv.GET("/products", pc.GetAllProducts)

	// Test path parameter Endpoint
	serv.GET("/products/:id", pc.GetProductByID)

	// Query string
	serv.GET("/products/query", pc.QueryProducts)

	serv.POST("/products", pc.CreateProduct)

	if err := serv.Start(); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}

// Test
// # Start server in one terminal
// go run main.go

// # In another terminal
// curl http://localhost:9099/products
// # {"data":[{"id":1,"name":"iphone","price":1556.4},...]}
// curl http://localhost:9099/products/1
// # {"data":{"id":1,"name":"iphone","price":1556.4}}
// curl http://localhost:9099/products/abc
// # {"error":"invalid id"}
// curl -X POST -H "Content-Type: application/json" -d '{"id": 5, "name": "tablet", "price": 999.99}' http://localhost:9099/products
// # {"data":{"id":5,"name":"tablet","price":999.99}}
// curl "http://localhost:9099/products/query?prdid=1&prdname=iphone"
// # Product ID: 1, Name: iphone
