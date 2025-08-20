package app

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/blib/go-template/services"
)

func NewHealthz() services.RoutesOutParams {
	routes := []services.Route{
		{Method: "GET", Path: "/healthz", Handler: Healthz},
	}
	return services.RoutesOutParams{Routes: routes}
}

func Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
