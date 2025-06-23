package monoworker

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type createRequest struct {
    Target string `json:"target"`
}

// Build Gin-based HTTP API for the worker
func BuildAPI(worker *Worker[string, string]) *gin.Engine {
    slog.Info("Creating API")
    g := gin.Default()

    g.GET("/ping", func (c *gin.Context) {
        c.JSON(http.StatusOK, gin.H {
            "result": "pong",
        })
    })

    g.POST("/task", func (c *gin.Context) {
        var json createRequest
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

        response := gin.H {
            "status": worker.GetTaskStatus(id),
        }

        if response["status"] == Ready {
            result, _ := worker.GetTaskResult(id)
            response["result"] = result.result
            response["created"] = result.created
            response["duration_sec"] = result.duration.Seconds()
        }

        c.JSON(http.StatusOK, response)
    })

    return g
}
