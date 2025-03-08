package server

import (
	"fmt"
	"net"
)

// This struct represents the response we’ll send back to the client.
type HttpResponse struct {
	StatusCode int
	Headers    Headers
	Body       []byte
}

func NewResponse(statusCode int, body []byte) *HttpResponse {
	return &HttpResponse{
		StatusCode: statusCode,
		Headers:    make(Headers),
		Body:       body,
	}
}

func (res *HttpResponse) Write(conn net.Conn) error {

	statusText := getStatusText(res.StatusCode)
	response := fmt.Sprintf("HTTP/1.1 %d %s\r\n", res.StatusCode, statusText)

	// Set Content-Length if body exists and not already set
	if res.Body != nil && len(res.Body) > 0 {
		if _, ok := res.Headers["Content-Length"]; !ok {
			res.Headers["Content-Length"] = fmt.Sprintf("%d", len(res.Body))
		}
	}
	// Write headers
	for key, value := range res.Headers {
		response += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	response += "\r\n" // End headers

	var fullResponse []byte
	if res.Body != nil {
		fullResponse = append([]byte(response), res.Body...)
	} else {
		fullResponse = []byte(response)
	}
	_, err := conn.Write(fullResponse)
	if err != nil {
		return fmt.Errorf("error writing response: %v", err)
	}
	return nil
}

func getStatusText(statusCode int) string {
	switch statusCode {
	case 200:
		return "OK"
	case 400:
		return "Bad Request"
	case 404:
		return "Not Found"
	case 500:
		return "Internal Server Error"
	default:
		return "Unknown"
	}
}
