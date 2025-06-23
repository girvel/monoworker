package monoworker

import (
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"time"
)

type task[T any] struct {
    target T
    id int
    created time.Time
}

type Result[T any] struct {
    result T
    created time.Time
    duration time.Duration
}

// Configuration for monoworker.Worker
type Config struct {
    MaxBufferedTasks int  // Max number of queued tasks; should be > 0
    MaxActiveTasks int  // Max number of executing tasks; unlimited if 0
}

// Abstract multithreaded asynchronous worker
type Worker[In any, Out any] struct {
    results map[int]Result[Out]
    resultsLock sync.Mutex
    lastId int
    lastIdLock sync.Mutex
    in chan task[In]
    process func(In) Out
    taskSemaphore chan struct{}
}

// Create new worker from a function doing the work
func NewWorker[In any, Out any](process func(In) Out, config Config) *Worker[In, Out] {
    return &Worker[In, Out]{
        results: make(map[int]Result[Out]),
        lastId: -1,
        in: make(chan task[In], config.MaxBufferedTasks),
        process: process,
        taskSemaphore: make(chan struct{}, config.MaxActiveTasks),
    }
}

// Launch the main loop; infinite, ignores SIGINTs if there are tasks in progress
func (w *Worker[In, Out]) Run() {
    slog.Info("Taking control of interrupts")
    go w.handleInterrupts()

    slog.Info("Starting the worker")
    for {
        if cap(w.taskSemaphore) > 0 {
            w.taskSemaphore <- struct{}{}
        }

        go func() {
            w.executeTask(<-w.in)
            if cap(w.taskSemaphore) > 0 {
                <-w.taskSemaphore
            }
        }()
    }
}

func (w *Worker[In, Out]) handleInterrupts() {
    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt)
    for {
        <-interrupt
        inProgress := w.GetStats().InProgress
        if inProgress == 0 {
            slog.Info("SIGINT, shutting down gracefully...")
            os.Exit(0)
        }

        slog.Warn("SIGINT attempt", "tasksInProgress", inProgress)
    }
}

func (w *Worker[In, Out]) executeTask(task task[In]) {
    start := time.Now()
    result := w.process(task.target)
    duration := time.Now().Sub(start)

    w.resultsLock.Lock()
    defer w.resultsLock.Unlock()
    w.results[task.id] = Result[Out]{
        result: result,
        created: task.created,
        duration: duration,
    }
}

// Plan new task; returns -1, false if queue is full
func (w *Worker[In, Out]) CreateTask(target In) (int, bool) {
    w.lastIdLock.Lock()
    defer w.lastIdLock.Unlock()

    select {
    case w.in <- task[In]{target, w.lastId + 1, time.Now()}:
        w.lastId++
        return w.lastId, true
    default:
        slog.Warn("Failed to create task, queue full")
        return -1, false
    }
}

type TaskStatus string

const (
    Ready       TaskStatus = "ready"
    InProgress  TaskStatus = "inProgress"
    NonExistent TaskStatus = "nonExistent"
)

func (w *Worker[In, Out]) GetTaskStatus(id int) TaskStatus {
    w.resultsLock.Lock()
    defer w.resultsLock.Unlock()

    if id > w.lastId {
        return NonExistent
    } else if _, exists := w.results[id]; exists {
        return Ready
    } else {
        return InProgress
    }
}

// General statistics about the worker & its tasks
type Stats struct {
    Ready int `json:"ready"`  // Number of finished tasks
    InProgress int `json:"inProgress"`  // Number of executing/queued tasks
}

// Retrieve general statistics about the worker & its tasks
func (w *Worker[In, Out]) GetStats() Stats {
    // because lastId and len(results) are connected
    w.resultsLock.Lock()
    w.lastIdLock.Lock()
    defer w.resultsLock.Unlock()
    defer w.lastIdLock.Unlock()

    return Stats{
        Ready: len(w.results),
        InProgress: w.lastId - len(w.results) + 1,
    }
}

// Returns false as a second result if task isn't finished/doesn't exist
func (w *Worker[In, Out]) GetTaskResult(id int) (Result[Out], bool) {
    w.resultsLock.Lock()
    defer w.resultsLock.Unlock()

    result, exists := w.results[id]
    return result, exists
}
