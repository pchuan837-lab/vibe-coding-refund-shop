# CHANGELOG

本文件记录 refund-shop 的版本与变更。格式参考 Keep a Changelog：`新增 / 修复 / 变更 / 移除`。

## [0.1.0] - 2026-08-20
### 新增
- 后端 6 条 REST 接口：下单 / 订单列表 / 订单详情 / 申请退款 / 退款列表 / 审批退款。
- 三页前端（下单页 / 订单页 / 审核页），由 `main.go` `go:embed` 打包，单进程起全栈。
- SQLite 持久化（`orders` / `refunds` 两表 + CHECK 约束）。
- 退款纯函数 `CalcRefundable`（用量分制）。
- 测试 10 条（`internal/domain` 7 + `internal/routes` 3），覆盖率约 53.7%。
- 教学配套文档 14 份（spec + code-map/key-files/B1 六样本/阅卷四件套/PROMPT/CHECKLIST 等）。

### 修复
- （本版本为教学基线，B2 预埋 Bug 不在此修——见 `docs/PITFALLS-B2.md`，需切 b2-bug 分支练习后修复。）

### 变更
- 技术栈由 Node/Express/better-sqlite3 调整为 Go/Gin/modernc.org/sqlite（详见 `docs/ADR-001-tech-stack.md`）。

### 移除
- 移除初版 Node.js 遗留文件（`server/`、`package.json`）。

---

> 命名建议：`[0.1.0]` 为 B1 基线。B2 修复后建议 `[0.2.0]`，B3 重构后 `[1.0.0]`。