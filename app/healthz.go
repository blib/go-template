package app

import (
	"net/http"

	"github.com/blib/go-template/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type outParams struct {
	fx.Out
	Routes services.Routes `group:"routes"`
}

func NewHealthz() outParams {
	routes := services.Routes{
		{Method: "GET", Path: "/healthz", Handler: Healthz},
	}
	return outParams{Routes: routes}
}

func Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
