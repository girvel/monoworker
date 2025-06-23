package main

import (
	"fmt"
	"math/rand/v2"
	"time"

	monoworker "github.com/girvel/monoworker/src"
)

func say_hello(input string) string {
    time.Sleep(time.Second * time.Duration(50 + rand.IntN(20)))
    return fmt.Sprintf("Hello, %s!", input)
}

func main() {
    worker := monoworker.NewWorker(say_hello)
    go worker.Run()
    monoworker.RunAPI(worker)
}
