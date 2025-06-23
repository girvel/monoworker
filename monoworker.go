package main

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	monoworker "github.com/girvel/monoworker/src"
)

func sayHello(input string) string {
    time.Sleep(time.Second * time.Duration(50 + rand.IntN(20)))
    return fmt.Sprintf("Hello, %s!", input)
}

func main() {
    slog.Info("Starting monoworker")
    worker := monoworker.NewWorker(sayHello, monoworker.Config{
        MaxBufferedTasks: 5,
        MaxActiveTasks: 10,
    })

    go worker.Run()

    api := monoworker.BuildAPI(worker)
    slog.Info("Running API")
    api.Run()
}
