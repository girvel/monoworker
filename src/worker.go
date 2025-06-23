package monoworker

import (
	"log/slog"
	"sync"
	"os"
	"os/signal"
)

type Task[T any] struct {
    target T
    id int
}

type Worker[In any, Out any] struct {
    results map[int]Out
    resultsLock sync.Mutex
    lastId int
    lastIdLock sync.Mutex
    in chan Task[In]
    process func(In) Out
}

func NewWorker[In any, Out any](process func(In) Out) *Worker[In, Out] {
    return &Worker[In, Out]{
        results: make(map[int]Out),
        lastId: -1,
        in: make(chan Task[In], 1024),
        process: process,
    }
}

func (w *Worker[In, Out]) Run() {
    slog.Info("Taking control of interrupts")
    go w.handleInterrupts()

    slog.Info("Starting the worker")
    for {
        go w.executeTask(<-w.in)
    }
}

func (w *Worker[In, Out]) handleInterrupts() {
    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt)
    for {
        <-interrupt
        in_progress := w.GetStats().InProgress
        if in_progress == 0 {
            slog.Info("SIGINT, shutting down gracefully...")
            os.Exit(0)
        }

        slog.Warn("SIGINT attempt", "tasks_in_progress", in_progress)
    }
}

func (w *Worker[In, Out]) executeTask(task Task[In]) {
    result := w.process(task.target)

    w.resultsLock.Lock()
    defer w.resultsLock.Unlock()
    w.results[task.id] = result
}

func (w *Worker[In, Out]) CreateTask(target In) (int, bool) {
    w.lastIdLock.Lock()
    defer w.lastIdLock.Unlock()

    select {
    case w.in <- Task[In]{target, w.lastId + 1}:
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

func (w Worker[In, Out]) GetTaskStatus(id int) TaskStatus {
    if id > w.lastId {
        return NonExistent
    } else if _, exists := w.results[id]; exists {
        return Ready
    } else {
        return InProgress
    }
}

type Stats struct {
    Ready int `json:"ready"`
    InProgress int `json:"in_progress"`
}

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

func (w Worker[In, Out]) GetTaskResult(id int) (Out, bool) {
    result, exists := w.results[id]
    return result, exists
}
