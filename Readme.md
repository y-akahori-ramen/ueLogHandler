# ueLogHandler

## Usage

### Parse Unreal Engine log format

```go
package main

import (
    "fmt"

    ueloghandler "github.com/y-akahori-ramen/ueLogHandler"
)

func main() {
    var log ueloghandler.Log
    log = ueloghandler.NewLog("[2022.05.21-13.38.22:383][  0]LogConfig: Applying CVar settings from Section [/Script/Engine.RendererSettings] File [Engine]")
    fmt.Printf("Time:%#v\nFrame:%#v\nCategory:%#v\nVerbosity:%#v\nLog:%#v\n", log.Time, log.Frame, log.Category, log.Verbosity, log.Log)
    // Output:
    // Time:"2022.05.21-13.38.22:383"
    // Frame:"  0"
    // Category:"LogConfig"
    // Verbosity:""
    // Log:"[2022.05.21-13.38.22:383][  0]LogConfig: Applying CVar settings from Section [/Script/Engine.RendererSettings] File [Engine]"

    log = ueloghandler.NewLog("[2022.05.21-15.58.22:810][843]LogWindows: Error: [Callstack] 0x00007ffe5bc37eef UnrealEditor-Core.dll!UnknownFunction []")
    fmt.Printf("Time:%#v\nFrame:%#v\nCategory:%#v\nVerbosity:%#v\nLog:%#v\n", log.Time, log.Frame, log.Category, log.Verbosity, log.Log)
    // Output:
    // Time:"2022.05.21-15.58.22:810"
    // Frame:"843"
    // Category:"LogWindows"
    // Verbosity:"Error"
    // Log:"[2022.05.21-15.58.22:810][843]LogWindows: Error: [Callstack] 0x00007ffe5bc37eef UnrealEditor-Core.dll!UnknownFunction []"
}
```

### Watch log file

Watch Unreal Engine log file and handle the log as a structured log for each update.

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
    logHandler := ueloghandler.NewLogHandler(func(log ueloghandler.Log) error {
        fmt.Printf("%#v\n", log)
        return nil
    })

    watcherLogHandler := ueloghandler.NewWatcherLogHandler(func(log ueloghandler.WatcherLog) error {
        fmt.Printf("Log:%#v LogFileOpenAt:%s\n", log.LogData, log.FileOpenTime)
        return nil
    })

    wacher.AddLogHandler(logHandler)
    wacher.AddWatcherLogHandler(watcherLogHandler)

    wacher.Watch(ctx, `ue.log`, time.Millisecond*500)
}
```
