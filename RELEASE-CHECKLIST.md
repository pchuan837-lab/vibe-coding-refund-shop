# refund-shop · 公开发布前检查清单

> 用途：本仓库目前保持**私有**。将来要切换为 Public 前，**必须**按本清单逐项勾销，
> 否则学员一旦 clone 就看到答案、三条教学轨道（B1/B2/B3）全部白练。
>
> 执行时机：确定要公开的那一天，先跑本清单 → 再在 GitHub 改可见性。

## A. 必删/必改名答案本（clonable 分支上不得存在）
| 文件 | 原因 | 处理 |
|---|---|---|
| `docs/PITFALLS-B1.md` | B1 需求标准答案 | 从公开仓库移除（可在私有归档分支下架） |
| `docs/PITFALLS-B2.md` | B2 Bug 修复红线 | 同上 |
| `docs/PITFALLS-B3.md` | B3 重构保真硬指标 | 同上 |
| `docs/CHECKLIST-AFTER-DOING.md` | 58 条查漏清单（泄露各轨道完成判定）| 同上 |
| `internal/b2bug/inject.go` + main.go 的 blank-import | b2-bug 分支 Bug 注入载体，只应存在于 b2-bug 分支 | 确认 `main` 分支不 import 该包；b2-bug 分支不公开或作独立练习包 |

## B. 公开前的 git 操作（推荐顺序）
```powershell
# 1) 在 main 分支移除答案本并提交
git checkout main
git rm docs/PITFALLS-B1.md docs/PITFALLS-B2.md docs/PITFALLS-B3.md docs/CHECKLIST-AFTER-DOING.md
git commit -m "chore(release): remove answer-key docs before public"
git push origin main

# 2) 确认 main 不含 b2bug 注入
#   main 本身不承诺 b2-bug 分支公开，按需决定
git grep -n "internal/b2bug" main -- main.go   # 期望：无输出

# 3) 最后一步才在 GitHub 改可见性
gh repo edit pchuan837-lab/refund-shop --visibility public
```

## C. 公开后建议
- 把 `docs/` 下已移除的 4 份答案本，归档到讲师私有位置（如独立私有 repo / 内部网盘），讲师阅卷仍可用。
- B2 轨道若公开，建议只公开到 `b2-bug` 分支且不附 `inject.go` 的说明；或仅提供 main 供 B1/B3 练习，b2-bug 由讲师私下下发。

> ✅ 勾完 A + B 后，方可公开。有任何漏项，先不要公开。