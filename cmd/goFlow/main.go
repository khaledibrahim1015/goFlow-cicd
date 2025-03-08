package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/khaledibrahim1015/goFlow-cicd/internal/server"
)

type product struct {
	ID    int     `json:"id"` // Note: Field names capitalized and tagged for JSON
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

var products []product = []product{
	{ID: 1, Name: "iphone", Price: 1556.4},
	{ID: 2, Name: "laptop", Price: 4588},
	{ID: 3, Name: "lenovo", Price: 58844},
	{ID: 4, Name: "mac", Price: 158766},
}

func main() {

	serv := server.NewHttpServer(":9099")

	serv.GET("/products", func(ctx *server.HttpContext) {
		if len(products) <= 0 {
			ctx.JSON(400, server.Generalesponse{"error": "no data exist "})
		}
		ctx.JSON(200, server.Generalesponse{
			"data": products,
		})
	})

	// Test path parameter Endpoint
	serv.GET("/products/:id", func(ctx *server.HttpContext) {
		idStr, err := ctx.Param("id")
		if err != nil {
			ctx.JSON(400, server.Generalesponse{"error": "invalid id"})
			return
		}
		// id := 0
		// fmt.Sscanf(idStr, "%d", &id)
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ctx.JSON(400, server.Generalesponse{"error": "invalid id"})
			return
		}
		for _, p := range products {
			if p.ID == id {
				ctx.JSON(200, server.Generalesponse{"data": p})
				return
			}
		}
		ctx.JSON(404, server.Generalesponse{"error": "product not found"})
	})

	// Query string
	serv.GET("/products/query", func(ctx *server.HttpContext) {
		productID, _ := ctx.Query("productid")
		productName, _ := ctx.Query("productname")
		ctx.Text(200, fmt.Sprintf("Product ID: %s, Name: %s", productID, productName))
	})

	// Issues with Current Approach
	// Generic Parsing: ParseBody returns an interface{}, requiring type assertion to map[string]interface{}.
	// Type Assertions: We manually cast id, name, and price from interface{} to specific types (float64, string), which is error-prone if the JSON doesn’t match expectations (e.g., id as a string instead of a number).
	// No Struct Mapping: We’re not leveraging Go’s ability to unmarshal JSON directly into a struct, which would simplify the code.
	// serv.POST("products", func(ctx *server.HttpContext) {
	// 	var newProduct product
	// 	body, err := ctx.Request.ParseBody()
	// 	if err != nil {
	// 		ctx.JSON(400, server.Generalesponse{"error": "invalid request body"})
	// 		return

	// 	}
	// 	// parse body
	// 	data, ok := body.(map[string]interface{})
	// 	if !ok {
	// 		ctx.JSON(400, server.Generalesponse{"error": "invalid JSON format"})
	// 		return
	// 	}
	// 	newProduct.ID = int(data["id"].(float64)) // JSON numbers are float64
	// 	newProduct.Name = data["name"].(string)
	// 	newProduct.Price = data["price"].(float64)
	// 	products = append(products, newProduct)
	// 	ctx.JSON(201, server.Generalesponse{"data": newProduct})

	// })

	serv.POST("/products", func(ctx *server.HttpContext) {
		var newProduct product
		if err := json.Unmarshal(ctx.Request.Body, &newProduct); err != nil {
			ctx.JSON(400, server.Generalesponse{"error": fmt.Sprintf("invalid JSON: %v", err)})
			return
		}
		// Optional: Validate fields if needed
		if newProduct.ID == 0 || newProduct.Name == "" || newProduct.Price <= 0 {
			ctx.JSON(400, server.Generalesponse{"error": "missing or invalid fields"})
			return
		}
		products = append(products, newProduct)
		ctx.JSON(201, server.Generalesponse{"data": newProduct})
	})

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
// curl "http://localhost:9099/products/query?productid=1&productname=iphone"
// # Product ID: 1, Name: iphone
