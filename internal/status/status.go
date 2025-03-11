package status

import (
	"sync"

	"github.com/khaledibrahim1015/goFlow-cicd/internal/server"
)

type PipelineStatus struct {
	ID     string `json:"id"`
	Status string `json:"status"` // "running", "success", "failed"
	Error  string `json:"error,omitempty"`
}

var (
	statuses = make(map[string]PipelineStatus)
	mu       sync.Mutex
)

func Add(id, status, errorMsg string) {
	mu.Lock()
	defer mu.Unlock()
	statuses[id] = PipelineStatus{ID: id, Status: status, Error: errorMsg}
}

func StatusHandler(ctx *server.HttpContext) {

	if ctx.Request.Method != "GET" {
		ctx.JSON(server.StatusMethodNotAllowed, server.Generalesponse{
			"error":   "Only GET allowed",
			"message": server.StatusCodeText[server.StatusMethodNotAllowed],
		})
		return
	}

	mu.Lock()
	defer mu.Unlock()
	ctx.JSON(server.StatusOK, server.Generalesponse{
		"data":    statuses,
		"message": server.StatusCodeText[server.StatusOK],
	})
}
