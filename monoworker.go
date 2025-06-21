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

type Task struct {
    target string
    id int
}

func createTask(in chan Task, target string) (int, bool) {
    lastIdLock.Lock()
    defer lastIdLock.Unlock()

    select {
    case in <- Task{target, lastId + 1}:
        lastId++
        return lastId, true
    default:
        return -1, false
    }
}

var results map[int]string
var resultsLock sync.Mutex
var lastId int = -1
var lastIdLock sync.Mutex

func say_hello(task Task) {
    time.Sleep(time.Second * time.Duration(50 + rand.IntN(20)))
    resultsLock.Lock()
    results[task.id] = fmt.Sprintf("Hello, %s!", task.target)
    resultsLock.Unlock()
}

func worker(in chan Task) {
    for {
        go say_hello(<-in)
    }
}

type CreateRequest struct {
    Target string `json:"target"`
}

func main() {
    results = make(map[int]string)

    in := make(chan Task, 1024)
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

        if id, ok := createTask(in, json.Target); ok {
            c.JSON(http.StatusOK, gin.H{"id": id})
        } else {
            c.JSON(http.StatusServiceUnavailable, gin.H {"error": "system busy"})
        }
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
        if id > lastId {
            status = "non_existent"
        } else if _, exists := results[id]; exists {
            status = "ready"
        } else {
            status = "in_progress"
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

        if result, exists := results[id]; exists {
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
