package main

import (
	"os"

	"github.com/gin-gonic/gin"
	gmw "github.com/waro163/gin-middleware"
)

func main() {
	r := gin.New()
	writer := os.Stdout
	log := &gmw.RequestLog{Output: writer}

	r.Use(gmw.AddRequestID(), log.AddRequestLog(), gmw.Recovery(writer))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/panic", func(ctx *gin.Context) {
		a := 0
		b := 10 / a
		c := []int{a, b}
		ctx.JSON(200, c[99])
		panic("error")
	})

	r.Run(":8081")
}
