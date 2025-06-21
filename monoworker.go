package main

import (
    "fmt"
    "net/http"
    "strconv"
    "time"
    "math/rand/v2"
    "sync"
    "sync/atomic"

    "github.com/gin-gonic/gin"
)

var results map[int]string
var resultsLock sync.Mutex
var lastId atomic.Int32

func say_hello(target string, id int) {
    time.Sleep(time.Second * time.Duration(50 + rand.IntN(20)))
    resultsLock.Lock()
    results[id] = fmt.Sprintf("Hello, %s!", target)
    resultsLock.Unlock()
}

func worker(in chan string) {
    for {
        target := <-in
        lastId.Add(1)
        go say_hello(target, int(lastId.Load()))
    }
}

type CreateRequest struct {
    Target string `json:"target"`
}

func main() {
    results = make(map[int]string)
    lastId.Store(-1)

    in := make(chan string, 1024)
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

        select {
        case in <- json.Target:
            c.JSON(http.StatusOK, gin.H {
                "id": lastId.Load() + 1,
            })
        default:
            c.JSON(http.StatusServiceUnavailable, gin.H {"error": "system busy"})
        }
    })

    g.GET("/task", func (c *gin.Context) {
        c.JSON(http.StatusOK, gin.H {
            "ready": len(results),
            "in_progress": int(lastId.Load()) - len(results) + 1,
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
        case id > int(lastId.Load()):
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
