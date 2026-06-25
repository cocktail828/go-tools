# xlog

Go 日志工具包，包含两个核心组件：

- **Logger**（lumberjack）— 日志文件自动轮转的 `io.WriteCloser`
- **colorful** — 带颜色和级别过滤的终端日志输出

## 安装

```bash
go get github.com/cocktail828/go-tools/xlog
```

## 日志级别

```go
const (
    LevelDebug  // -1
    LevelInfo   // 0
    LevelWarn   // 1
    LevelError  // 2
    LevelFatal  // 3
)
```

级别越高越严重。设置 level 后，低于该级别的日志不会输出。

---

## Logger（日志轮转）

基于 lumberjack 实现的日志文件 writer，支持按大小自动轮转、保留备份数、按天过期清理、gzip 压缩。

### 基本用法

```go
import "github.com/cocktail828/go-tools/xlog"

w := &xlog.Logger{
    Filename:   "/var/log/app/server.log",
    MaxSize:    100,  // MB，默认 100
    MaxAge:     7,    // 保留天数，0 表示不按时间清理
    MaxBackups: 5,    // 保留备份数，0 表示不按数量清理
    Compress:   true, // 对轮转后的文件 gzip 压缩
    BufSize:    10,   // 写缓冲区大小（MB），0 表示不缓冲
}
defer w.Close()

// 作为 io.Writer 使用
w.Write([]byte("hello\n"))

// 配合标准库 log
log.SetOutput(w)

// 配合 slog
slog.SetDefault(slog.New(slog.NewJSONHandler(w, nil)))
```

### 配置字段

| 字段       | 类型   | 默认值 | 说明                                      |
|-----------|--------|--------|------------------------------------------|
| Filename  | string | 临时目录 | 日志文件路径                              |
| MaxSize   | int    | 100    | 单文件最大 MB，超过后轮转                   |
| MaxAge    | int    | 0      | 备份保留天数，0 不限                       |
| MaxBackups| int    | 0      | 备份保留数量，0 不限                       |
| Compress  | bool   | false  | 是否 gzip 压缩已轮转文件                   |
| BufSize   | int    | 0      | 写缓冲区 MB，0 表示直接写磁盘              |
| Level     | string | error  | 预留字段，供外部日志框架使用                |
| Verbose   | bool   | false  | 预留字段，供外部日志框架使用                |

### 轮转规则

- 当前写入会导致文件超过 `MaxSize` 时，关闭当前文件并重命名为 `name-2006-01-02T15-04-05.000.ext`
- 异步清理旧文件（按 MaxBackups 和 MaxAge）
- 支持手动触发轮转：`w.Rotate()`

---

## colorful（彩色终端日志）

基于标准库 `log.Logger` 封装，提供带颜色的分级日志输出。

### 基本用法

```go
import (
    "github.com/cocktail828/go-tools/xlog"
    "github.com/cocktail828/go-tools/xlog/colorful"
)

// 使用默认 logger（输出到 stderr）
colorful.SetLevel(xlog.LevelDebug)
colorful.SetPrefix("[app] ")
colorful.SetFlags(colorful.LstdFlags | colorful.Lshortfile)

colorful.Debugf("connecting to %s", addr)
colorful.Infof("server started on port %d", port)
colorful.Warnf("slow query: %v", duration)
colorful.Errorf("request failed: %v", err)
```

### 创建自定义 logger

```go
import "os"

l := colorful.NewColorful(os.Stdout, "[myapp] ", colorful.LstdFlags|colorful.Lshortfile)
l.SetLevel(xlog.LevelInfo)

l.Info("ready")
l.Errorf("something went wrong: %v", err)
```

### 级别对应颜色

| 级别   | 颜色            |
|--------|----------------|
| Debug  | 绿色斜体       |
| Info   | 默认           |
| Warn   | 黄色           |
| Error  | 红色           |
| Fatal  | 红色加粗       |

### 自定义颜色

```go
import "github.com/fatih/color"

l.SetColor(xlog.LevelWarn, color.New(color.FgMagenta, color.Bold))
```

### 禁用/启用颜色

```go
// 禁用所有级别颜色
l.DisableColor()

// 仅禁用 Debug 级别颜色
l.DisableColor(xlog.LevelDebug)

// 重新启用
l.EnableColor()
```

### 限流日志

当相同日志大量输出时，使用 `Limited` 限制输出频率，避免日志风暴：

```go
import "golang.org/x/time/rate"

// 每秒最多输出 1 条，被抑制的日志会在下次输出时汇报数量
limited := l.Limited(rate.NewLimiter(rate.Every(time.Second), 1))
for {
    limited.Errorf("connection refused: %v", err)
    time.Sleep(10 * time.Millisecond)
}
// 输出：
// Error connection refused: ...
// (1s 后)
// Error connection refused: ...
// Error about 96 lines of log has been supressed...
```

---

## Printer 接口

`xlog` 定义了一个简单的 `Printer` 接口，可用于依赖注入：

```go
type Printer interface {
    Printf(format string, v ...any)
}
```

不需要日志时可使用 `xlog.NopPrinter{}`。
