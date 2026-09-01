# refund-shop · 售后退款教学项目

> 《Vibe Coding 工程化》教程的配套教学项目：用最少的命令跑起来的全栈 Web，用于对照练习 **S-A / B1/B2/B3 / S-B/S-C** 三章工作流。
> 
> **教程在线阅读**：[《Vibe Coding 工程化》手册](https://pchuan837-lab.github.io/vibe-coding-engineering/index.html)（GitHub Pages 托管） · **教程源码仓**：[vibe-coding-engineering](https://github.com/pchuan837-lab/vibe-coding-engineering)
>
> **源码位置**：本仓库（`refund-shop`）`main` 分支即教学基线代码；B1 练习切 `b1`、B2 练习切 `b2-bug`、B3 练习切 `main`（已修复基线）、答案对照切 `solutions`；历史分支 `b3` 已废弃/保留参考、勿用于练习（详见下方「三条练习轨道入口」）。

## 新手 3 句话设计卡
- **数据放哪** → SQLite（`internal/db/schema.sql` 建表 + `internal/db/db.go` 连库）。
- **入口在哪** → `main.go`，一条 `go run .` 起全栈（后端 + 三页前端，同端口 3000）。
- **坑在哪** → 金额一律用**分**存整数（别用浮点）；`CHECK(amount>0)` 约束别删；测试用 `:memory:` 别污染 `data.db`。

---

## 〇、前置条件（非技术读者请先确认）
- 已装 **Go 1.21+**（命令行验证：`go version` 有版本号输出）
- 已装 **Git**（命令行验证：`git --version` 有版本号输出）
- 已 clone 本仓库到本地

## 一、3 条命令启动（新克隆验收线）
```powershell
go mod download
go run .          # 默认 http://localhost:3000
go test ./...     # 全量测试
```

## 二、冒烟 5 步（验证全链路）
浏览器开 http://localhost:3000 依次：
1. 下单页 `index.html`：下单「保温杯 99.00」。
2. 订单页 `orders.html`：看到该订单 → 申请退款。
3. 审核页 `admin.html`：审批该笔退款为「通过」。
4. 回订单页：订单状态变 `partial_refunded`（若退满则 `fully_refunded`）。
或 CLI（PowerShell）：
```powershell
$h=@{'Content-Type'='application/json'}
Invoke-RestMethod -Uri http://localhost:3000/api/orders -Method Post -Headers $h -Body '{"product_name":"保温杯","amount":9900,"shipping":500,"coupon_used":0}'
```

---

## 三、三条练习轨道入口

### S-A 探索既有代码库（新项目简化版）
- 地图：`docs/code-map.md` · 速查表：`docs/key-files-quickref.md`

### B1 需求澄清（b1 分支·已具备）
- 开始前：`git checkout b1`（本轨道教学现场；4 条需求锚点——3 个模糊点 + 1 个幂等漏项——锚在 `refund_rules.go` 顶部 TODO，原始需求话术见 spec 附录-C）。
- 六样本：`docs/design-skeleton.md` / `ADR-001` / `implementation-sample-record` / `test-report-b1-base` / `delivery-checklist` / `api-reference`

### B2 Bug 修复（需切 b2-bug 分支）
```powershell
git fetch origin
git checkout b2-bug
```
`b2-bug` 是**公开教学分支**，相对 main 只新增 `internal/b2bug/` 注入包，预埋 2 个可复现 Bug；
main 分支不 import 该包，始终保持正确（amount=0 → 400）。

### B3 重构（行为保真，消除坏味道，切 main 分支练习）
- 开始前：`git checkout main`（`b3` 分支已废弃/保留参考、勿用于 B3 练习）。
- 坏味道锚点注释已在 `internal/routes/*.go`、`internal/db/schema.sql`、`refund_rules.go` 标注。

### 阅卷 / 查漏（练习后，切 solutions 分支）
- AI 阅卷提示词：`docs/PROMPT-AI-REVIEWER.md`（本分支仍可打开）
- 人工查漏 58 条 + 三条轨道答案本：切到 `solutions` 分支查看
  （⚠️ 练习完成前严禁提前切 solutions，否则泄题自废练习价值）

---

## 四、目录 / 技术栈
```
main.go                入口（embed 前端 + 路由挂载）
internal/routes/       HTTP 层（orders/refunds 6 端点）
internal/domain/       退款纯函数 CalcRemaining
internal/db/           schema.sql + db.go
public/                三页前端
docs/                  教学配套文档（10 份）；阅卷/答案见 solutions 分支
```
Go 1.21+ · Gin · modernc.org/sqlite（零 CGO）· 详见 `docs/ADR-001-tech-stack.md`。

## 五、更多
- 接口契约：`docs/api-reference.md`
- 接口变更记录：`CHANGELOG.md`
- 分轨道完整导航：`GUIDE-LEARN-PATH.md`（项目根目录）