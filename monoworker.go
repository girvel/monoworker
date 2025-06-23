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

// TODO generic over string?
type Worker struct {
    results map[int]string
    resultsLock sync.Mutex
    lastId int
    lastIdLock sync.Mutex
    in chan Task
}

func NewWorker() *Worker {
    return &Worker{
        results: make(map[int]string),
        lastId: -1,
        in: make(chan Task, 1024),
    }
}

func (w *Worker) Run() {
    for {
        go w.executeTask(<-w.in)
    }
}

func (w *Worker) executeTask(task Task) {
    result := say_hello(task)

    w.resultsLock.Lock()
    defer w.resultsLock.Unlock()
    w.results[task.id] = result
}

func (w *Worker) CreateTask(target string) (int, bool) {
    w.lastIdLock.Lock()
    defer w.lastIdLock.Unlock()

    select {
    case w.in <- Task{target, w.lastId + 1}:
        w.lastId++
        return w.lastId, true
    default:
        return -1, false
    }
}

type TaskStatus string

const (
    Ready       TaskStatus = "ready"
    InProgress  TaskStatus = "in_progress"
    NonExistent TaskStatus = "non_existent"
)

func (w Worker) GetTaskStatus(id int) TaskStatus {
    if id > w.lastId {
        return "non_existent"
    } else if _, exists := w.results[id]; exists {
        return "ready"
    } else {
        return "in_progress"
    }
}

func (w Worker) GetStats() map[string]int {
    return map[string]int{
        "ready": len(w.results),
        "in_progress": w.lastId - len(w.results) + 1,
    }
}

func (w Worker) GetTaskResult(id int) (string, bool) {
    result, exists := w.results[id]
    return result, exists
}

func say_hello(task Task) string {
    time.Sleep(time.Second * time.Duration(50 + rand.IntN(20)))
    return fmt.Sprintf("Hello, %s!", task.target)
}

type CreateRequest struct {
    Target string `json:"target"`
}

func main() {
    worker := NewWorker()
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
