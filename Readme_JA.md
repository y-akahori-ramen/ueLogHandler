# ueLogHandler
GoでUnrealEngine形式のログを処理するためのユーティリティです。

[English Version](./Readme.md)

## 機能
- UnrealEngine形式のログ構文解析
- UnrealEngineのログファイル監視
- 構造化ログの出力とハンドリング

## UnrealEngine形式のログ構文解析
UnrealEngine形式のログを解析して以下の情報を取得することができます。
- ログ出力時刻
- ログ出力フレーム
- ログカテゴリ
- Verbosity

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

## UnrealEngineのログファイル監視
UnrealEngineのログを監視して更新ごとにログをハンドリングすることができます。

```go
package main

import (
    "context"
    "fmt"
    "log"
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

    // Create Log handler
    logHandler := ueloghandler.NewLogHandler(func(log ueloghandler.Log) error {
        // This handler is called when the log file is updated
        _, err := fmt.Printf("%#v\n", log)
        return err
    })

    // Create file notifier
    pathToLogFile := "ue.log"
    watchInterval := time.Millisecond * 500
    fileNotifier := ueloghandler.NewFileNotifier(pathToLogFile, watchInterval)

    // Start watching log
    wacher := ueloghandler.NewWatcher()
    wacher.AddLogHandler(logHandler)
    err := wacher.Watch(ctx, fileNotifier)
    if err != nil {
        log.Fatal(err)
    }
}
```

## 構造化ログの出力とハンドリング
構造化ログフォーマットをスキーマファイルで定義でき、スキーマファイルから以下のソースコードを生成することができます。

- 構造化ログをUnrealEngineのログとして出力するためのユーティリティ
- 構造化ログのログハンドラ

生成されたソースコードを使用することで構造化ログ出力とハンドリングを行うことができます。

### 使い方
#### 1.構造化ログフォーマットの定義
構造化ログのフォーマットをYAMLで定義します。

```yaml
structures:
  list:
    Sample:
      Meta:
        Tag: DataTag
        Insert: False
        Value: 1.23
        Value2: 123
      Body:
        Damage: int32
        Name: string
        Position: vector3
    Sample2:
      Meta:
        Insert: True
      Body:
        Count: int32
```

#### 2.ソースコード生成
ソースコードジェネレータをビルドし、スキーマファイルを渡してソースコードを生成します。

```
cd cmds/structuregen
go build -o gen
```

```
./gen -cpp-namespace structuredLog -cpp-out sample.h -go-package main -go-out sample.go -src structure.yaml
```

実行すると `sample.h` と `sample.go` が出力されます。

sample.h
```cpp
// Code generated by structuregen. DO NOT EDIT.
#pragma once
#include "CoreMinimal.h"
namespace structuredLog {

FString LogSample(int32 Damage,const FString& Name,const FVector& Position)
{
	return FString::Printf(TEXT(R"(_BEGIN_STRUCTURED_{"Body":{"Body":{"Damage":%d,"Name":"%s","Position":{"X":%f,"Y":%f,"Z":%f}},"Meta":{"Insert":false,"Tag":"DataTag","Value":1.23,"Value2":123}},"Meta":{"Type":"Sample"}}_END_STRUCTURED_)"),Damage,*Name,Position.X,Position.Y,Position.Z);
}

FString LogSample2(int32 Count)
{
	return FString::Printf(TEXT(R"(_BEGIN_STRUCTURED_{"Body":{"Body":{"Count":%d},"Meta":{"Insert":true}},"Meta":{"Type":"Sample2"}}_END_STRUCTURED_)"),Count);
}

}
```

sample.go
```go
// Code generated by structuregen. DO NOT EDIT.

package main

import ueloghandler "github.com/y-akahori-ramen/ueLogHandler"

type SampleMeta struct {
	Tag    string
	Insert bool
	Value  float64
	Value2 int32
}
type SampleBody struct {
	Name     string
	Position ueloghandler.FVector
	Damage   int32
}
type SampleData ueloghandler.TStructuredData[SampleMeta, SampleBody]
type SampleHandlerFunc func(SampleData, ueloghandler.Log) error
type SampleLogHandler struct {
	f SampleHandlerFunc
}

func (h *SampleLogHandler) Type() string {
	return "Sample"
}
func (h *SampleLogHandler) Handle(json string, log ueloghandler.Log) error {
	data, err := ueloghandler.JSONToStructuredData[SampleMeta, SampleBody](json)
	if err != nil {
		return err
	}
	return h.f(SampleData(data), log)
}
func NewSampleLogHandler(f SampleHandlerFunc) ueloghandler.StructuredLogDataHandler {
	return &SampleLogHandler{f: f}
}

type Sample2Meta struct {
	Insert bool
}
type Sample2Body struct {
	Count int32
}
type Sample2Data ueloghandler.TStructuredData[Sample2Meta, Sample2Body]
type Sample2HandlerFunc func(Sample2Data, ueloghandler.Log) error
type Sample2LogHandler struct {
	f Sample2HandlerFunc
}

func (h *Sample2LogHandler) Type() string {
	return "Sample2"
}
func (h *Sample2LogHandler) Handle(json string, log ueloghandler.Log) error {
	data, err := ueloghandler.JSONToStructuredData[Sample2Meta, Sample2Body](json)
	if err != nil {
		return err
	}
	return h.f(Sample2Data(data), log)
}
func NewSample2LogHandler(f Sample2HandlerFunc) ueloghandler.StructuredLogDataHandler {
	return &Sample2LogHandler{f: f}
}
```

#### 3.構造化ログの出力
生成された `sample.h` には構造化ログ文字列を作成する関数が定義されています。  
この関数を使用し構造化ログを出力します。

```cpp
#include "sample.h"

void SampleFunc()
{
    // Log output using the generated structured log string output function
    UE_LOG(LogTemp, Log, TEXT("%s"), *structuredLog::LogSample(10, TEXT("SampleActor"), FVector(1,2,3)));
    UE_LOG(LogTemp, Log, TEXT("%s"), *structuredLog::LogSample2(20));
}
```

#### 4.構造化ログのハンドリング
生成された `sample.go` には構造化ログのハンドラが定義されています。  
このハンドラをWatcherに登録することで構造化ログをハンドリングすることができます。

```go
package main

import (
	"context"
	"fmt"
	"log"
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

	// Create Sample structured log handler
	sampleLogHandler := NewSampleLogHandler(func(data SampleData, l ueloghandler.Log) error {
		_, err := fmt.Printf("Handle SampleLog Meta:%#v Body:%#v\n", data.Meta, data.Body)
		return err
	})

	// Create Sample2 structured log handler
	sample2LogHandler := NewSample2LogHandler(func(data Sample2Data, l ueloghandler.Log) error {
		_, err := fmt.Printf("Handle Sample2Log Meta:%#v Body:%#v\n", data.Meta, data.Body)
		return err
	})

	// Create structured log handler
	structuredLogHandler := ueloghandler.NewStructuredLogHandler()
	structuredLogHandler.AddHandler(sampleLogHandler)
	structuredLogHandler.AddHandler(sample2LogHandler)

	// Create file notifier
	pathToLogFile := "ue.log"
	watchInterval := time.Millisecond * 500
	fileNotifier := ueloghandler.NewFileNotifier(pathToLogFile, watchInterval)

	// Start watching log
	wacher := ueloghandler.NewWatcher()
	wacher.AddLogHandler(structuredLogHandler)
	err := wacher.Watch(ctx, fileNotifier)
	if err != nil {
		log.Fatal(err)
	}
}
```

### スキーマファイル詳細

#### フォーマット
スキーマファイルのフォーマットは以下のCUEファイルで定義されています。  
[schema.cue](./gen/schema.cue)

#### MetaとBody
Metaはスキーマファイルで指定された値が常に出力されます。  
MetaはGo言語の構造化ログハンドラ側でどのようにログを扱うかの情報を伝えるために使用します。  
例えば構造化ログをデータベースに登録する場合の登録先テーブル名等です。

BodyはUnrealEngineのログ出力時に指定された値が出力されます。  
例えばダメージを受けたキャラクターの場所等です。
