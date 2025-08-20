package services

import (
	"github.com/gin-gonic/gin"
)

type Route struct {
	Path    string
	Method  string
	Handler func(c *gin.Context)
}

type Routes []Route
