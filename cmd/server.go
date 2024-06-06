package main

import (
	"deeplx/server"
	"fmt"

	"github.com/gin-gonic/gin"
)

func run() {
	gin.SetMode(gin.ReleaseMode)
	g := gin.Default()

	server.Router(g)

	fmt.Println("server start success!!")
	g.Run(":9001")
}
