package monoworker

import (
    "fmt"
    "time"
    "math/rand/v2"
    "sync"
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
