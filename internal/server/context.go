package server

import (
	"encoding/json"
	"fmt"
	"net"
)

type HttpContext struct {
	Request  *HttpRequest
	Response *HttpResponse
	conn     net.Conn // Private, used to write the response
}

func NewHttpContext(conn net.Conn, req *HttpRequest) *HttpContext {
	return &HttpContext{
		Request:  req,
		Response: NewResponse(200, nil),
		conn:     conn,
	}
}
func (ctx *HttpContext) JSON(statusCode int, obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}
	ctx.Response.StatusCode = statusCode
	ctx.Response.Body = data
	// if ctx.Request.GetHeader("accept-")
	ctx.Response.Headers["Content-Type"] = APPLICATION_JSON
	return nil

}

func (ctx *HttpContext) Text(statusCode int, text string) {
	ctx.Response.StatusCode = statusCode
	ctx.Response.Body = []byte(text)
	ctx.Response.Headers["Content-Type"] = TEXT_PLAIN
}

func (ctx *HttpContext) Query(key string) (string, error) {
	return ctx.Request.GetQueryParam(key)
}

func (ctx *HttpContext) Param(key string) (string, error) {
	return ctx.Request.GetPathParam(key)
}

func (ctx *HttpContext) WriteResponse() error {
	return ctx.Response.Write(ctx.conn)
}
