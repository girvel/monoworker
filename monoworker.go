package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func main() {
    g := gin.Default()

    g.GET("/ping", func (c *gin.Context) {
        c.JSON(http.StatusOK, gin.H {
            "result": "pong",
        })
    })

    g.Run()
}
