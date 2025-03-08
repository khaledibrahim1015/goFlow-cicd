package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

// Custom Data TYpes for QueryParameters and Headers
type QueryParms map[string]string // e.g., {"productid": "1"}
type Headers map[string]string    // e.g., {"Content-Type": "application/json"}
type PathParams map[string]string // e.g., {"id": "123"}

type HttpRequest struct {
	Method      string
	Path        string
	QueryParms  QueryParms
	PathParms   PathParams
	Headers     Headers
	ContentType string
	Body        []byte
}

func ParseRequest(conn net.Conn) (*HttpRequest, error) {

	// Read request
	reader := bufio.NewReader(conn)
	// Parse request line (e.g., "GET /users/123?key=value HTTP/1.1")
	// Read RequestLine
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading request line %v", err)
	}
	// parse request line
	requestLineParts := strings.Split(requestLine, " ")

	if len(requestLineParts) != 3 {
		return nil, fmt.Errorf("invalid HTTP request")
	}
	method, path, _ := requestLineParts[0], requestLineParts[1], requestLineParts[2]

	// Parse Header
	headers := make(Headers)
	var contentType string
	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == "\r\n" || line == "\n" { // that mean end of headers part
			break // End of headers
		}

		if strings.TrimSpace(line) == "" {
			continue
		}
		// parse header data
		headerPrts := strings.SplitN(line, ":", 2)
		if len(headerPrts) == 2 {
			key := strings.TrimSpace(headerPrts[0])
			value := strings.TrimSpace(headerPrts[1])
			headers[key] = value
			if key == "Content-Type" {
				contentType = value
			}
		}

	}

	// Parse QueryParameters it exist
	// ex: /product?productid=1&productname=iphone
	// it contain path and querystrings keypairs separte with ? and keypairs it self separtae with &
	queryParms := make(QueryParms)
	if strings.Contains(path, "?") {
		parts := strings.SplitN(path, "?", 2)
		path = parts[0]
		queryStr := parts[1]
		queryParis := strings.Split(queryStr, "&")
		for _, pairs := range queryParis {
			parts := strings.Split(pairs, "=")
			if len(parts) == 2 {
				queryParms[parts[0]] = parts[1]
			}
		}
	}

	// Parse Body
	var body []byte
	if method == POST_METHOD || method == PUT_METHOD {
		contentLength := 0
		if cl, ok := headers["Content-Length"]; ok {
			fmt.Sscanf(cl, "%d", &contentLength)
		}
		body = make([]byte, contentLength)
		_, err = reader.Read(body) // fill body
		if err != nil {
			return nil, fmt.Errorf("error reading request body: %v", err)
		}

	}

	return &HttpRequest{
		Method:      method,
		Path:        path,
		Headers:     headers,
		QueryParms:  queryParms,
		PathParms:   make(PathParams), // Initialize path parameters , Will be populated by router
		Body:        body,
		ContentType: contentType,
	}, nil

}

// ParseBody parse requestbody based on Content-Type header
func (req *HttpRequest) ParseBody() (interface{}, error) {

	switch req.ContentType {
	case APPLICATION_JSON:
		// CONVERT IT BODY CONTENT TO JSON
		var data map[string]interface{}
		if err := json.Unmarshal(req.Body, &data); err != nil {
			return nil, fmt.Errorf("error parsing JSON body: %v", err)
		}
		return data, nil
	case TEXT_PLAIN:
		return string(req.Body), nil
	default:
		return nil, fmt.Errorf("unsupported Content-Type: %s", req.ContentType)

	}
}

func (req *HttpRequest) GetBody() ([]byte, error) {
	if req.Method == POST_METHOD || req.Method == PUT_METHOD {
		return req.Body, nil
	}
	return nil, fmt.Errorf("not supported body data for method :%v", req.Method)
}

func (req *HttpRequest) GetHeader(key string) (string, error) {
	if cl, ok := req.Headers[key]; ok {
		return cl, nil
	}
	return "", fmt.Errorf("key NOt exist :%v", key)
}
func (req *HttpRequest) GetQueryParam(key string) (string, error) {
	if cl, ok := req.QueryParms[key]; ok {
		return cl, nil
	}
	return "", fmt.Errorf("key NOt exist :%v", key)
}
func (req *HttpRequest) GetPathParam(key string) (string, error) {
	if cl, ok := req.PathParms[key]; ok {
		return cl, nil
	}
	return "", fmt.Errorf("key not exist :%v", key)
}
