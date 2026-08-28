# GUIDE-LEARN-PATH · 三章分轨道学习导航

> 这是你通关 refund-shop 的主路线。三章可独立，建议顺序：**S-A → B1 → B2 → B3 → S-B/S-C**。
> ⚠️ 兜底纪律：练每队长前，**不要**打开 `docs/PITFALLS-*.md`（答案本）与 `docs/CHECKLIST-AFTER-DOING.md`。

---

## PART 0 · 先读这些（5 分钟）
1. `README.md` 新手 3 句话（数据/入口/坑）+ 3 命令。
2. `docs/code-map.md` 六板块地图 → 知道项目长啥样。
3. `docs/key-files-quickref.md` → 5 个关键文件速查。

## PART 1 · S-A 探索既有代码库（共享单元）
- **目标**：不写代码，先描述这个项目。
- **做**：翻开 `code-map.md` 填空 6 板块 → 用 `key-files-quickref.md` 找到 5 个关键文件入口。
- **验收**：能对着 air 说清「架构/模块/入口/数据/依赖」+ 指出风险。
- 产出：口头/一页笔记即可，不落盘。

## PART 2 · B1 需求澄清（main 分支已具备）
- **目标**：把 refund-shop 的模糊点澄清成可验收需求。
- **输入**：原始需求话术（见 spec 附录-C 或教程给定）。
- **步骤**① 定位 3 个模糊点（运费退法/用券返法/超额策略）→ ② 写 `design-skeleton.md` 七段 → ③ `ADR-001` → ④ 定每点验收 → ⑤ 分步实现 → ⑥ 复盘。
- **完成**：`docs/test-report-b1-base.md` 填上用例与结果。
- **自检**：AI 用 `docs/PROMPT-AI-REVIEWER.md` 阅卷（答案本 PITFALLS-B1）。

## PART 3 · B2 Bug 修复（切 b2-bug 分支）
- **目标**：修掉预埋的两个可复现 Bug。
- **前置**：本仓库 main 目前已修复（教学基线）；b2-bug 是公开教学分支，学员 `git checkout b2-bug` 直接切过去练习。
- **步骤**：红测试→修→绿→防再犯测试→最小 diff。
- **自检**：`docs/PITFALLS-B2.md` + AI-REVIEWER。
- **完成后必做**：在 b2-bug 分支跑 `go test ./...` 全绿 → 对比 `solutions` 分支公开答案自查（`git checkout solutions`）。

## PART 4 · B3 重构（行为保真）
- **目标**：在**契约完全不变**前提下消除 3 个坏味道（Handler 重复样板 / Schema 单文件两表 / 规则超长函数）。
- **铁律**：`api-reference.md` 与 `*_test.go` **diff 必须为空**；覆盖率不下滑；9 测试全绿。
- **流程**：重构→跑 `api-reference diff` 空→`go test` 绿→`go vet` 净→AI 复核保真。
- **自检**：`docs/PITFALLS-B3.md` + AI-REVIEWER。

## PART 5 · S-B 测试与回归 / S-C 文档交付（贯穿）
- B2/B3 每步都跑 `go test ./...`（回归）。
- 每章完结用 `docs/delivery-checklist.md` 7 大项打勾交付。
- 改契约先查 `docs/api-reference.md` 并同步三处（文档↔测试↔前端 fetch）。

---

## 通关自检总清单
做完 B1/B2/B3 后：跑 `docs/CHECKLIST-AFTER-DOING.md`（58 条），填完成声明。
若超过 40 条达标且章内结论为 💪，即可宣称通关；否则按「漏了去哪补线索」返工。