# ADR-001 · 技术栈决策记录（B1③ 决策记录）

> 状态：已接受（2026-08-20）
> 对齐 vibeCoding 共享单元 B「需求澄清」之「③ 决策记录 ADR」。

---

## 背景
`refund-shop` 是《Vibe Coding 工程化》教程配套教学项目，需满足：教学性（可预埋问题）、可运行性（新学员克隆后几条命令能跑）、易搜索（StackOverflow 资料多的主流栈）。最初尝试 Node.js + Express + better-sqlite3，在 Windows + Node 24 下 `better-sqlite3` 无预编译二进制、需 CGO/Python 现场编译，friction 过高。

## 候选方案
| # | 方案 | 说明 |
|---|---|---|
| A | Node + Express + better-sqlite3 | 初选，生态最大 |
| B | **Go + Gin + modernc.org/sqlite** | Go 纯编译 + 纯 Go SQLite |
| C | Go + 标准库 net/http + sqlite | 更少依赖 |
| D | Go + Gin + 外部 MySQL/Postgres | 多一层运维 |

## 决策
采用 **B：Go 1.21+ + Gin + SQLite(modernc.org/sqlite 纯 Go 驱动)**。前端用 Go `embed` 打包三页静态文件，单二进制 `go run .` 起全栈。

## 理由
1. **零 CGO 依赖**：modernc sqlite 纯 Go 实现，Windows `go build` 直接过，根治编译摩擦根因。
2. **一条命令**：`go run .` 同时起后端 + embed 前端，符合"新克隆 3 命令可跑"验收线。
3. **测试友好**：`net/http/httptest` 内置，`go test ./...` 即集成测。
4. **教学搜索友好**：Gin 是主流框架，新手报错好搜。
5. 后端为作者熟练栈，能更快写出高质量教学样板。

## 代价
- 无 HMR（hot reload），改代码手动重启 `go run .`。
- 生态资料少于 Node，但替代资料充足。
- modernc sqlite 相对 C sqlite 略有性能差——对本教学项目规模可忽略。

---

### 决策时间线
| 日期 | 动作 |
|---|---|
| 2026-08-20 | 换用 B 方案（由 02-权威设计 批准记录） |