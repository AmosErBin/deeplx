package server

import (
	"deeplx/service"

	"github.com/gin-gonic/gin"
)

var srv *service.Service

func Router(g *gin.Engine) {
	srv = service.NewService()

	g.Use(gin.Recovery())
	g.POST("/translate", translate)
	g.POST("/transport", transport)
}
