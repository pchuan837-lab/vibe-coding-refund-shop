# GUIDE-LEARN-PATH · 三章分轨道学习导航

> 这是你通关 refund-shop 的主路线。三章可独立，建议顺序：**S-A → B1 → B2 → B3 → S-B/S-C**。
> ⚠️ 兜底纪律：练每段完毕前，**不要**去 `solutions` 分支翻看答案（阅卷答案本只放在 solutions 分支，练习期间打开 = 白练）。
>
> **行号容差声明**：本仓库文档随版本演进会漂移，本文件及 README 的行号仅供当前版本定位参考。外部引用本仓库某行时，请以「章节标题 + 关键词」为准。

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

## PART 2 · B1 需求澄清（b1 分支已具备）
- **目标**：把 refund-shop 的模糊点澄清成可验收需求。
- **输入**：原始需求话术（见 spec 附录-C 或教程给定）。
- **步骤**① 定位 4 条需求锚点（3 个模糊点 + 1 个幂等漏项：运费退法/用券返法/超额策略/幂等兜底）→ ② 写 `design-skeleton.md` 七段 → ③ `ADR-001` → ④ 定每点验收 → ⑤ 分步实现 → ⑥ 复盘。
- **完成**：`docs/test-report-b1-base.md` 填上用例与结果。
- **完成后自检**：全部做完后，切到 `solutions` 分支（`git checkout solutions`），用 `docs/PROMPT-AI-REVIEWER.md` 让 AI 按标准阅卷自查。（答案本只放在 solutions 分支，练习期间请勿回看解决方案。）

## PART 3 · B2 Bug 修复（切 b2-bug 分支）
- **目标**：修掉预埋的两个可复现 Bug。
- **前置**：本仓库 main 目前已修复（教学基线）；b2-bug 是公开教学分支，学员 `git checkout b2-bug` 直接切过去练习。
- **步骤**：红测试→修→绿→防再犯测试→最小 diff。
- **自检**：切到 `solutions` 分支后，用 `docs/PROMPT-AI-REVIEWER.md` 让 AI 按标准阅卷自查。
- **完成后必做**：在 b2-bug 分支跑 `go test ./...` 全绿 → 对比 `solutions` 分支公开答案自查（`git checkout solutions`）。

## PART 4 · B3 重构（行为保真，切 main 分支练习）
- **开始前**：git checkout main（B3 的坏味道锚点在已修复基线 main 分支；b3 分支已废弃/保留参考、勿用于 B3 练习）。
- **目标**：在**契约完全不变**前提下消除 3 个坏味道（Handler 重复样板 / Schema 单文件两表 / 规则超长函数）。
- **铁律**：`api-reference.md` 与 `*_test.go` **diff 必须为空**；覆盖率不下滑；15 测试全绿（domain 7 + routes 8，见 `docs/test-report-b1-base.md`）。
- **流程**：重构→跑 `api-reference diff` 空→`go test` 绿→`go vet` 净→AI 复核保真。
- **自检**：切到 `solutions` 分支后，用 `docs/PROMPT-AI-REVIEWER.md` 让 AI 按标准阅卷自查。

## PART 5 · S-B 测试与回归 / S-C 文档交付（贯穿）
- B2/B3 每步都跑 `go test ./...`（回归）。
- 每章完结用 `docs/delivery-checklist.md` 7 大项打勾交付。
- 改契约先查 `docs/api-reference.md` 并同步三处（文档↔测试↔前端 fetch）。

---

## 通关自检总清单
做完 B1/B2/B3 后：跑 `docs/CHECKLIST-AFTER-DOING.md`（58 条），填完成声明。
若超过 40 条达标且章内结论为 💪，即可宣称通关；否则按「漏了去哪补线索」返工。