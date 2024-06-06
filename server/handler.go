package server

import (
	"deeplx/model"
	"fmt"

	"github.com/gin-gonic/gin"
)

func translate(c *gin.Context) {
	auth := c.Query("auth")
	if auth != "123" {
		c.JSON(404, nil)
		return
	}
	cookie := c.Query("cookie")
	if cookie == "" {
		c.JSON(404, nil)
		return
	}
	var req model.TranslateReq
	if err := c.Bind(&req); err != nil {
		c.JSON(404, nil)
		return
	}
	data, err := srv.Translate(req.Text, req.SourceLang, req.TargetLang, cookie)
	if err != nil {
		c.JSON(500, gin.H{
			"err": err,
		})
		return
	}
	c.JSON(200, data)
}

func transport(c *gin.Context) {
	data, err := srv.Transport(c)
	if err != nil {
		fmt.Println(err)
		c.JSON(404, nil)
		return
	}
	c.Status(200)
	c.Header("Content-Type", "application/json")
	c.Writer.Write(data)
}
