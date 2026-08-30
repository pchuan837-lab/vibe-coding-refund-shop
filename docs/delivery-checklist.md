# delivery-checklist · S-C 文档交付核对清单（7 大项）

> 对齐 vibeCoding 共享单元 S-C「文档交付」。
> 每次交付前逐项打勾，冲突即回改。当前状态：B1 基线交付。点击即跳转对应文件。

| # | 交付项 | 覆盖文件 | 状态 | 说明 |
|---|---|---|---|---|
| 1 | 代码 | `main.go` + `internal/**` + `public/**` | ✅ | 6 路由 + 三页端到端可跑 |
| 2 | 设计 | `docs/design-skeleton.md` | ✅ | 七段方案（4/5 已预填） |
| 3 | 测试 | `docs/test-report-b1-base.md` + 各 `*_test.go` | ✅ | 10 测试全绿（domain 7 + routes 3），覆盖 53.7% |
| 4 | 部署 | `README.md`（3 命令）+ `ADR-001`（选型） | ✅ | 新克隆 3 命令可跑 |
| 5 | API | `docs/api-reference.md` | ✅ | 6 路由契约 |
| 6 | CHANGELOG | `CHANGELOG.md` | ✅ | 版本 0.1.0，分类四 |
| 7 | README | 根目录 `README.md` | ✅ | 三启动命令 + 三轨道入口 + 新手 3 句话卡 |

### 校验提示
- 改动**契约**（路由/字段）：回看第 5 项 `api-reference.md`，并查前端的 fetch 是否同步。
- 改动**建表**：回看 `schema.sql`，跑一次 `go test ./...` 确认无回归。
- 改动**入口**：回看 `main.go`，确认 embed 路径与三个 html 一致。

### 完成声明
- 当前同步：`main` 分支基线 + `dev` 分支开发中。
- 待办：B2/B3 各赛道交付时，本表逐项复核。