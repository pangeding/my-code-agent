# Go 语言学习计划

## 📅 学习周期建议
- 总周期：8-12 周（可根据自身基础调整）
- 每日投入：1-2 小时

## 🟢 第一阶段：基础入门（第 1-2 周）
**目标**：掌握 Go 语法基础，能编写简单命令行程序
**核心内容**：
- 环境搭建（Go 安装、Go Modules、IDE 配置）
- 基本语法：变量、常量、数据类型、运算符
- 流程控制：if/else、for、switch
- 函数：定义、参数、返回值、defer、匿名函数与闭包
- 数组与切片（Slice）
- 映射（Map）
**推荐资源**：
- [A Tour of Go](https://go.dev/tour/)
- 《Go 语言圣经》第 1-4 章
**练习**：完成 A Tour of Go 全部练习，编写一个简单的命令行工具（如计算器或待办事项列表）

## 🟡 第二阶段：核心特性（第 3-4 周）
**目标**：深入理解 Go 的核心编程范式
**核心内容**：
- 结构体（Struct）与方法
- 接口（Interface）与组合优于继承
- 错误处理（Error Handling）
- 并发编程基础：Goroutine、Channel、select
- 同步机制：sync 包（WaitGroup, Mutex, RWMutex）
- 包管理与依赖（go mod）
**推荐资源**：
- [Go by Example](https://gobyexample.com/)
- 《Go 语言圣经》第 5-8 章
**练习**：实现一个并发安全的缓存/计数器；编写使用 Channel 的生产者-消费者模型

## 🟠 第三阶段：工程实践（第 5-7 周）
**目标**：掌握 Go 在真实项目中的应用与最佳实践
**核心内容**：
- 标准库精讲：net/http、io、os、encoding/json、log
- Web 开发基础：路由、中间件、RESTful API 设计
- 数据库交互：database/sql、GORM 或 sqlx
- 测试：单元测试、基准测试、Table-Driven Tests
- 性能调优与 pprof、trace 使用
- 项目结构与规范（Standard Go Project Layout、命名、注释规范）
**推荐资源**：
- [Go 标准库文档](https://pkg.go.dev/std)
- [Effective Go](https://go.dev/doc/effective_go)
- 官方 Go Testing 指南
**练习**：开发一个完整的 RESTful API 服务（如博客/用户管理系统），包含 CRUD、数据库操作、单元测试和日志记录

## 🔴 第四阶段：进阶与实战（第 8-10+ 周）
**目标**：具备独立开发生产级应用的能力
**核心内容**：
- 高级并发模式：Context、Worker Pool、Pipeline、Rate Limiter
- 微服务基础：gRPC、Protobuf、服务注册与发现
- 容器化与部署：Docker、Makefile、CI/CD 基础
- 源码阅读：标准库或优秀开源项目（如 Gin, Cobra, Viper）
- 常见设计模式在 Go 中的实现
**推荐资源**：
- Go 官方高级并发模式教程
- GitHub 优质开源项目（关注 `good first issue`）
**实战项目**：
- 开发一个支持高并发的爬虫或文件同步服务
- 参与一个开源 Go 项目，尝试提交 PR

## 📝 学习建议与最佳实践
- ✅ 坚持写代码：光看不练假把式，每天至少动手写 30 分钟
- ✅ 善用官方文档：`go doc`、`pkg.go.dev` 是你的第一参考源
- ✅ 理解 Go 哲学：简单、显式、组合优于继承、错误即值
- ✅ 掌握调试技巧：`delve` (dlv)、结构化日志、pprof 性能分析
- ✅ 加入社区：Go 官方论坛、Reddit r/golang、GitHub Discussions、技术博客

## 📊 进度追踪表
| 阶段 | 状态 | 完成日期 | 备注 |
|------|------|----------|------|
| 基础入门 | ⬜ 未开始 | | |
| 核心特性 | ⬜ 未开始 | | |
| 工程实践 | ⬜ 未开始 | | |
| 进阶实战 | ⬜ 未开始 | | |

> 💡 **提示**：本计划为通用参考模板，请根据自身基础和学习节奏灵活调整。Go 的精髓在于**动手实践**与**阅读源码**。祝你学习顺利，早日成为 Gopher！🚀