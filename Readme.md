# ueLogHandler
Watch Unreal Engine log file and handle the log as a structured log for each update.

## Usage

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    ueloghandler "github.com/y-akahori-ramen/ueLogHandler"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigChan
        cancel()
    }()

    wacher := ueloghandler.NewWatcher()
    handler := ueloghandler.NewLogHandler(func(log ueloghandler.Log) error {
        fmt.Printf("%#v\n", log)
        return nil
    })

    wacher.AddHandler(handler)

    wacher.Watch(ctx, "uelog.txt", time.Millisecond*500)
}

```
