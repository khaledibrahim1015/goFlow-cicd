package handlers

import (
	"fmt"

	"github.com/khaledibrahim1015/goFlow-cicd/internal/config"
	"github.com/khaledibrahim1015/goFlow-cicd/internal/git"
	"github.com/khaledibrahim1015/goFlow-cicd/internal/server"
	"github.com/sirupsen/logrus"
)

func WebHookHandlerWithConfig(ctx *server.HttpContext, cfg *config.PipelineConfig) {
	if ctx.Request.Method != "POST" {
		ctx.JSON(server.StatusMethodNotAllowed, server.Generalesponse{
			"error":   server.ResponseMessage["invalid_method"],
			"message": server.StatusCodeText[server.StatusMethodNotAllowed],
		})
		return
	}

	provider := git.DetermineGitProvider(ctx.Request)
	if provider == "unknown" {
		ctx.JSON(server.StatusBadRequest, server.Generalesponse{
			"error":   fmt.Sprintf("unsupported git provider :%v  ", provider),
			"message": server.StatusCodeText[server.StatusBadRequest],
		})
		return
	}
	logrus.Infof("current provider request :%v", provider)

	switch provider {
	case git.Github:
		git.Githubhandler(ctx, cfg)
		return
	case git.Gitlab:
		git.GitLabhandler(ctx, cfg)
	}

}
