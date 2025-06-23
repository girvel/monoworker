package monoworker

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CreateRequest struct {
    Target string `json:"target"`
}

func RunAPI(worker *Worker[string, string]) {
    slog.Info("Creating API")
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

    slog.Info("Running API")
    g.Run()
}
