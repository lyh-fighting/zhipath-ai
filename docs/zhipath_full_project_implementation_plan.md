# 知途 AI 全项目实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use `SP-subagent-driven-development`（推荐）或 `SP-executing-plans` 按任务逐项实现。每完成一个任务必须运行验收命令并提交，禁止跳过失败项继续开发。

**目标：** 从空目录构建可本地完整运行的知途 C 端应用，实现微信小程序、Hertz API、LangGraph 情感/职业/决策分析、MBTI 画像、OCR、MySQL/Redis/Qdrant/MinIO、异步任务、n8n 回访及安全治理，并允许后续仅通过配置替换为外部云服务。

**架构：** Go Hertz 是公网入口和 MySQL 业务事实的唯一写入者；Python LangGraph 服务只执行分析并返回结构化结果；OCR 服务只识别文件；Worker 通过 transactional outbox 幂等同步 Qdrant、触发 n8n 和生成报告。本地依赖全部使用 Docker Compose，云上通过存储与服务适配接口替换。

**技术栈：** Go 1.23、Hertz、MySQL 8.4、Redis 7.4、Python 3.12、FastAPI、LangGraph、Qdrant、MinIO、PaddleOCR、n8n、Taro React TypeScript、Docker Compose、pytest、Go test、Vitest。

**必读规格：**

- `docs/langgraph_consult_agent_architecture.md`
- `docs/consult_agent_project_bootstrap_guide.md`

---

## 一、强制执行规则

1. 每个任务遵循：写失败测试 → 确认失败 → 最小实现 → 测试通过 → 提交。
2. 不使用 LangChain，不安装任何 `langchain-*` 包。
3. 不把客户端传入的 `user_id` 当作可信身份；用户身份必须来自登录态。
4. AI、OCR、Worker 不得直接修改用户、订单、画像和消息等业务事实。
5. MySQL 是事实源；Qdrant、Redis、MinIO 索引和 n8n 触达均可重建。
6. MySQL 与异步系统间使用 `outbox_events`，禁止请求内直接双写 MySQL 和 Qdrant。
7. 深度分析必须读取当前有效 MBTI；缺失时先提示测试，用户跳过后只能生成基础分析。
8. 所有敏感文本、OCR 结果和风险详情必须加密存储或受对象权限保护。
9. 本地所有后端依赖由 Docker Compose 启动；微信小程序使用微信开发者工具。
10. 外部模型、微信支付和微信登录没有真实凭证时，使用明确的 Mock Adapter，不能静默伪造成功。

## 二、最终目录

```text
zhipath-ai/
├── apps/miniapp/
├── services/
│   ├── api-gateway/
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   ├── migrations/
│   │   ├── tests/
│   │   └── Dockerfile
│   ├── ai-agent-service/
│   │   ├── app/
│   │   ├── tests/
│   │   ├── pyproject.toml
│   │   └── Dockerfile
│   ├── ocr-service/
│   │   ├── app/
│   │   ├── tests/
│   │   ├── pyproject.toml
│   │   └── Dockerfile
│   └── worker-service/
│       ├── app/
│       ├── tests/
│       ├── pyproject.toml
│       └── Dockerfile
├── packages/
│   ├── api-contracts/
│   └── api-client/
├── infra/
│   ├── docker-compose.yml
│   ├── mysql/init/
│   ├── qdrant/
│   ├── n8n/workflows/
│   ├── nginx/
│   └── scripts/
├── tests/e2e/
├── docs/
├── .env.example
├── Makefile
└── README.md
```

---

## Task 1：初始化 Monorepo 和质量门禁

**文件：**

- 创建：`.gitignore`
- 创建：`.env.example`
- 创建：`Makefile`
- 创建：`README.md`
- 创建：`.github/workflows/ci.yml`
- 创建：上述最终目录中的空模块目录

- [ ] 创建目录、初始化 Git，并将默认分支设为 `main`。
- [ ] 生成根目录配置；`.env`、密钥、上传文件、数据库卷、模型文件不得入库。
- [ ] `Makefile` 提供 `bootstrap`、`infra-up`、`infra-down`、`migrate`、`init-qdrant`、`test`、`lint`、`e2e`。
- [ ] CI 分别运行 Go test、pytest、前端 Vitest、格式检查和依赖锁检查。
- [ ] 运行 `git status --short`，确认只有预期文件。
- [ ] 提交：`chore(infra): initialize zhipath monorepo`

**验收：**

```bash
make help
git status --short
```

预期：`make help` 列出所有目标；无 `.env`、卷数据和密钥被跟踪。

---

## Task 2：建立共享 API 契约

**文件：**

- 创建：`packages/api-contracts/openapi.yaml`
- 创建：`packages/api-contracts/schemas/*.json`
- 创建：`packages/api-client/src/index.ts`
- 创建：`packages/api-client/tests/contracts.test.ts`

- [ ] 先写契约测试，验证 OpenAPI 至少包含健康检查、登录、画像、MBTI、会话、消息、文件、OCR、报告、反馈接口。
- [ ] 定义统一响应：`code`、`message`、`data`、`trace_id`。
- [ ] 定义错误码：认证失败、资源越权、参数错误、MBTI 缺失、模型超时、OCR 失败、危机转人工。
- [ ] 所有公网请求不包含可信 `user_id`；用户身份由 Authorization token 推导。
- [ ] 生成 TypeScript API client 和 Go/Python 数据结构，生成物纳入 CI 一致性检查。
- [ ] 提交：`feat(contract): define public and internal APIs`

**验收：**

```bash
pnpm --dir packages/api-client test
npx @redocly/cli lint packages/api-contracts/openapi.yaml
```

---

## Task 3：Docker-first 基础设施

**文件：**

- 创建：`infra/docker-compose.yml`
- 创建：`infra/mysql/init/000-default-tenant.sql`
- 创建：`infra/qdrant/init_collections.py`
- 创建：`infra/scripts/wait-for-stack.sh`
- 创建：各服务 `Dockerfile`

- [ ] Compose 固定版本启动 MySQL、Redis、Qdrant、MinIO、n8n、API、AI、OCR、Worker。
- [ ] 为数据库、缓存、向量库和应用服务配置健康检查。
- [ ] MinIO 初始化任务创建 `zhipath-ai` bucket。
- [ ] Qdrant 初始化脚本读取 `EMBEDDING_DIMENSION`，幂等创建三个 collection 和 payload index。
- [ ] 所有数据服务使用 named volume。
- [ ] 禁止在 compose 中放生产凭证；本地密码只能来自 `.env`。
- [ ] 提交：`chore(infra): add docker development stack`

**验收：**

```bash
cp .env.example .env
docker compose -f infra/docker-compose.yml config
docker compose -f infra/docker-compose.yml up -d --build
./infra/scripts/wait-for-stack.sh
```

预期：所有容器为 healthy 或 completed。

---

## Task 4：MySQL Migration、种子数据与回滚

**文件：**

- 创建：`services/api-gateway/migrations/000001_init_schema.{up,down}.sql`
- 创建：`000002_mbti_profile.{up,down}.sql`
- 创建：`000003_agent_memory.{up,down}.sql`
- 创建：`000004_order_payment.{up,down}.sql`
- 创建：`000005_risk_followup_outbox.{up,down}.sql`
- 创建：`services/api-gateway/internal/repository/migration_test.go`

- [ ] 按架构文档创建全部业务表、索引和外键。
- [ ] 所有业务唯一键使用租户复合唯一键，例如 `UNIQUE(tenant_id, user_id)`。
- [ ] `user_profiles.current_mbti_result_id` 指向当前用户确认结果。
- [ ] `agent_runs`、`decision_records` 保存 `mbti_result_id` 和 `mbti_snapshot`。
- [ ] 创建 `outbox_events` 及轮询索引。
- [ ] 创建默认租户 `default_consumer`。
- [ ] 测试 migration up、down、再次 up。
- [ ] 提交：`feat(database): add versioned mysql schema`

**验收：**

```bash
make migrate
make migration-test
```

预期：全部表存在；回滚再升级后结构一致。

---

## Task 5：Go Hertz 服务基础与资源适配器

**文件：**

- 创建：`services/api-gateway/cmd/server/main.go`
- 创建：`internal/config/config.go`
- 创建：`internal/platform/mysql.go`
- 创建：`internal/platform/redis.go`
- 创建：`internal/platform/object_store.go`
- 创建：`internal/middleware/{trace,errors,recovery}.go`
- 创建：对应 `*_test.go`

- [ ] 配置加载失败必须返回明确错误，禁止默认连接生产资源。
- [ ] 定义 `RelationalStore`、`CacheStore`、`ObjectStore` 接口。
- [ ] 本地对象存储使用 S3-compatible MinIO adapter。
- [ ] 增加 `/healthz` 和 `/readyz`；readiness 检查 MySQL、Redis、AI Service。
- [ ] 为每个请求生成或透传 `trace_id`、`request_id`。
- [ ] 提交：`feat(api): bootstrap hertz gateway`

**验收：**

```bash
cd services/api-gateway
go test ./...
curl -f http://localhost:8080/healthz
curl -f http://localhost:8080/readyz
```

---

## Task 6：认证、授权和服务间鉴权

**文件：**

- 创建：`internal/auth/{service,token,wechat,mock}.go`
- 创建：`internal/middleware/auth.go`
- 创建：`internal/middleware/internal_auth.go`
- 创建：`internal/handler/auth.go`
- 测试：`internal/auth/*_test.go`

- [ ] 本地使用 Mock 微信登录，生产 adapter 调微信接口。
- [ ] Access token 和 refresh token 分离，refresh token 存哈希。
- [ ] 公网接口从 token 获取 `tenant_id/user_id`，忽略请求体中的同名字段。
- [ ] Go 调 AI/OCR 使用内部 service token；服务拒绝无凭证调用。
- [ ] 测试跨用户访问会话、文件、MBTI 返回 403。
- [ ] 提交：`feat(auth): add user and service authentication`

---

## Task 7：用户画像与 MBTI 全流程

**文件：**

- 创建：`internal/domain/profile/*.go`
- 创建：`internal/domain/mbti/*.go`
- 创建：`internal/handler/profile.go`
- 创建：`internal/handler/mbti.go`
- 创建：`apps/miniapp/src/pages/profile/*`
- 创建：`apps/miniapp/src/pages/mbti/*`
- 测试：Go handler/service tests、前端 Vitest

- [ ] 实现画像查询和更新。
- [ ] 实现 MBTI 历史列表、当前结果、手工提交、截图提交、用户确认和切换当前结果。
- [ ] MBTI 维度保存“极性 + 百分比”，不能仅保存无方向数字。
- [ ] 截图识别结果必须由用户确认后才能成为当前结果。
- [ ] 深度分析入口无有效 MBTI 时展示测试链接和上传入口。
- [ ] 用户跳过时写入 `analysis_level=basic`，不得伪装为深度分析。
- [ ] 提交：`feat(profile): add mbti profile lifecycle`

**验收：**

```bash
cd services/api-gateway && go test ./internal/domain/profile/... ./internal/domain/mbti/...
pnpm --dir apps/miniapp test
```

---

## Task 8：会话、消息与幂等

**文件：**

- 创建：`internal/domain/conversation/*.go`
- 创建：`internal/handler/conversation.go`
- 创建：`internal/middleware/idempotency.go`
- 测试：`internal/domain/conversation/*_test.go`

- [ ] 创建会话、发送消息、游标分页查询消息。
- [ ] `request_id` 建立唯一约束，同一请求重试返回相同结果。
- [ ] 事务内写入用户消息、更新会话时间和创建 Agent run。
- [ ] 失败时不留下半完成消息。
- [ ] 提交：`feat(conversation): add idempotent messaging`

---

## Task 9：Python AI Service 基础和 Provider 接口

**文件：**

- 创建：`services/ai-agent-service/app/main.py`
- 创建：`app/config.py`
- 创建：`app/providers/{base,mock,openai_compatible,anthropic}.py`
- 创建：`app/stores/{cache,vector,checkpoint}.py`
- 创建：`tests/providers/*`
- 创建：`Dockerfile`

- [ ] 定义 `ModelProvider` 和结构化输出协议。
- [ ] Mock provider 返回确定性结果，供本地 E2E 使用。
- [ ] OpenAI-compatible adapter 支持 DeepSeek/OpenAI base URL 切换。
- [ ] fallback 只处理超时、限流和 5xx，不吞掉参数及安全错误。
- [ ] Redis checkpoint 以 `tenant_id:user_id:conversation_id` 为 thread key，配置 TTL。
- [ ] `/healthz`、`/readyz` 和内部鉴权生效。
- [ ] 提交：`feat(ai): bootstrap langgraph service`

---

## Task 10：LangGraph 主图与恢复能力

**文件：**

- 创建：`app/graph/state.py`
- 创建：`app/graph/builder.py`
- 创建：`app/graph/routes.py`
- 创建：`app/graph/nodes/*.py`
- 测试：`tests/graph/test_routes.py`
- 测试：`tests/graph/test_resume.py`

- [ ] 实现附件检查、上下文加载、意图路由、风险检测、MBTI 检查、画像完整性、检索、Agent 分派、质量检查、修复、最终化。
- [ ] 深度分析必须校验 `current_mbti_result_id`。
- [ ] 每次执行生成不可变 `mbti_snapshot`。
- [ ] 修复循环最多两次。
- [ ] 在中断后使用同一 thread key 恢复执行，不能重复调用已完成工具。
- [ ] 提交：`feat(ai): implement consultation graph`

---

## Task 11：情感、职业和决策 Agent

**文件：**

- 创建：`app/agents/emotion.py`
- 创建：`app/agents/career.py`
- 创建：`app/agents/decision_coach.py`
- 创建：`app/prompts/*.py`
- 创建：`app/schemas/analysis.py`
- 测试：`tests/agents/*`

- [ ] 统一输出问题本质、现实情况、MBTI 参考、选项、风险、推荐、行动计划。
- [ ] MBTI 是现实情况之后的主要分析依据。
- [ ] 情感 Agent 禁止医学诊断；职业 Agent 禁止保证 offer、薪资和晋升。
- [ ] 混合问题并行执行情感和职业分析，由决策教练合并。
- [ ] Prompt 版本写入每次运行结果。
- [ ] 提交：`feat(ai): add consultation specialists`

---

## Task 12：质量、安全与危机响应

**文件：**

- 创建：`app/safety/{rules,classifier,response}.py`
- 创建：`app/quality/{scorer,grounding}.py`
- 创建：`internal/domain/risk/*.go`
- 创建：`infra/config/emergency_resources.zh-CN.json`
- 测试：`tests/safety/*`、Go risk tests

- [ ] 检测自伤、自杀、家暴、未成年人、暴力威胁。
- [ ] critical 风险不进入普通生成链，返回安全响应并创建人工介入事件。
- [ ] 定义人工值守 SLA、离线降级和地区化紧急联系方式。
- [ ] 质量评分覆盖事实依据、可执行性、安全、结构完整性。
- [ ] 建立至少 50 条安全回归样本。
- [ ] 提交：`feat(safety): add crisis and quality gates`

---

## Task 13：知识库导入与 RAG

**文件：**

- 创建：`services/worker-service/app/jobs/knowledge_ingest.py`
- 创建：`ai-agent-service/app/rag/{splitter,retriever,reranker}.py`
- 创建：`infra/qdrant/init_collections.py`
- 创建：`tests/fixtures/knowledge/*`
- 测试：RAG unit/integration tests

- [ ] 文档按领域、标题和段落切分，chunk 带版本和审核状态。
- [ ] embedding provider 可配置；本地测试使用确定性 fake embedding。
- [ ] 仅检索 `status=published`、正确 tenant/domain/version 的内容。
- [ ] MySQL 保存 chunk 元数据，通过 Outbox 同步 Qdrant。
- [ ] 同一事件重复消费不产生重复 point。
- [ ] 提交：`feat(rag): add versioned knowledge retrieval`

---

## Task 14：文件上传和 OCR

**文件：**

- 创建：Go `internal/domain/file/*.go`
- 创建：OCR `app/engine/{base,mock,paddle}.py`
- 创建：OCR `app/preprocess/image.py`
- 创建：`tests/fixtures/ocr/*`
- 测试：上传、权限、OCR、低置信度测试

- [ ] 使用预签名 URL 上传 MinIO，限制类型、大小和哈希。
- [ ] OCR 只能使用短期签名 URL 读取文件。
- [ ] MBTI 截图解析出类型、A/T 和带方向的维度比例。
- [ ] 低置信度结果必须要求用户确认。
- [ ] 原始文件和 OCR 文本支持用户删除。
- [ ] 提交：`feat(ocr): add secure document extraction`

---

## Task 15：Outbox Worker、Qdrant 同步和 n8n

**文件：**

- 创建：`worker-service/app/outbox/{poller,handlers}.py`
- 创建：`worker-service/app/clients/{qdrant,n8n}.py`
- 创建：`infra/n8n/workflows/*.json`
- 测试：`worker-service/tests/test_outbox.py`

- [ ] 使用行锁批量获取 pending 事件。
- [ ] 基于 `event_id` 幂等处理。
- [ ] 失败采用指数退避，达到上限进入 dead 状态并报警。
- [ ] 实现用户记忆、会话摘要、知识 chunk 三类 Qdrant handler。
- [ ] 实现 3 天回访、7 天复盘、报告完成三个 n8n 工作流。
- [ ] 提交：`feat(worker): add reliable event processing`

---

## Task 16：订单、会员和支付 Adapter

**文件：**

- 创建：`internal/domain/order/*.go`
- 创建：`internal/domain/payment/{base,mock,wechat}.go`
- 创建：`internal/handler/order.go`
- 测试：订单状态机和回调幂等测试

- [ ] 产品支持会员、深度报告两类。
- [ ] 本地 Mock 支付必须显式标记 `provider=mock`。
- [ ] 微信回调验签、金额校验、幂等更新。
- [ ] 付款成功后发 Outbox 事件，不直接调用下游。
- [ ] 提交：`feat(payment): add order and payment lifecycle`

---

## Task 17：深度报告生成

**文件：**

- 创建：`worker-service/app/jobs/report.py`
- 创建：`ai-agent-service/app/schemas/report.py`
- 创建：Go `internal/domain/report/*.go`
- 创建：小程序 `pages/report/*`
- 测试：报告快照和权限测试

- [ ] 报告必须记录现实依据、MBTI 测试版本、MBTI 分析、风险、30/90/180 天行动。
- [ ] 报告任务可重试且不重复扣权益。
- [ ] 报告文件写 MinIO，用户只能访问自己的预签名 URL。
- [ ] 完成后通过 Outbox 触发 n8n 通知。
- [ ] 提交：`feat(report): add deep consultation reports`

---

## Task 18：微信小程序完整用户流程

**文件：**

- 创建：`apps/miniapp/src/pages/{index,login,chat,profile,mbti,report,orders,settings}/`
- 创建：`apps/miniapp/src/services/api.ts`
- 创建：`apps/miniapp/src/stores/*.ts`
- 测试：Vitest 页面状态测试

- [ ] 登录、画像补全、MBTI 测试引导和截图上传。
- [ ] 基础咨询、深度分析入口、流式/轮询状态展示。
- [ ] 风险场景使用专用安全页面。
- [ ] 历史会话、报告、订单和数据设置。
- [ ] 网络重试复用相同 `request_id`。
- [ ] 提交：`feat(miniapp): implement consumer experience`

---

## Task 19：隐私、同意、导出和删除

**文件：**

- 创建 migration：`000006_privacy_consent.*.sql`
- 创建：Go `internal/domain/privacy/*.go`
- 创建：小程序 `pages/settings/privacy.tsx`
- 测试：导出、撤回、删除传播测试

- [ ] 用户单独同意存储 MBTI、聊天截图和长期记忆。
- [ ] 支持撤回长期记忆授权。
- [ ] 支持导出画像、MBTI、会话和报告。
- [ ] 注销时删除/匿名化 MySQL 数据，并通过 Outbox 删除 Qdrant point 和 MinIO 文件。
- [ ] 审计日志不得记录正文和凭证。
- [ ] 提交：`feat(privacy): add consent and data lifecycle`

---

## Task 20：可观测性、限流和安全加固

**文件：**

- 创建：`infra/observability/*`
- 创建：Go/Python metrics 和 tracing middleware
- 创建：`infra/nginx/nginx.conf`
- 测试：限流、超时、熔断集成测试

- [ ] 指标覆盖 QPS、p95/p99、模型耗时、token、风险命中、Outbox 延迟。
- [ ] 日志通过同一 `trace_id` 串联。
- [ ] 用户/IP/接口三级限流。
- [ ] 模型、OCR、n8n 设置超时、重试和熔断。
- [ ] 上传文件扫描类型并阻止路径穿越。
- [ ] 提交：`feat(observability): add production guardrails`

---

## Task 21：全链路 E2E 和 Agent 评估

**文件：**

- 创建：`tests/e2e/*.py`
- 创建：`tests/evaluation/golden_cases.jsonl`
- 创建：`tests/evaluation/run_eval.py`
- 创建：`infra/scripts/e2e.sh`

- [ ] 覆盖注册、MBTI、基础咨询、深度分析、OCR、RAG、报告、支付 Mock、回访事件。
- [ ] 验证没有 MBTI 时只能得到基础分析。
- [ ] 验证切换 MBTI 后历史报告仍保留旧快照。
- [ ] 验证危机文本不返回普通行动建议。
- [ ] 黄金集至少包含情感 50 条、职业 50 条、混合 20 条、安全 50 条。
- [ ] 提交：`test(e2e): cover complete consultation journey`

**验收：**

```bash
make test
make e2e
make eval
```

---

## Task 22：云资源替换演练

**文件：**

- 创建：`docs/cloud-migration.md`
- 创建：`infra/docker-compose.external.yml`
- 创建：`infra/scripts/verify-external-services.sh`

- [ ] 使用 Compose override 停止本地 MySQL/Redis/Qdrant/MinIO/n8n。
- [ ] 仅修改环境变量连接外部测试资源。
- [ ] 执行 migration、Qdrant 初始化和 E2E。
- [ ] 验证领域代码和 LangGraph 节点零修改。
- [ ] 记录备份、回滚、DNS/连接池、TLS 和凭证轮换步骤。
- [ ] 提交：`docs(infra): verify cloud service portability`

---

## Task 23：发布验收

- [ ] 全部依赖和镜像固定版本，提交 `uv.lock`、`go.sum`、`pnpm-lock.yaml`。
- [ ] `docker compose config` 无错误。
- [ ] 所有 migration 可升级和回滚。
- [ ] 所有健康检查通过。
- [ ] 单元、集成、E2E、安全和评估测试通过。
- [ ] 仓库不存在密钥、真实用户数据和上传文件。
- [ ] 人工危机响应流程完成一次演练。
- [ ] 发布文档包含启动、备份、恢复、回滚和云资源替换。
- [ ] 提交：`chore(release): prepare zhipath v0.1.0`

## 三、最终完成标准

只有同时满足以下条件，AI 才能声明项目完成：

```text
1. 从全新目录按文档可生成项目。
2. 一条命令启动全部后端容器。
3. 小程序可完成完整 C 端用户流程。
4. 深度分析强提醒 MBTI，并保存当前结果、历史结果和分析快照。
5. 情感、职业、混合分析都有结构化结果和安全控制。
6. MySQL、Qdrant、MinIO、n8n 数据链路可验证。
7. Outbox 重试不会产生重复向量或重复通知。
8. 用户可导出和删除自己的敏感数据。
9. 断电或服务重启后，LangGraph 会话可恢复。
10. 替换外部云资源只修改配置和 adapter，不修改领域逻辑。
11. 所有自动化测试通过。
12. 没有未完成占位符、静默降级和伪造成功。
```
