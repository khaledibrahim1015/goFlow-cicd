package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/khaledibrahim1015/goFlow-cicd/internal/server"
)

// Models
// Product represents a product entity
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// Intial datasets Repository
var productLst []Product = []Product{
	{ID: 1, Name: "iphone", Price: 1556.4},
	{ID: 2, Name: "laptop", Price: 4588},
	{ID: 3, Name: "lenovo", Price: 58844},
	{ID: 4, Name: "mac", Price: 158766},
}

type ProductController struct {
	products []Product
}

func NewProductController() *ProductController {
	return &ProductController{
		products: productLst,
	}
}

// handlers
// GetAllProducts handles GET /products
func (pc *ProductController) GetAllProducts(ctx *server.HttpContext) {

	if len(pc.products) <= 0 {
		ctx.JSON(400, server.Generalesponse{
			"error": "no data exist",
		})
		return
	}
	ctx.JSON(server.StatusOK, server.Generalesponse{
		"data":    pc.products,
		"message": server.StatusTextOK,
	})
}

// GetProductByID handles GET /products/:id
func (pc *ProductController) GetProductByID(ctx *server.HttpContext) {
	idStr, err := ctx.Param("id")
	if err != nil {
		ctx.JSON(server.StatusBadRequest, server.Generalesponse{
			"error":   server.ResponseMessage["invalid_id"],
			"message": server.StatusCodeText[server.StatusBadRequest],
		})
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(server.StatusBadRequest, server.Generalesponse{
			"error":   server.ResponseMessage["invalid_id"],
			"message": server.StatusCodeText[server.StatusBadRequest],
		})
		return
	}
	for _, value := range pc.products {
		if value.ID == id {
			ctx.JSON(server.StatusOK, server.Generalesponse{
				"data":    value,
				"message": server.StatusCodeText[server.StatusOK],
			})
			return
		}
	}

	ctx.JSON(server.StatusNotFound, server.Generalesponse{
		"error":   server.ResponseMessage["not_found"],
		"message": server.StatusCodeText[server.StatusNotFound],
	})

}

// QueryProducts handles GET /products/query
// expected /products/query?prdid=value&prdname=value

func (pc *ProductController) QueryProducts(ctx *server.HttpContext) {
	prdidstr, err := ctx.Query("prdid")
	if err != nil {
		ctx.Text(server.StatusBadRequest, fmt.Sprintf("prdid : %s", server.ResponseMessage["invalid_id"]))
		return
	}
	prdid, err := strconv.Atoi(prdidstr)
	if err != nil {
		ctx.Text(server.StatusBadRequest, fmt.Sprintf("prdid : %s", server.ResponseMessage["invalid_id"]))
		return
	}

	prdname, err := ctx.Query("prdname")
	if err != nil {
		ctx.Text(server.StatusBadRequest, fmt.Sprintf("prdname : %s", server.ResponseMessage["invalid_fields"]))
		return
	}

	for _, product := range pc.products {
		if product.ID == prdid && product.Name == prdname {
			ctx.Text(server.StatusOK, fmt.Sprintf("Product ID: %v, Name: %v, Product Price: %v",
				product.ID, product.Name, product.Price))
		}
	}

	ctx.Text(server.StatusNotFound, fmt.Sprintf("Product: %s ", server.ResponseMessage["not_found"]))

}

// CreateProduct handles POST /products
func (pc *ProductController) CreateProduct(ctx *server.HttpContext) {
	var newProduct Product
	if err := json.Unmarshal(ctx.Request.Body, &newProduct); err != nil {
		ctx.JSON(server.StatusBadRequest, server.Generalesponse{
			"error":   fmt.Sprintf("%s: %v", server.ResponseMessage["invalid_json"], err),
			"message": server.StatusCodeText[server.StatusBadRequest],
		})
		return
	}

	// Validate fields
	if newProduct.ID == 0 || newProduct.Name == "" || newProduct.Price <= 0 {
		ctx.JSON(server.StatusBadRequest, server.Generalesponse{
			"error":   server.ResponseMessage["invalid_fields"],
			"message": server.StatusCodeText[server.StatusBadRequest],
		})
		return
	}

	pc.products = append(pc.products, newProduct)
	ctx.JSON(server.StatusCreated, server.Generalesponse{
		"data":    newProduct,
		"message": server.StatusCodeText[server.StatusCreated],
	})
}

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
