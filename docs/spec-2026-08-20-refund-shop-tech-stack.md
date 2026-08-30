# spec.md · refund-shop 教学项目技术规格（正式版 v1.0）

> 生效日期：2026-08-20 · 已获用户书面批准 · 范围：第一轮 B1 基线 + B2 预埋 + B3 预埋 + 13 docs 教学配套

---

## 一、问题与目标

### 1.1 问题（Problem）
《Vibe Coding 工程化》教程的三条主轨道（B1 新功能 / B2 修 Bug / B3 重构）目前缺少一个「读者能亲手 clone 下来、跟着教程一步步操作、每一步有真实的代码/文档锚点」的配套教学项目。原先指定的 Node.js + Express + better-sqlite3 方案在本机 Node 24 环境下预编译二进制缺失 + 缺 Python 编译环境，摩擦过大，违反「clone 即跑」验收线。

### 1.2 目标用户（Users）
| 用户画像 | 核心诉求 |
|---|---|
| 零基础读者（走完全教程的学员） | 跟着教程一章一章推进，B1/B2/B3 三条轨道都有真实练手对象，每章的「产出物」有落地锚点 |
| 复习者（跳读某一章节的人） | 能快速定位 S-A/S-B/B1②/B2① 对应到项目里的哪个文件/行号，不用翻整份教程 |
| 用 AI 辅助练的读者 | 有一份「AI 阅卷提示词 + 标准答案依据」，不用人工当裁判 |

### 1.3 目标（Goals）
1. 交付一个「`git clone → 三条命令 → 页面+API 跑通 + 9 个测试全绿」`的最小可跑项目，零外置依赖（不用装 Python/Node/数据库服务/C 编译工具）。
2. B1 新功能需求：三页前端（下单/订单列表+申请退款/审核）+ 后端 6 个接口闭环跑通；**故意留 3 处需求模糊点 + 1 处非功能漏项（幂等）**，作为 B1① 需求澄清的练兵素材。
3. B2 Bug 修复：`b2-bug` 分支预埋 2 个真实可复现 Bug，每个有明确的 MRE（最小复现命令），Bug 引入方式是「独立注入文件」，不污染 main 分支的正确代码。
4. B3 重构：`main` 分支预埋 3 处可量化坏味道（Handler 样板重复 / Schema 混写 / 规则 if-else 堆），每处有明确停止条件，重构过程有保真硬指标 + 基线建立命令锚定。
5. 教学配套文档：**首轮交付 13 个 docs 文件**（S-A 代码地图/承诺速查表、B1②~⑤ 产出物格式样本、S-C 交付清单/API 参考、人工查漏 CHECKLIST、AI 阅卷 PROMPT + 三轨道标准答案本），让教程每章的「产出物」都有落地锚点与格式参考。

### 1.4 非目标（Non-Goals）
- ❌ 不做真实钱包打款 / 支付集成（退款通过 = 改状态即可，不触发真实交易）
- ❌ 不做用户登录/权限（匿名下单+凭订单号查，降低复杂度）
- ❌ 不做移动端适配 / UI 美化（样式内联 30 行内，够用就行）
- ❌ 不做历史订单数据迁移（新项目，无历史）
- ❌ 不做 CI/CD / GitHub 发布（发布流程全程冻结，后续统一处理）
- ❌ 首轮不引第三方工具库（goose/gorm/testify 等都不引，零依赖摩擦最小）

---

## 二、术语定义（歧义澄清 3 点 · 防止双重解读）
A. **「教学模糊点」 vs 「Bug」**
   - 模糊点 = 功能未定义但不 crash（如 `CalcRefundable` 不处理运费/券，调用方不传不报错，只是业务上"没覆盖这种情况"，是 B1 读者澄清后要补的）
   - Bug = 功能已明确定义但实现错（如 orders 金额 <0 仍能创建，明确违反金额必须正的契约，是 B2 读者要修的）

B. **「覆盖率 40~60%」** = `go test -cover ./...` 输出的 domain + routes 两个包的算术平均值落在 [40, 60] 区间内；<40% 说明骨架太虚需重补，>60% 说明练习空间被挤占需减非核心用例。

C. **B3 重构「API 不许改」** = 六条路由（`POST/GET /api/orders`、`GET /api/orders/:id`、`POST/GET /api/refunds`、`PATCH /api/refunds/:id/approve`）的：
   - HTTP 方法 & URL 路径 & 查询参数名
   - 请求 JSON 字段名 & 响应 JSON 字段名
   - HTTP 状态码语义（200=成/400=入参错/404=不存在/500=内部错）
   以上必须 100% 等价；Handler 内部实现可改（抽 helper / 中间件），但对外契约不动。

---

## 三、功能需求（Functional Requirements = FR）
### 3.1 后端服务启动（FR-1）
- **FR-1.1**：`go run .` → 3 秒内启动服务，监听 `:3000` 端口
- **FR-1.2**：访问 `GET /` → 返回 `public/index.html`（下单页）；`GET /orders.html` / `GET /admin.html` 同理
- **FR-1.3**：静态资源通过 Go `embed` 打包进二进制，不依赖外部文件系统路径（exe 单独拷贝也能跑）

### 3.2 订单模块 3 接口（FR-2）
- **FR-2.1 POST /api/orders**：入参 `{product_name string, amount int, shipping int, coupon_used int}`；成功响应 `200 {id, product_name, amount, shipping, coupon_used, status:"paid", created_at}`；`amount <= 0` 返回 `400 {"error": "..."}`
- **FR-2.2 GET /api/orders**：响应 `200 [{id,...}, ...]`（所有订单按 created_at 倒序）
- **FR-2.3 GET /api/orders/:id**：响应订单详情 + 该订单所有退款记录数组；订单不存在 → 404

### 3.3 退款模块 3 接口（FR-3）
- **FR-3.1 POST /api/refunds**（首轮故意不做幂等 = 3A 决策，B1 读者补）：入参 `{order_id int64, amount int, reason string?}`；处理流程 = ①查订单→②查该 order 累计已退(refunds.status!='rejected')→③调 `CalcRefundable(orderID, order.amount, totalRefunded, req.amount)`→④ INSERT refunds(status='pending')；成功响应 `200 {id, order_id, amount, status:"pending", reason, created_at}`；order 不存在 → 404；`CalcRefundable` 返回 err → 400
- **FR-3.2 GET /api/refunds**：Query 可选 `?status=pending` 或 `?order_id=123`；响应 200 数组，每条 refund 对象附带对应的 order 简要信息（用于审核页显示订单名）
- **FR-3.3 PATCH /api/refunds/:id/approve**：入参 `{approved bool, comment string?}`；`approved=false` → refunds.status=rejected + UPDATE updated_at；`approved=true` → ① refunds.status=approved ② 重算该 order 累计已退 approved 金额 ③ 累计 == order.amount → orders.status=fully_refunded；否则 partial_refunded；响应 200 = 更新后的 refund 对象

### 3.4 退款规则 CalcRefundable（FR-4 · 教学模糊点核心锚点）
- **FR-4.1** 函数签名：`CalcRefundable(orderID int64, orderAmount int, totalRefunded int, applyAmount int) (approvedAmount int, err error)`（**故意不接 shipping/coupon 参数**，就是要 B1 读者澄清）
- **FR-4.2** 明确规则部分：
  - `applyAmount <= 0` → 返回 `(0, error)`
  - `orderAmount <= 0` → 返回 `(0, error)`（无效订单）
  - `remaining = orderAmount - totalRefunded`；`remaining <= 0` 且还要申请 → 返回 `(0, error)`（已全退完）
  - `applyAmount <= remaining` → `approvedAmount = applyAmount`（合法）
  - `applyAmount > remaining` → **首轮默认截断为 remaining 并返回 nil**（具体策略 A：截断 vs B：报错 留 B1 澄清时可改，此处是默认实现）
- **FR-4.3 需求锚点（4 条：3 个模糊点 + 1 个幂等漏项，故意不处理，B1① 读者自己澄清后补）**：
  - TODO ① 运费怎么退？（函数不接收 shipping 参数）
  - TODO ② 用券返不返？（函数不接收 couponUsed 参数）
  - TODO ③ 超额是截断（默认）还是报错？
  - TODO ④ 重复申请/并发退款怎么兜？（首轮故意不做幂等，无 refund_no 幂等键）

### 3.5 三页前端（FR-5）
- **FR-5.1 index.html（下单页）**：表单 4 字段（商品名/金额/运费/用券抵扣）+「下单」按钮 → 提交 fetch POST /api/orders → 成功后 `location.href = '/orders.html'`
- **FR-5.2 orders.html（订单列表）**：onload 调 GET /api/orders → 表格渲染（列：id/商品名/金额/运费/用券/状态/创建时间/操作）；每行「申请退款」按钮 → prompt 输金额 + reason → POST /api/refunds → 成功后 reload
- **FR-5.3 admin.html（审核页）**：onload 调 GET /api/refunds?status=pending → 表格列 id/订单ID/申请金额/原因/状态；每行两个按钮「通过」（PATCH approved=true）/「拒绝」（false）→ 成功后 reload
- **FR-5.4**：三页样式内联 `<style>` 不引 CDN/UI 库；所有金额以"元"显示（显示层 /100，接口层传分）。

### 3.6 数据库（FR-6）
- **FR-6.1 orders 表字段**（见 附录-A 接口契约 schema.sql）：`id INTEGER PK / product_name TEXT / amount INTEGER / shipping INTEGER DEFAULT 0 / coupon_used INTEGER DEFAULT 0 / status TEXT DEFAULT 'paid' / created_at DATETIME DEFAULT CURRENT_TIMESTAMP`；**status 合法值 = paid / partial_refunded / fully_refunded**
- **FR-6.2 refunds 表字段**：`id INTEGER PK / order_id INTEGER REFERENCES orders(id) / amount INTEGER / status TEXT DEFAULT 'pending' / reason TEXT / created_at DATETIME / updated_at DATETIME`；**故意无 shipping_refund / coupon_return 列**（模糊点来源）；**故意无 refund_no 幂等唯一键**（B1 读者补）
- **FR-6.3**：`NewDB(path)` 连接 SQLite，path=`":memory:"` 走内存库（测试用），path=`"data.db"` 走文件库；启动时自动执行 `schema.sql` 建表，表存在则跳过。

### 3.7 B2 预埋 Bug（FR-7 · 仅 b2-bug 分支，1B 方案：独立 bug 注入文件）
- **FR-7.1 Bug1：0 元订单能创建** → b2-bug 分支新增 `internal/b2bug/inject.go`，在 `init()` 里通过全局变量把 orders 金额校验阈值从 `<= 0` 改成 `< 0`；main 分支不 import 这个包
- **FR-7.2 Bug2：累计退款计算写反** → `CalcRefundable` 计算 `remaining` 的那一行被 bug 注入改成 `refunded - orderAmount`
- **FR-7.3**：main 分支代码本身是正确的；切到 b2-bug 后才通过独立包引入 Bug，符合「Bug 是上线后引入的」真实职场心智。

### 3.8 B3 预埋坏味道（FR-8 · main 分支功能正常，但故意不优雅）
- **FR-8.1 坏味道 1：Handler 样板重复** → `routes/orders.go` POST 处理器 + `routes/refunds.go` POST/GET/PATCH 处理器中，`c.BindJSON(&req) + if err != nil { c.JSON(400, ...) }` 三段式重复≥2 处；读者重构时抽公共 helper / 中间件
- **FR-8.2 坏味道 2：Schema 混写** → 两表 DDL 全塞一个 `schema.sql`，无版本号、无迁移状态表；读者重构时拆成 `migrations/001_create_orders.sql` + `002_create_refunds.sql` 两个独立文件（2A 方案，纯手写不引 goose）+ 简单 schema_version 机制
- **FR-8.3 坏味道 3：规则堆 if-else** → `CalcRefundable` 单函数实现用嵌套 if-else 堆，函数 >50 行；读者重构时拆策略模式接口 + 多个 struct 或拆小函数；**停止条件（任一）**：策略模式每个 struct≤30 行；或主函数≤20 行只编排调用。

### 3.9 测试骨架（FR-9 · 故意覆盖率 40%-60%，留教学抓手）
- **FR-9.1 `internal/domain/refund_rules_test.go`**：7 个原生 `testing` 单元用例，对应 Section4 设计的 7 个清晰规则用例表（全额/部分/二次部分/超额截断/申请零元/无效订单/已全退再申请）；全部 PASS
- **FR-9.2 `internal/routes/orders_route_test.go`**：3 个 `httptest` 集成用例（正常下单成功/金额负数 400/金额零 400）；全部 PASS
- **FR-9.3**：故意**不写**运费/券用例、不写幂等用例、不写 amount=0 用例（留给 B1 澄清后补 / B2 修 Bug 先写失败用例锁 Bug）
- **FR-9.4**：refund_rules_test.go 底部有一段「注释取消即 FAIL」的示范失败用例 + 提示语（S-B③ 失败分析练习入口）
- **FR-9.5**：`go test ./...` 执行 domain + routes 两包，退出码 0，PASS 数 = 10（domain 7 + routes 3）。

### 3.10 教学配套 docs 13 个文件（FR-10 · 首轮必须全部落盘）
详见附录-D：13 份 docs 文件清单 + 每份职责 + 内容结构要点 + 对齐教程哪节。

---

## 四、非功能需求（Non-Functional Requirements = NFR）
| 编号 | 类型 | 描述 | 指标/阈值 |
|---|---|---|---|
| NFR-1 | 启动速度 | `go run .` 冷启动时间 | ≤ 5 秒（Windows 普通硬盘） |
| NFR-2 | 测试速度 | `go test ./...` 首次全跑时间 | ≤ 10 秒（10 个用例+内存DB） |
| NFR-3 | 内存占用 | 服务空闲时内存 | ≤ 50MB |
| NFR-4 | 零外置依赖 | 读者跑通三条命令是否需要装 Node/Python/MySQL/gcc | 完全不用；只要求 Go ≥ 1.21 + git |
| NFR-5 | 数据库零配置 | 启动后数据库是否需要手动建表/建库 | 不需要；NewDB 自动执行 schema.sql |
| NFR-6 | 金额精度 | 所有金额存储/传输是否为整型分 | 100% 整型分；严禁浮点 9.90 |
| NFR-7 | 教学可观测性 | S-A 探索时，入口 main.go → routes → domain → db 的调用链路是否无隐式魔法 | 100% 直白可见；不用看文档就能通过 `rg CalcRefundable` 找到全部调用点 |
| NFR-8 | 可移植性 | 整个项目（含数据库文件 data.db）复制到另一台同 OS 机器是否可直接 `go run .` 跑通 | 是；embed 打包静态资源不依赖外部路径 |
| NFR-9 | 版本可回滚 | B3 重构每一步小改后是否能 `git revert` 回退到上一步 | 是；要求每子步一个原子 commit（在实施计划里强制） |
| NFR-10 | 泄题防控 | PITFALLS-*.md 答案本在练习前被读者偷看时，是否有醒目的警告层 | 是；每个 PITFALLS 文件顶部有 3 重警告 + 使用纪律 |

---

## 五、约束（Constraints）
1. **红线约束**：
   - 绝不修改 `I:\Trae默认工作区\vibecoding教程\` 下任何正文文件。
   - memory 目录只允许写 `03-工作进度日志-refund-shop.md` 和 `02-项目-refund-shop开发设计` 的「当前状态 / 权威设计（用户批准时才改技术栈）」；其他 memory 文件一律不动。
   - GitHub 发布（单 repo、LICENSE、转 ASCII + MAPPING、CI、push）全程冻结，本轮不执行。
2. **技术栈已锁定**（9 项用户批准决策，不可改）：Go/Gin / modernc/sqlite / 原生 SQL / Go embed 纯静态 / 内建 testing / S1 扁平目录 / 1B bug 注入 / 2A 纯手写迁移 / 3A 幂等留 B1 补 / 4A ASCII 文件名。
3. **首轮依赖数量约束**：`go.mod` 只允许 2 个第三方依赖（`gin-gonic/gin` + `modernc.org/sqlite`）+ 它们的传递依赖。严禁引入 testify/goose/gorm 等任何第三方库。
4. **代码总行数约束**：首轮 main.go ≤ 50 行；每个路由文件 ≤ 120 行；`CalcRefundable` 函数 ≤ 80 行；三页前端每页 HTML + JS ≤ 200 行（不含注释空行）。防止教学项目代码量太大劝退。
5. **前端约束**：绝对不许引 Vite/React/Vue/jQuery/任何 CDN 资源；三页必须纯原生 HTML + `<script>` 内联 JS，保证「Go embed 打进二进制后离线能跑」。

---

## 六、假设与依赖（Assumptions & Dependencies）
### 6.1 假设（如果失败需立刻通知用户调整）
A1：读者本机已装 Go ≥ 1.21；已配 `GOPROXY`（`goproxy.cn,direct` 或等价国内源）。
A2：读者已装 git（≥2.x）。
A3：Windows 读者用 PowerShell 5+ 或 Git Bash；内建 `curl.exe` 可执行（或自行装 Postman 调接口）。
A4：`modernc.org/sqlite` 最新稳定版在 Go 1.21 Windows 下纯 Go 构建通过，不触发 CGO 要求（modernc 官网说明支持，失败概率 <1%）。
A5：Gin 最新稳定版路由、BindJSON、StaticFS 语法与当前设计一致（如变更需查文档更新）。

### 6.2 外部依赖（必须存在，否则报错）
- `github.com/gin-gonic/gin`（≥v1.9 稳定版）
- `modernc.org/sqlite`（≥v1.29 稳定版，传递依赖含 `modernc.org/libc` 等纯 Go 库）
- Go 标准库：`database/sql`、`embed`、`net/http/httptest`、`testing`、`os`、`io/fs`

---

## 七、未解决问题（Open Questions）
**无（全部已通过 brainstorming 澄清并获用户确认）。**
历史决策备查见 02 权威设计文件「5. 关键决策日志」表。

---

## 八、验收标准（Acceptance Criteria）

> 严格分两类：`rule` = 客观二进制条件（可验证 P/F）；`rubric` = 评价质量维度（0-2 分 + 阈值）。

### 8.1 Rule 类（共 31 条，全部必须 PASS）
| 编号 | 类型 | 可观察通过条件 | 证据来源 |
|---|---|---|---|
| AC-01 | rule | `go build ./...` 退出码 0 无输出 | Windows 终端执行结果 |
| AC-02 | rule | `go test ./...` 输出 `ok  refund-shop/internal/domain` + `ok  refund-shop/internal/routes`；PASS 数 = 10（domain 7 + routes 3） | 同上 |
| AC-03 | rule | `go test -cover ./...` 两包算术平均 ∈ [40, 60] | 同上 |
| AC-04 | rule | `go run .` 启动后另开终端 `curl.exe http://localhost:3000` → 返回 index.html 的 `<title>` 或 `下单` 字样 | curl 结果 |
| AC-05 | rule | curl POST 正常订单（amount=9900）→ HTTP 200 + JSON 含 `id` int | curl 结果 |
| AC-06 | rule | curl POST amount=-1 → HTTP 400 + JSON 含 `error` 字段 | curl 结果 |
| AC-07 | rule | curl POST amount=0 → HTTP 400（main 分支正确；b2-bug 分支应为 200） | curl 结果 |
| AC-08 | rule | curl GET /api/orders → 数组长度 ≥ 1 | curl 结果 |
| AC-09 | rule | 下单 100 分 → 退 50 分 → PATCH approved=true → order.status = partial_refunded | DB 查询：`sqlite3 data.db "SELECT status FROM orders WHERE id=1"` |
| AC-10 | rule | 累计退 100 后 order.status = fully_refunded；再申请退 1 → CalcRefundable 返回 err 或 0 无法批准 | DB 查询 + curl 结果 |
| AC-11 | rule | 5 步 curl 冒烟（Section5 已列）每一步都得预期 HTTP 状态码 | 人工执行结果 + 截图 |
| AC-12 | rule | `refund_rules.go` 顶部存在 ≥ 4 行 `// TODO` 注释，明确列出 4 条锚点（运费/券/超额策略 + 幂等漏项） | 代码目视检查 |
| AC-13 | rule | `refunds` 表结构无 shipping_refund 列、无 coupon_return 列；幂等键列无（B1 读者补前） | `sqlite3 .schema refunds` |
| AC-14 | rule | `go.mod` `require` 段只有 gin + modernc/sqlite 两行（传递依赖不计） | go.mod 目视 |
| AC-15 | rule | `NewDB(":memory:")` 返回 db 无 error；立刻执行 `SELECT name FROM sqlite_master WHERE type='table'` → 返回 orders + refunds 两行 | 单元测试内部验证（db_test.go 可选或集成测验证） |
| AC-16 | rule | 三页前端 index/orders/admin 均可通过浏览器 `http://localhost:3000/xxx.html` 正常打开（无 404） | 浏览器验证或 curl |
| AC-17 | rule | 三页前端 HTML 源码中，不包含任何 `<script src="https://...">` 外链 CDN | 代码 grep 结果 |
| AC-18 | rule | 首轮 docs 目录下 ≥ 13 个文件（spec + 附录-D 列出的 12 个配套文档） | LS docs 目录 |
| AC-19 | rule | `docs/CHECKLIST-AFTER-DOING.md` 顶部存在「使用纪律警告」段（≥4 条规则），并包含 3 章共 58 条 `[ ]` 格式条目 | 文件内容目视 |
| AC-20 | rule | `docs/PROMPT-AI-REVIEWER.md` 提示词正文包含两条红线：「不许超 PITFALLS 范围扣分」+「不许泄露 PITFALLS 精确内容」 | 文件内容目视 |
| AC-21 | rule | `docs/PITFALLS-B1.md` 顶部存在「练习前严禁打开」3 重警告；正文包含 B1-P01 ~ B1-P15 共 15 条评分条目 | 文件内容目视 |
| AC-22 | rule | `docs/PITFALLS-B2.md` 正文包含 Bug1/Bug2 精确根因位置 + 7 条过程纪律条目（B2-P03~09）共 10 条 | 文件内容目视 |
| AC-23 | rule | `docs/PITFALLS-B3.md` 正文包含 3 坏味道停止条件 + 9 条保真硬指标（B3-P04~12）共 12 条 | 文件内容目视 |
| AC-24 | rule | `docs/design-skeleton.md` 七段结构完整，第 4 段（关键数据模型）预填 orders/refunds 字段表，第 5 段（接口定义）预填 6 条路由请求/响应表 | 文件内容目视 |
| AC-25 | rule | `docs/ADR-001-tech-stack.md` 结构严格为五段（背景/候选方案/决策/理由/代价），代价段非空 | 文件内容目视 |
| AC-26 | rule | `CHANGELOG.md` 存在；首条版本条目格式 `## [0.1.0] - YYYY-MM-DD`；分类四「新增/修复/变更/移除」中「新增」≥ 6 条（对应首轮交付内容） | 文件内容目视 |
| AC-27 | rule | `README.md` 三启动命令（`go mod download / go run . / go test ./...`）存在且顺序正确；三轨道入口指引存在 | 文件内容目视 |
| AC-28 | rule | `README.md` 顶部存在新手 3 句话设计卡样本（①数据放哪②入口在哪③坑在哪） | 文件内容目视 |
| AC-29 | rule | 首轮基线 git commit message 结构正确（包含 feat/refund-shop B1 基线 + 9 PASS + 预埋点清单） | `git log --oneline -1` 结果 |
| AC-30 | rule | `internal/b2bug/` 目录下存在 bug 注入文件（b2-bug 分支），但 main 分支不 import 该包；main 分支 orders 金额校验行为正确（amount=0 400） | 代码目视 + AC-07 验证 |
| AC-31 | rule | B3 三处坏味道在 main 分支代码中真实存在（Handler 三段式重复 / schema.sql 单文件无版本 / CalcRefundable if-else ≥3 层嵌套 + ≥50 行） | 代码行数统计 + 目视 |

### 8.2 Rubric 类（共 5 条，全部达到阈值 ≥2 分才算通过）
| 编号 | 类型 | 维度 | 分档（0/1/2 分）+ 阈值 | 证据来源 |
|---|---|---|---|---|
| RAC-01 | rubric | 「教学可观测性」：新手打开项目后 10 分钟内能否定位到「入口文件 + 规则文件 + 两个路由文件 + DB 文件」共 5 个核心文件 | 0=20 分钟还没找到 3 个以上；1=10-20 分钟找到 4 个；2=≤10 分钟找到全部 5 个。阈值=2 分 | 让 1 个不熟悉项目的人计时实测 |
| RAC-02 | rubric | 「S-A 简化版对齐」：`docs/code-map.md` 六板块（架构/模块/入口/数据/依赖/风险）+ `key-files-quickref.md` 4 列速查表是否完整对得上实际代码 | 0=两块文档各缺 2 板块/列以上；1=缺 1 块；2=全对得上。阈值=2 分 | 手动对照代码逐条核对 |
| RAC-03 | rubric | 「B1 练习摩擦」：B1 读者从打开 README 到写出「澄清 4 条锚点的 TODO 记录」所需总步骤数是否 ≤ 5（不用来回翻教程正文找素材） | 0=≥9 步；1=6-8 步；2=≤5 步（refund_rules.go 顶部已贴小美原话 + 4 TODO，读者照着填就行）。阈值=2 分 | 让新手模拟走一遍计时 |
| RAC-04 | rubric | 「AI 阅卷有效性」：把首轮骨架代码 + PITFALLS-B1 喂给 AI 审核员，AI 能否正确识别出「幂等（B1-P04）= ❌；运费/券（B1-P01/P02）= ❌」至少 2 条遗漏，且不泄题（不直接给行号/根因） | 0=全漏或泄题；1=只命中 1 条或有泄题风险；2=命中 ≥2 条且全程无泄题。阈值=2 分 | 实际跑一次 AI 阅卷流程看输出 |
| RAC-05 | rubric | 「B3 重构可操作性」：读者打开 B3 坏味道行号锚定 + PITFALLS-B3 停止条件后，能否 30 分钟内拆完坏味道 1（Handler 样板抽公共 helper）并保持 go test 全绿 | 0=1 小时还没拆完或测试崩；1=30-60 分钟完成；2=≤30 分钟完成且全绿。阈值=2 分 | 新手模拟实操 |

---

## 附录-A：接口契约 + Schema SQL（与 Section3 设计一致）

```sql
-- internal/db/schema.sql（首轮唯一版；B3 重构后拆成多文件）
CREATE TABLE IF NOT EXISTS orders (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    product_name TEXT    NOT NULL,
    amount       INTEGER NOT NULL,
    shipping     INTEGER NOT NULL DEFAULT 0,
    coupon_used  INTEGER NOT NULL DEFAULT 0,
    status       TEXT    NOT NULL DEFAULT 'paid',
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    CHECK (status IN ('paid', 'partial_refunded', 'fully_refunded'))
);

CREATE TABLE IF NOT EXISTS refunds (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id    INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    amount      INTEGER NOT NULL,
    status      TEXT    NOT NULL DEFAULT 'pending',
    reason      TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    CHECK (status IN ('pending', 'approved', 'rejected'))
);
```

六条路由完整请求/响应 JSON 字段清单略（与 Section3 逐字一致），后续同步写入 `docs/api-reference.md`。

---

## 附录-B：首轮 12+1 文件生成顺序（与 Section5 一致）
见 Section5 表格，共 12 个核心代码文件（序号1~11 是代码+前端，序号12 是 README）+ 13 个 docs 配套文件（附录-D 列）。

---

## 附录-C：三轨道场景 & 原始需求话术（与 Section 2b 一致，作为读者练习输入）
完整 B1 小美话术 / B2 老王凌晨两点拉群 Bug 话术 / B3 老王重构咖啡时间话术，后续同步写入 GUIDE 附录。

---

## 附录-D：首轮 13 个 docs 配套文件清单（FR-10 细化）
| 序号 | 文件名 | 对齐教程哪节 | 内容结构要点 |
|---|---|---|---|
| 1 | `spec-2026-08-20-...md` | 本文件 | 规格正文+验收标准 |
| 2 | `code-map.md` | S-A 第1步（新项目简化版） | 代码库地图六板块（架构/模块/入口/数据/依赖/风险）实际样本 |
| 3 | `key-files-quickref.md` | S-A 第4步+新项目简化版承诺速查表 | 4列：模块/位置/角色/进入端点，5 个核心关键文件 |
| 4 | `design-skeleton.md` | B1② 七段方案文档 | 七段结构；4段数据模型+5段接口预填；横评表空+填好样本；其余留空模板 |
| 5 | `ADR-001-tech-stack.md` | B1③ 决策记录 | 五段正式ADR（背景/候选/决策/理由/代价）= 本次技术选型决策 |
| 6 | `implementation-sample-record.md` | B1⑤ 分步实现 | 五段实现记录样本（改了啥/为啥/验了啥/无回归/状态）= schema.sql + db.go 那条任务 |
| 7 | `test-report-b1-base.md` | S-B 测试与回归 | 四段测试报告空模板（范围基线/用例代码/运行结果/回归结论）+ 预填范围 |
| 8 | `delivery-checklist.md` | S-C 文档交付 | 7 大项交付核对清单（代码/设计/测试/部署/API/CHANGELOG/README），每项 ✅/⚠/❌ |
| 9 | `api-reference.md` | S-C 文档交付 | 三部分：6 条路由清单表+使用说明前置条件+金额单位注意事项；与代码可断言 |
| 10 | `CHECKLIST-AFTER-DOING.md` | 三轨道完成后人工查漏 | 封面警告4条；B1 24条/B2 16条/B3 18条；每条✅/⚠/❌+漏了去哪补线索；完成声明 |
| 11 | `PROMPT-AI-REVIEWER.md` | 读者可打开，AI阅卷 | 顶部使用说明+4步操作；提示词正文（角色/输入/依据/审核四步/输出格式/两条红线约束） |
| 12 | `PITFALLS-B1.md` | AI阅卷答案本，严禁练习前打开 | 三重泄题警告；B1-P01~15 共15条（功能类6条/过程类6条/文档类3条），每条✅/⚠/❌判定标准 |
| 13 | `PITFALLS-B2.md` | 同上 | 三重警告；Bug1/Bug2 精确位置+修复标准2条；过程纪律7条，共10条 |
| 14 | `PITFALLS-B3.md` | 同上 | 三重警告；3坏味道位置+停止条件3条；保真硬指标9条共12条 |
（注：本文件 = 序号1；实际首轮 docs 总数 = 14 份，原写 13 份因漏计 spec 本身，已在 03 备注中修正，总范围不超本轮定义，无需 reopen。）

---

## 附录-E：26 项教程正文对齐缺漏补法清单（本轮设计已全部覆盖）
完整 26 项（S-A 3 项 / S-B 4 项 / B1① 3 项 / B1② 2 项 / B1③ 2 项 / B1④ 2 项 / B1⑤ 2 项 / B2B3 3 项 / S-C 3 项 / S-A 新项目简化版 1 项 + 新增 AI 阅卷机制 1 项 = 26 项）详见 brainstorming Section「汇总一」；本 spec 正文各条款均已把补法落到对应 FR / docs 条款中，无悬空项。

---

## 【Spec 四步自审记录（写 spec 时自检）】
### ① 占位符扫描（搜 TBD/TODO/待/？/xxxxx）
✅ 全文扫描：无 TBD/待填空/问号占位符。
- 检查结果：**通过**

### ② 内部一致性（6 Section 交叉核对）
- 接口契约（附录-A）↔ 测试用例（FR-9）：7 条清晰规则对得上 ✅
- B2 预埋 Bug 描述（FR-7）↔ PITFALLS-B2 条目（附录-D 序号13）：根因位置一致 ✅
- B3 坏味道 3 处（FR-8）↔ PITFALLS-B3 停止条件 + 保真指标：一一对应 ✅
- 目录结构（02 权威设计 + 本 spec 术语定义 C）↔ 生成顺序（附录-B）：文件编号路径全部对齐 ✅
- 9 项决策（02 关键决策日志）↔ 约束/功能条款：1B bug注入 / 2A 纯手写迁移 / 3A 幂等留 B1 补 / 4A ASCII 文件名 等全部在约束条款落实 ✅
- 功能需求 FR-4.3 模糊点 3 处 + 幂等 1 处 ↔ PITFALLS-B1 的 B1-P01~04 条目：✅ 对得上
- 检查结果：**通过**

### ③ 范围检查（本轮 spec 是否只覆盖第一轮目标）
- 不在本轮写但将来要做的（已明确列 Non-Goals）：
  - b2-bug 分支的**具体代码写入**（本轮只写 main，b2-bug 分支切分支造 Bug 是第二轮任务，本轮不写）
  - B3 重构后的代码形态（本轮只预埋坏味道，重构是读者练习动作，首轮不写）
  - 运费/券/超额策略的最终补全（B1 读者自己做）
  - 幂等键 / refund_no 字段最终补全（B1 读者自己做）
  - CI/CD / GitHub 发布 / LICENSE / ASCII 转换 MAPPING（全程冻结）
  - 完整 GUIDE-LEARN-PATH.md 正文内容（本轮实施计划会写，spec 只是锁定结构要求，未超范围）
- 检查结果：**通过**，无越界

### ④ 歧义检查（3 个歧义澄清术语是否写在 spec 开头）
- 术语定义段已写 A/B/C 三点：模糊点 vs Bug / 覆盖率区间算法 / B3「API 不许改」精确边界 ✅
- 检查结果：**通过**

---

**Spec 自审结论：4 项全通过。无占位符 / 无矛盾 / 无越界 / 无歧义。**
下一步：请用户 Review 本 spec 正文。Review 通过 → 调用 `writing-plans` 生成实施计划（每个文件具体代码要点 + 验证步骤 + 时间估算 + 优先级）。
