package server

import "strings"

// Custom Data types
type HandlerFunc func(*HttpContext)

type RouteEntry struct {
	Method  string
	Path    string
	Handler HandlerFunc
	Params  []string // Parameter names (e.g., ["id"])
}

type Router struct {
	routes []RouteEntry
}

func NewRouter() *Router {
	return &Router{
		routes: []RouteEntry{},
	}
}

func (r *Router) handle(method, path string, handler HandlerFunc) {
	params := []string{}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			params = append(params, strings.TrimPrefix(part, ":"))
		}
	}
	r.routes = append(r.routes, RouteEntry{
		Method:  method,
		Path:    path,
		Handler: handler,
		Params:  params,
	})

}
func (r *Router) GET(path string, handler HandlerFunc) {
	r.handle(GET_METHOD, path, handler)
}

func (r *Router) POST(path string, handler HandlerFunc) {
	r.handle(POST_METHOD, path, handler)
}

func (r *Router) PUT(path string, handler HandlerFunc) {
	r.handle(PUT_METHOD, path, handler)
}

func (r *Router) DELETE(path string, handler HandlerFunc) {
	r.handle(DELETE_METHOD, path, handler)
}

func (r *Router) FindHandler(req *HttpRequest) (HandlerFunc, bool) {
	reqPathParts := strings.Split(strings.Trim(req.Path, "/"), "/")

	// First pass: Look for exact matches
	for _, route := range r.routes {
		if route.Method != req.Method {
			continue
		}
		routePathParts := strings.Split(strings.Trim(route.Path, "/"), "/")
		if len(routePathParts) != len(reqPathParts) {
			continue
		}
		isExactMatch := true
		for i, routePart := range routePathParts {
			if strings.HasPrefix(routePart, ":") {
				isExactMatch = false // Contains a parameter, not an exact match
				break
			}
			if routePart != reqPathParts[i] {
				isExactMatch = false
				break
			}
		}
		if isExactMatch {
			req.PathParms = make(PathParams) // No params for exact match
			return route.Handler, true
		}
	}

	// Second pass: Look for parameterized matches
	for _, route := range r.routes {
		if route.Method != req.Method {
			continue
		}
		routePathParts := strings.Split(strings.Trim(route.Path, "/"), "/")
		if len(routePathParts) != len(reqPathParts) {
			continue
		}
		matches := true
		req.PathParms = make(PathParams)
		for i, routePart := range routePathParts {
			reqPart := reqPathParts[i]
			if strings.HasPrefix(routePart, ":") {
				paramName := strings.TrimPrefix(routePart, ":")
				req.PathParms[paramName] = reqPart
			} else if routePart != reqPart {
				matches = false
				break
			}
		}
		if matches {
			return route.Handler, true
		}
	}

	return nil, false
}
