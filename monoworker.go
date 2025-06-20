package main

import (
    "fmt"
    "net/http"
    "strconv"
    "time"
    "math/rand/v2"
    "sync"

    "github.com/gin-gonic/gin"
)

var results map[int]string
var resultsLock sync.Mutex
var lastId = -1

func say_hello(target string, id int) {
    time.Sleep(time.Second * time.Duration(50 + rand.IntN(20)))
    resultsLock.Lock()
    results[id] = fmt.Sprintf("Hello, %s!", target)
    resultsLock.Unlock()
}

func worker(in chan string) {
    for {
        target := <-in
        lastId += 1
        go say_hello(target, lastId)
    }
}

type CreateRequest struct {
    Target string `json:"target"`
}

func main() {
    results = make(map[int]string)
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
        c.JSON(http.StatusOK, gin.H {
            "id": lastId + 1,
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
