# configor

轻量级 Go 配置加载库，支持多种数据格式、环境变量覆盖和 default tag 默认值。

## 特性

- 支持 TOML、HCL2、JSON 等任意 Unmarshaller
- 支持通过 struct tag 设置默认值 (`default`)
- 支持环境变量覆盖 (`env`)，可配置前缀
- 支持 `validate` tag 校验（基于 go-playground/validator）
- 支持指针字段和嵌套结构体
- 支持 `time.Duration` 类型

## 优先级

配置值的加载优先级（高到低）：

```
环境变量 (env) > 配置文件 (file) > 默认值 (default tag)
```

即：如果环境变量存在，则覆盖文件中的值；文件中的值覆盖 default tag 的值。

## 安装

```bash
go get github.com/cocktail828/go-tools/configor
```

## 快速开始

```go
package main

import (
    "fmt"
    "time"

    "github.com/cocktail828/go-tools/configor"
)

type DBConfig struct {
    Host     string        `env:"DB_HOST" default:"localhost" toml:"host"`
    Port     int           `env:"DB_PORT" default:"3306" toml:"port"`
    User     string        `env:"DB_USER" default:"root" toml:"user"`
    Password string        `env:"DB_PASSWORD" toml:"password"`
    Timeout  time.Duration `env:"DB_TIMEOUT" default:"5s" toml:"timeout"`
}

type AppConfig struct {
    Name string   `env:"APP_NAME" default:"myapp" toml:"name"`
    DB   DBConfig `toml:"db"`
}

func main() {
    tomlData := []byte(`
name = "awesome-app"

[db]
host = "db.example.com"
port = 5432
user = "admin"
password = "secret"
`)

    var cfg AppConfig
    if err := configor.Load(&cfg, tomlData); err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", cfg)
}
```

运行时如果设置了 `DB_PORT=9090`，则 `cfg.DB.Port` 为 `9090`（环境变量覆盖文件值）。

## Struct Tag

| Tag       | 说明                         | 示例                        |
|-----------|-----------------------------|-----------------------------|
| `env`     | 绑定的环境变量名             | `env:"DB_HOST"`            |
| `default` | 无文件值且无环境变量时的默认值 | `default:"localhost"`      |
| `validate`| 校验规则（go-playground）    | `validate:"required"`      |

`env` 设为 `"-"` 表示忽略该字段的环境变量绑定。

## 环境变量前缀

```go
c := &configor.Configor{
    LoadEnv:      true,
    EnvPrefix:    "MYAPP",
    Unmarshaller: toml.Unmarshal,
    Validator:    validator.New().Struct,
}

// 此时 env:"DB_HOST" 对应环境变量 MYAPP_DB_HOST
var cfg AppConfig
err := c.Load(&cfg, tomlData)
```

## 多格式混合加载

使用 `LoadWithUnmarshaller` 可以一次加载多个不同格式的配置片段：

```go
import (
    "encoding/json"
    "github.com/BurntSushi/toml"
    "github.com/cocktail828/go-tools/configor"
)

err := configor.LoadWithUnmarshaller(&cfg,
    configor.Pair{Data: tomlBytes, Unmarshaller: toml.Unmarshal},
    configor.Pair{Data: jsonBytes, Unmarshaller: json.Unmarshal},
)
```

后加载的文件内容会覆盖先加载的同名字段。

## 支持的字段类型

- `string`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `bool`
- `time.Duration`（如 `"5s"`, `"100ms"`, `"2h30m"`）
- 指针类型（`*string`, `*int` 等）
- 嵌套结构体

## 自定义 Configor

```go
c := &configor.Configor{
    LoadEnv:      true,             // 开启环境变量读取
    EnvPrefix:    "APP",            // 环境变量前缀
    Unmarshaller: toml.Unmarshal,   // 解析器
    Validator:    validator.New().Struct, // 校验器（可为 nil 跳过校验）
}
err := c.Load(&cfg, fileBytes)
```
