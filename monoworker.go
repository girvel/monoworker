package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var results []string  // TODO random delay+hashmap
var lastId = -1

func say_hello(target string) {
    time.Sleep(time.Second * 60)
    results = append(results, fmt.Sprintf("Hello, %s!", target))
}

func worker(in chan string) {
    for {
        go say_hello(<-in)
    }
}

type CreateRequest struct {
    Target string `json:"target"`
}

func main() {
    in := make(chan string)
    go worker(in)

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

        in <- json.Target  // TODO may be blocking; use buffering+select
        lastId += 1
        c.JSON(http.StatusOK, gin.H {
            "id": lastId,
        })
    })

    g.GET("/task", func (c *gin.Context) {
        c.JSON(http.StatusOK, gin.H {
            "ready": len(results),
            "in_progress": lastId - len(results) + 1,
        })
    })

    g.GET("/task/:id", func (c *gin.Context) {
        id, err := strconv.Atoi(c.Param("id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H {"error": "id should be a number"})
            return
        }

        var status string
        switch {
        case id > lastId:
            status = "non_existent"
        case id >= len(results):
            status = "in_progress"
        default:
            status = "ready"
        }

        c.JSON(http.StatusOK, gin.H {
            "status": status,
        })
    })

    g.GET("/result/:id", func(c *gin.Context) {
        id, err := strconv.Atoi(c.Param("id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H {"error": "id should be a number"})
            return
        }

        if id >= len(results) {
            c.JSON(http.StatusBadRequest, gin.H {
                "error": fmt.Sprintf("no result with id %d", id),
            })
            return
        }

        c.JSON(http.StatusOK, gin.H {
            "result": results[id],
        })
    })

    g.Run()
}
