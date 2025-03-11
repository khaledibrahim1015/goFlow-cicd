package git

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/khaledibrahim1015/goFlow-cicd/internal/config"
	"github.com/khaledibrahim1015/goFlow-cicd/internal/pipeline"
	"github.com/khaledibrahim1015/goFlow-cicd/internal/server"
	"github.com/khaledibrahim1015/goFlow-cicd/internal/status"
	"github.com/sirupsen/logrus"
)

func GitLabhandler(ctx *server.HttpContext, cfg *config.PipelineConfig) {
	if ctx.Request.Headers["X-Gitlab-Event"] != "Push Hook" {
		ctx.JSON(server.StatusBadRequest, server.Generalesponse{
			"error":   fmt.Sprintf("Only push events supported"),
			"message": server.StatusCodeText[server.StatusBadRequest],
		})
		return
	}
	token, err := ctx.Request.GetHeader("X-Gitlab-Token")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, server.Generalesponse{
			"error":   "Invalid token",
			"message": "Unauthorized",
		})
		return
	}

	var repo *config.RepositoryConfig
	for _, r := range cfg.Repositories {
		if r.Secret == token {
			repo = &r
			break
		}
	}
	var payload = ctx.Request.Body

	if repo == nil {
		ctx.JSON(http.StatusUnauthorized, server.Generalesponse{
			"error":   "Invalid token",
			"message": "Unauthorized",
		})
		return
	}
	var event struct {
		Ref     string `json:"ref"`
		Project struct {
			URL string `json:"web_url"`
		} `json:"project"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		ctx.JSON(http.StatusBadRequest, server.Generalesponse{
			"error":   "Invalid payload",
			"message": server.StatusCodeText[server.StatusBadRequest],
		})
		return

	}

	if event.Ref != "refs/heads/"+repo.Branch {
		ctx.JSON(http.StatusOK, server.Generalesponse{
			"error":   "Ignored (wrong branch)",
			"message": server.StatusCodeText[server.StatusOK],
		})
		return
	}

	pipelineID := fmt.Sprintf("%s-%s", repo.URL, event.Ref)
	repoPath, err := Clone(repo.URL, repo.Branch)
	if err != nil {
		ctx.JSON(server.StatusInternalServerError, server.Generalesponse{
			"error":   fmt.Sprintf("Clone failed: %v", err),
			"message": server.StatusCodeText[server.StatusInternalServerError],
		})
		return

	}

	// TRigger Pipeline
	status.Add(pipelineID, "running", "")
	go func() {
		p := pipeline.New(cfg, repoPath)
		if err := p.Run(); err != nil {
			logrus.Errorf("Pipeline %s failed: %v", pipelineID, err)
			status.Add(pipelineID, "failed", err.Error())
		} else {
			status.Add(pipelineID, "success", "")
		}
	}()
	ctx.JSON(server.StatusOK, server.Generalesponse{
		"message": fmt.Sprintf("Pipeline %s started", pipelineID),
	})

}
