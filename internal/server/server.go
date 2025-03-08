package server

import (
	"fmt"
	"net"
)

// Custom Datatypes
type Generalesponse map[string]interface{}

type Server struct {
	Addr   string
	Router *Router
}

func NewHttpServer(addr string) *Server {
	return &Server{
		Addr:   addr,
		Router: NewRouter(),
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("error starting server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("Server listening on %s\n", s.Addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		go s.handleIncomingConnection(conn)
	}
}

func (s *Server) handleIncomingConnection(conn net.Conn) {
	defer conn.Close()
	request, err := ParseRequest(conn)
	if err != nil {
		ctx := NewHttpContext(conn, &HttpRequest{})
		ctx.Text(400, "Bad Request")
		ctx.WriteResponse()
		return
	}

	handler, found := s.Router.FindHandler(request)
	if !found {
		ctx := NewHttpContext(conn, request)
		ctx.Text(404, "Not Found")
		ctx.WriteResponse()
		return
	}
	ctx := NewHttpContext(conn, request)
	// invoke handler
	handler(ctx)

	if err := ctx.WriteResponse(); err != nil {
		fmt.Printf("Error writing response: %v\n", err)
	}

}
func (s *Server) GET(path string, handler HandlerFunc) {
	s.Router.GET(path, handler)
}
func (s *Server) POST(path string, handler HandlerFunc) {
	s.Router.POST(path, handler)
}

func (s *Server) PUT(path string, handler HandlerFunc) {
	s.Router.PUT(path, handler)
}

func (s *Server) DELETE(path string, handler HandlerFunc) {
	s.Router.DELETE(path, handler)
}
