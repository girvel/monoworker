package main

import (
	"fmt"
	"net/http"
	"strconv"
    "time"
    "math/rand/v2"

	"github.com/gin-gonic/gin"
	monoworker "github.com/girvel/monoworker/src"
)

func say_hello(input string) string {
    time.Sleep(time.Second * time.Duration(50 + rand.IntN(20)))
    return fmt.Sprintf("Hello, %s!", input)
}

type CreateRequest struct {
    Target string `json:"target"`
}

func main() {
    worker := monoworker.NewWorker(say_hello)
    go worker.Run()

    g := gin.Default()

    g.GET("/ping", func (c *gin.Context) {
        c.JSON(http.StatusOK, gin.H {
            "result": "pong",
        })
    })

    g.POST("/task", func (c *gin.Context) {
        var json CreateRequest
        if err := c.ShouldBindJSON(&json); err != nil {
            c.JSON(http.StatusBadRequest, gin.H {"error": err.Error()})
            return
        }

        if id, ok := worker.CreateTask(json.Target); ok {
            c.JSON(http.StatusOK, gin.H{"id": id})
        } else {
            c.JSON(http.StatusServiceUnavailable, gin.H {"error": "system busy"})
        }
    })

    g.GET("/task", func (c *gin.Context) {
        c.JSON(http.StatusOK, worker.GetStats())
    })

    g.GET("/task/:id", func (c *gin.Context) {
        id, err := strconv.Atoi(c.Param("id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H {"error": "id should be a number"})
            return
        }

        c.JSON(http.StatusOK, gin.H {
            "status": worker.GetTaskStatus(id),
        })
    })

    g.GET("/result/:id", func(c *gin.Context) {
        id, err := strconv.Atoi(c.Param("id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H {"error": "id should be a number"})
            return
        }

        if result, exists := worker.GetTaskResult(id); exists {
            c.JSON(http.StatusOK, gin.H {
                "result": result,
            })
        } else {
            c.JSON(http.StatusNotFound, gin.H {
                "error": fmt.Sprintf("no result with id %d", id),
            })
        }
    })

    g.Run()
}
