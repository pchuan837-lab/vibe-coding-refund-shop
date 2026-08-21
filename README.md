# refund-shop · 售后退款教学项目

> 《Vibe Coding 工程化》教程的配套教学项目：用最少的命令跑起来的全栈 Web，用于对照练习 **S-A / B1/B2/B3 / S-B/S-C** 三章工作流。

## 新手 3 句话设计卡
- **数据放哪** → SQLite（`schema.sql` 建表 + `internal/db/db.go` 连库）。
- **入口在哪** → `main.go`，一条 `go run .` 起全栈（后端 + 三页前端，同端口 3000）。
- **坑在哪** → 金额一律用**分**存整数（别用浮点）；`CHECK(amount>0)` 约束别删；测试用 `:memory:` 别污染 `data.db`。

---

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

### B1 需求澄清（main 分支·已具备）
- 六样本：`docs/design-skeleton.md` / `ADR-001` / `implementation-sample-record` / `test-report-b1-base` / `delivery-checklist` / `api-reference`

### B2 Bug 修复（需切 b2-bug 分支）
```powershell
git fetch origin
git checkout -b b2-bug origin/b2-bug
```
`b2-bug` 是本仓库的一个**独立教学分支**（公开分支，本地 + 远程均有），相对 main 只新增
`internal/b2bug/`、放宽 schema、blank-import 注入，预埋了可复现的 2 个 Bug；
main 分支不 import 该包，始终保持正确（amount=0 → 400）。

### B3 重构（行为保真，消除坏味道）
- 坏味道锚点注释已在 `internal/routes/*.go`、`schema.sql`、`refund_rules.go` 标注。

### 阅卷 / 查漏（练习后，切 solutions 分支）
- AI 阅卷提示词：`docs/PROMPT-AI-REVIEWER.md`（本分支仍可打开）
- 人工查漏 58 条 + 三条轨道答案本：切到 `solutions` 分支查看
  （⚠️ 练习完成前严禁提前切 solutions，否则泄题自废练习价值）

---

## 四、目录 / 技术栈
```
main.go                入口（embed 前端 + 路由挂载）
internal/routes/       HTTP 层（orders/refunds 6 端点）
internal/domain/       退款纯函数 CalcRefundable
internal/db/           schema.sql + db.go
public/                三页前端
docs/                  教学配套文档（10 份）；阅卷/答案见 solutions 分支
```
Go 1.21+ · Gin · modernc.org/sqlite（零 CGO）· 详见 `docs/ADR-001-tech-stack.md`。

## 五、更多
- 接口契约：`docs/api-reference.md`
- 接口变更记录：`CHANGELOG.md`
- 分轨道完整导航：`GUIDE-LEARN-PATH.md`（项目根目录）

---

## 六、solutions 分支 · 使用纪律（练习完成前严禁读取本节以下内容）

> ⚠️ **本分支 = 答案区**，包含三轨道标准答案与人工查漏清单。**必须先在对应教学分支（b1/b2-bug/b3/main）独立完成练习，再切到本分支对照。**
> 提前看答案 = 自废练习价值。

### 📚 本分支答案本索引（均在 `docs/`）
| 文件 | 用途 | 对应轨道 |
|---|---|---|
| `docs/PITFALLS-B1.md` | B1 需求澄清 15 条答案与完成判定标准 | B1 |
| `docs/PITFALLS-B2.md` | B2 Bug 修复 2 个 Bug 精确定位 + 7 条过程纪律 | B2 |
| `docs/PITFALLS-B3.md` | B3 重构 3 处坏味道定位 + 停止条件 + 9 条保真指标 | B3 |
| `docs/CHECKLIST-AFTER-DOING.md` | 人工查漏 58 项（B1 24 + B2 16 + B3 18）+ 完成声明 | 三轨通用 |
| `docs/PROMPT-AI-REVIEWER.md` | AI 阅卷提示词（练习后用 AI 对照审核） | 三轨通用 |

### 🤖 用 AI 审核的框架（严格照做才有公信力）
1. **审核基准 = 行为 + 验收，不量外观**：你的实现**允许字段名、文件结构、实现方式与 solutions 不同**；只要行为等价且满足验收标准即通过。AI 必须量"测试用例 + 验收清单 + 关键观景点"，**禁止比对代码文本**。
2. **上下文隔离纪律（防记忆污染，必须执行）**：**不要**在做练习的那个对话窗口里做审核。最低要求：**新建一个对话窗口**跑审核；最佳实践：**用一个独立的 coding 工具**（新开 agent 或独立 IDE）打开 solutions 分支做审核——避免教学对话的历史上下文让 AI "自我循环确认"，导致判分不客观。
3. **审核后要能解释差异**：看完 solutions 后，如果你的实现与答案不同，先自行解释"我的做法在行为上等价吗？少/多了什么？"再决定是否回改——**写法不同不代表错**，只有行为或验收项缺失才是真的错。

### 🔁 切回练习分支的提醒
- 看完答案后，如需回去继续修改：`git checkout main`（或 `b1` / `b2-bug` / `b3`），不要在本分支直接改业务代码（本分支仅作答案对照）。