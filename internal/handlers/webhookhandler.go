package handlers

import (
	"fmt"

	"github.com/khaledibrahim1015/goFlow-cicd/internal/config"
	"github.com/khaledibrahim1015/goFlow-cicd/internal/git"
	"github.com/khaledibrahim1015/goFlow-cicd/internal/server"
)

func WebHookHandler(ctx *server.HttpContext) {
	if ctx.Request.Method != "POST" {
		ctx.JSON(server.StatusMethodNotAllowed, server.Generalesponse{
			"error":   server.ResponseMessage["invalid_method"],
			"message": server.StatusCodeText[server.StatusMethodNotAllowed],
		})
		return
	}
	var cfg *config.PipelineConfig
	cfg, err := config.Load()
	if err != nil {
		ctx.JSON(server.StatusInternalServerError, server.Generalesponse{
			"error":   fmt.Sprintf("Config load failed: %v", err),
			"message": server.StatusCodeText[server.StatusInternalServerError],
		})
		return
	}
	provider := git.DetermineGitProvider(ctx.Request)
	if provider == "unknown" {
		ctx.JSON(server.StatusBadRequest, server.Generalesponse{
			"error":   fmt.Sprintf("unsupported git provider: %v", err),
			"message": server.StatusCodeText[server.StatusBadRequest],
		})
		return
	}
	switch provider {
	case git.Github:
		git.Githubhandler(ctx, cfg)
		return
	case git.Gitlab:
		git.GitLabhandler(ctx, cfg)
	}

}
