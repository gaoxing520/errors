# 应用错误处理库（集成日志功能）

[English](../README.MD) | [中文文档](README_zh.md)

`errors` 是一个Go包，提供结构化的应用程序特定错误处理方法，集成了整数错误码、错误链和与 `zerolog` 结构化日志库的无缝集成。

它旨在标准化应用程序中的错误处理，使其更容易：

* 使用预定义代码识别特定错误条件
* 使用错误包装维护上下文和原因
* 使用相关结构化数据（如错误代码和链）一致地记录错误
* 使用通用结构统一成功和错误的API响应

## 功能特性

* **自定义错误类型** (`AppError`)：
    * 为应用程序错误定义标准结构
* **整数错误码**：
    * 每个 `AppError` 包含一个独特的整数代码，对API响应或内部逻辑分支很有用
* **错误链**：
    * 支持包装底层错误 (`WithCause`) 和添加上下文消息 (`With`)，
      保持与 `errors.Is` 和 `errors.As` 兼容的错误链
* **`zerolog` 集成**：
    * 提供便利方法 (`LogError`, `LogWarn`, `LogInfo` 等) 直接记录 `AppError` 实例，
      自动包含代码和完整错误链
    * 线程安全的日志器，使用延迟初始化以获得最佳性能
* **预定义常见错误**：
    * 包含一组常见场景的标准错误代码和消息
* **统一响应结构**：
    * `Error()` 方法格式化后适合直接用于API响应消息和代码。
      对 `Success` 代码进行特殊处理以确保清洁输出

## 安装

使用 `go get` 安装包：

```bash
go get github.com/gaoxing520/errors
go mod tidy
```

## 使用方法

### 基本错误创建

创建带有整数代码的应用程序特定错误：

```go
package main

import (
	"fmt"
	"github.com/gaoxing520/errors"
)

func main() {
	// 创建带有代码和消息的自定义错误
	err := errors.NewAppError(1001, "用户未找到")
	fmt.Printf("错误: %s, 代码: %d\n", err.Error(), err.Code())
	// 输出: 错误: 用户未找到, code=1001, 代码: 1001
}
```

### 使用预定义错误

库提供了几种预定义的错误类型：

```go
// 成功情况 (代码: 0)
successErr := errors.Success
fmt.Println(successErr.Error()) // 输出: "" (成功时为空字符串)

// 未知错误 (代码: -1)
unknownErr := errors.Unknown
fmt.Println(unknownErr.Error()) // 输出: unknown error, code=-1

// 系统错误 (代码: 999999)
systemErr := errors.ErrSystem
fmt.Println(systemErr.Error()) // 输出: system error, code=999999

// HTTP状态码错误
notFoundErr := errors.ErrNotFound
fmt.Println(notFoundErr.Error()) // 输出: not found, code=404
```

### 错误链和上下文

添加上下文和包装底层错误：

```go
// 为错误添加上下文
err := errors.NewAppError(1002, "数据库操作失败")
contextErr := err.With("更新用户 %s 失败", "张三")
fmt.Println(contextErr.Error())
// 输出: 更新用户 张三 失败: 数据库操作失败, code=1002

// 包装底层错误
originalErr := fmt.Errorf("连接超时")
wrappedErr := errors.NewAppError(1003, "服务不可用").WithCause(originalErr)
fmt.Println(wrappedErr.Error())
// 输出: 服务不可用: 连接超时, code=1003
```

### 日志集成

库与zerolog无缝集成进行结构化日志记录：

```go
// 不同级别的错误日志记录
err := errors.NewAppError(1004, "验证失败")

// 在不同级别记录日志
err.LogInfo() // 在INFO级别记录日志
err.LogWarn()  // 在WARN级别记录日志
err.LogError() // 在ERROR级别记录日志
err.LogDebug() // 在DEBUG级别记录日志
err.LogTrace() // 在TRACE级别记录日志

// 致命和恐慌日志记录
err.LogFatal() // 记录日志并退出应用程序
err.LogPanic() // 记录日志并恐慌
```

### 链式操作

您可以链接错误操作以获得简洁的代码：

```go
err := errors.NewAppError(1005, "身份验证失败").
With("用户尝试使用无效凭据登录").
WithCause(fmt.Errorf("密码不匹配")).
LogWarn() // 链式日志记录

// 错误被记录并且仍然可以使用
return err
```

### 错误比较

使用标准Go错误比较方法：

```go
err1 := errors.NewAppError(1006, "未找到")
err2 := errors.NewAppError(1006, "未找到")

// 按代码比较错误
if errors.Is(err1, err2) {
fmt.Println("错误具有相同的代码")
}

// 检查特定错误类型
var appErr *errors.AppCommonError
if errors.As(err1, &appErr) {
fmt.Printf("错误代码: %d\n", appErr.Code())
}
```

### 自定义日志器配置

配置错误日志记录方法使用的默认日志器。日志器是线程安全的，使用延迟初始化以获得最佳性能：

```go
import (
"os"
"github.com/rs/zerolog"
"github.com/gaoxing520/errors"
)

// 设置自定义输出
errors.SetLogOutput(os.Stdout)

// 设置日志级别（默认为InfoLevel）
errors.SetLogLevel(zerolog.DebugLevel)

// 或设置自定义日志器
customLogger := zerolog.New(os.Stdout).
With().
Timestamp().
Caller().
Logger().
Level(zerolog.InfoLevel)
errors.SetLogger(&customLogger)
```

### 完整示例

以下是在应用程序中典型使用的完整示例：

```go
package main

import (
	"fmt"
	"github.com/gaoxing520/errors"
)

// 定义应用程序特定错误
var (
	ErrUserNotFound    = errors.NewAppError(2001, "用户未找到")
	ErrInvalidPassword = errors.NewAppError(2002, "密码无效")
	ErrDatabaseError   = errors.NewAppError(2003, "数据库错误")
)

func authenticateUser(username, password string) error {
	// 模拟用户查找
	if username == "" {
		return ErrUserNotFound.With("用户名为空").LogWarn()
	}

	// 模拟密码验证
	if password != "secret" {
		return ErrInvalidPassword.
			With("用户 %s 身份验证失败", username).
			LogError()
	}

	// 模拟数据库错误
	if username == "dbfail" {
		dbErr := fmt.Errorf("连接丢失")
		return ErrDatabaseError.
			WithCause(dbErr).
			With("验证用户 %s 失败", username).
			LogError()
	}

	// 成功情况
	errors.Success.With("用户 %s 身份验证成功", username).LogInfo()
	return nil
}

func main() {
	// 测试不同场景
	scenarios := []struct {
		username, password string
	}{
		{"", "secret"},       // 空用户名
		{"张三", "wrong"},    // 错误密码
		{"dbfail", "secret"}, // 数据库错误
		{"张三", "secret"},   // 成功
	}

	for _, scenario := range scenarios {
		err := authenticateUser(scenario.username, scenario.password)
		if err != nil {
			fmt.Printf("身份验证失败: %s\n", err.Error())
		} else {
			fmt.Println("身份验证成功")
		}
	}
}
```

## 工具函数

库提供了有用的工具函数：

```go
// 检查错误是否为AppError
if appErr, ok := errors.IsAppError(err); ok {
fmt.Printf("应用错误，代码: %d\n", appErr.Code())
}

// 获取错误代码（如果不是AppError则返回-1）
code := errors.GetErrorCode(err)
fmt.Printf("错误代码: %d\n", code)
```

## 预定义错误

库包含常见HTTP状态码的预定义错误：

```go
errors.ErrInvalidInput // 400 - 无效输入
errors.ErrUnauthorized // 401 - 未授权
errors.ErrForbidden          // 403 - 禁止访问
errors.ErrNotFound           // 404 - 未找到
errors.ErrConflict           // 409 - 冲突
errors.ErrInternalError      // 500 - 内部服务器错误
errors.ErrServiceUnavailable // 503 - 服务不可用
```

## 开发

### 运行测试

```bash
# 运行所有测试
make test

# 运行带覆盖率的测试
make cover

# 运行基准测试
make bench

# 运行所有检查（格式化、审查、测试、基准）
make all
```

### 项目结构

```
├── error.go           # 核心错误类型和函数
├── logger.go          # 日志集成
├── *_test.go         # 测试文件
├── docs/             # 文档
│   └── README_zh.md  # 中文文档（本文件）
└── .github/workflows/ # CI/CD工作流
```

## 贡献

1. Fork 仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 进行更改
4. 为更改添加测试
5. 运行测试套件 (`make all`)
6. 提交更改 (`git commit -m 'Add amazing feature'`)
7. 推送到分支 (`git push origin feature/amazing-feature`)
8. 打开Pull Request

## 许可证

此项目根据 [MIT LICENSE] 许可 - 详情请参见 [LICENSE](../LICENSE) 文件。
