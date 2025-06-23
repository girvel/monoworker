package monoworker

import (
	"testing"
	"time"
)

func TestAsynchronicity(t *testing.T) {
    worker := NewWorker(func(_ string) string {
        time.Sleep(time.Millisecond * 10)
        return ""
    }, Config{MaxBufferedTasks: 2})

    for i := range 2 {
        if _, ok := worker.CreateTask(""); !ok {
            t.Errorf("Attempt %d to create task failed", i)
        }
    }

    go worker.Run()
    time.Sleep(time.Millisecond * 15)

    if ready := worker.GetStats().Ready; ready != 2 {
        t.Errorf("Expected 2 tasks to finish, got %d", ready)
    }
}
