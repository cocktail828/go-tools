# go-tools
A set of useful go tools/packages.

## 项目结构
``` plainText
go-tools/
├── algo/          # 算法相关模块
├── configor/      # 配置管理模块
├── exp/           # 实验性模块
├── pkg/           # 公共包
├── protocol/      # 协议定义
├── tools/         # 工具类
├── xlog/          # 日志模块
└── z/             # 通用工具集
```

## 模块说明
### algo/

算法相关模块，提供多种常用算法实现。

- **balancer/**: 负载均衡器实现
  - 支持随机、轮询、权重轮询等策略
  - 提供节点健康检查和故障转移机制

- **btree/**: B树实现
  - 支持泛型的B树数据结构
  - 提供高效的插入、删除和查询操作

- **cm4/**: 计数算法实现

- **gcache/**: 缓存实现
  - 支持LRU、LFU、ARC等多种缓存算法
  - 提供缓存统计和自动加载功能

- **hash/**: 哈希算法实现
  - 包含murmur3、xxhash等高性能哈希算法

- **hashring/**: 一致性哈希环实现
  - 支持节点的添加、删除和查找

- **mathx/**: 数学相关工具

- **pool/**: 对象池实现
  - 提供资源的复用和管理

- **queue/**: 队列实现
  - 优先级队列
  - 环形列表

- **rolling/**: 滚动窗口实现
  - 支持统计一段时间内的数据

- **snowflake/**: 雪花算法实现
  - 生成唯一ID

- **topn/**: TopN算法实现
  - 找出最大的N个元素

### configor/

配置管理模块，支持从环境变量加载配置。

- 支持结构体字段与环境变量的绑定
- 支持默认值设置
- 支持多种类型的解析（字符串、整数、浮点数、时间等）
- 支持自定义解析器和校验器

### exp/

实验性模块，包含一些新功能的尝试。

- **healthy/**: 健康检查相关
- **hystrix/**: 熔断机制实现

### pkg/

公共包，提供各种通用功能。

- **encoding/**: 编码实现
  - AES加密/解密
  - Base58、Base62、Base64等编码
  - 数据扁平化

- **mapset/**: 集合实现
  - 支持线程安全和非线程安全版本
  - 提供集合的各种操作（添加、删除、交集、并集等）

- **mapx/**: 并发映射实现
  - 提供高性能的并发安全映射

- **nacs/**: 服务发现相关
  - 支持etcd、nacos等服务发现机制
  - 提供服务的注册和发现功能

- **netx/**: 网络相关工具
  - 下载器
  - HTTP请求和响应处理

- **retry/**: 重试机制实现
  - 支持多种重试策略
  - 可配置重试间隔和最大重试次数

- **tries/**: Trie树实现
  - 支持前缀匹配和路由查找

### protocol/

协议定义模块，包含Protocol Buffers定义文件。

- **message.proto**: 消息定义

### tools/

工具类模块。

- **errorify/**: 错误处理工具

### xlog/

日志模块，提供日志记录和滚动功能。

- 支持不同日志级别（Debug, Info, Warn, Error, Fatal）
- 支持日志文件的滚动和压缩
- 支持彩色日志输出

### z/

通用工具集，提供各种辅助功能。

- **chain/**: 拦截器链实现
- **environ/**: 环境变量工具
- **lock.go**: 锁相关实现
- **memory.go**: 内存相关工具
- **reflectx/**: 反射相关工具
- **runnable/**: 可运行任务相关
  - 弹性作业
  - 优雅关闭
- **stringx/**: 字符串相关工具
  - 随机字符串生成
  - 字符串转换
  - 版本比较
- **timex/**: 时间相关工具
  - 时间跨度计算
  - 时间格式化

## 安装

```bash
go get github.com/yourusername/go-tools
```

## 使用示例

### 负载均衡器

```go
import "github.com/yourusername/go-tools/algo/balancer"

// 创建节点
nodes := []balancer.Node{
    &yourNodeImpl{value: "node1", weight: 1},
    &yourNodeImpl{value: "node2", weight: 2},
}

// 创建权重轮询负载均衡器
lb := balancer.NewWeightRoundRobin(nodes)

// 选择节点
node := lb.Pick()
```

### 配置管理

```go
import "github.com/yourusername/go-tools/configor"

type Config struct {
    Port     int    `env:"PORT" default:"8080"`
    Database string `env:"DATABASE" default:"sqlite3"`
}

var cfg Config
c := &configor.Configor{
    LoadEnv:   true,
    EnvPrefix: "APP",
}

if err := c.bindEnv(&cfg); err != nil {
    log.Fatal(err)
}
```

### 日志

```go
import "github.com/yourusername/go-tools/xlog"

// 创建日志记录器
logger := xlog.NewLogger(xlog.LoggerConfig{
    Level:      xlog.LevelInfo,
    Filename:   "/var/log/app.log",
    MaxSize:    100, // MB
    MaxBackups: 5,
    MaxAge:     30, // days
    Compress:   true,
})

// 记录日志
logger.Info("Application started")
logger.Errorf("Error occurred: %v", err)
```

## 测试

运行所有测试：

```bash
go test ./...
```

运行特定模块的测试：

```bash
go test ./algo/balancer
```

## 许可证
MIT License

Copyright (c) 2023 Your Name

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.