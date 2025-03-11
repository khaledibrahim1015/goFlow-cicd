package server

const (
	POST_METHOD      = "POST"
	PUT_METHOD       = "PUT"
	GET_METHOD       = "GET"
	DELETE_METHOD    = "DELETE"
	APPLICATION_JSON = "application/json"
	TEXT_PLAIN       = "text/plain"
)

// Status codes as constants
const (
	StatusOK                  = 200
	StatusCreated             = 201
	StatusBadRequest          = 400
	StatusNotFound            = 404
	StatusInternalServerError = 500
	StatusMethodNotAllowed    = 405
)

// Status text as constants
const (
	StatusTextOK                  = "OK"
	StatusTextCreated             = "Created"
	StatusTextBadRequest          = "Bad Request"
	StatusTextNotFound            = "Not Found"
	StatusTextInternalServerError = "Internal Server Error"
	StatusTextMethodNotAllowed    = "Method Not Allowed"
)

// StatusCodeText maps status codes to their text (initialized with constants)
var StatusCodeText = map[int]string{
	StatusOK:                  StatusTextOK,
	StatusCreated:             StatusTextCreated,
	StatusBadRequest:          StatusTextBadRequest,
	StatusNotFound:            StatusTextNotFound,
	StatusInternalServerError: StatusTextInternalServerError,
	StatusMethodNotAllowed:    StatusTextMethodNotAllowed,
}

// ResponseMessage provides common response messages
var ResponseMessage = map[string]string{
	"no_data":        "no data exist",
	"invalid_id":     "invalid id",
	"not_found":      "product not found",
	"invalid_json":   "invalid JSON",
	"invalid_fields": "missing or invalid fields",
	"invalid_body":   "invalid request body",
	"invalid_method": "invalid method request",
}
