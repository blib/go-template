package services

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type Route struct {
	Path    string
	Method  string
	Handler func(c *gin.Context)
}

type RoutesOutParams struct {
	fx.Out
	Routes []Route `group:"routes"`
}
