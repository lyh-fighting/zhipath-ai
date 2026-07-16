# 知途 AI（ZhiPath）项目开发任务追踪表

> 项目代号：`zhipath` / 仓库：`zhipath-ai` / 架构：monorepo + 多服务独立启动
> 技术栈：Go 1.23、Hertz、MySQL 8.4、Redis 7.4、Python 3.12、FastAPI、LangGraph、Qdrant、MinIO、PaddleOCR、n8n、Taro React TypeScript、Docker Compose、pytest、Go test、Vitest

## 一、权威文档关系

| 文档 | 作用 | 优先级 |
| --- | --- | --- |
| `zhipath_full_project_implementation_plan.md` | **逐项开发与验收顺序（Task 1-23）**，含强制规则与最终完成标准 | 最高（执行依据） |
| `consult_agent_project_bootstrap_guide.md` | 项目创建、目录、服务初始化、启动顺序 | 参考 |
| `langgraph_consult_agent_architecture.md` | 业务与数据规则、LangGraph 图、MySQL/Qdrant 表设计 | 参考 |

完整实现时三份文档必须同时读取。本追踪表与实施计划 Task 1-23 一一对应，并已同步到 WorkBuddy TaskList。

## 二、强制执行规则（不可违背）

1. 每个任务：写失败测试 → 确认失败 → 最小实现 → 测试通过 → 提交
2. 不使用 LangChain，不安装任何 `langchain-*` 包
3. 不把客户端传入的 `user_id` 当可信身份；身份必须来自登录态
4. AI/OCR/Worker 不得直接修改用户、订单、画像、消息等业务事实
5. MySQL 是事实源；Qdrant/Redis/MinIO 索引和 n8n 触达均可重建
6. MySQL 与异步系统间用 `outbox_events`，禁止请求内双写 MySQL 和 Qdrant
7. 深度分析必须读取当前有效 MBTI；缺失时先提示测试，用户跳过只生成基础分析
8. 敏感文本、OCR 结果、风险详情必须加密存储或受对象权限保护
9. 本地所有后端依赖由 Docker Compose 启动；小程序用微信开发者工具
10. 外部模型/微信支付/微信登录无真实凭证时用明确 Mock Adapter，不能静默伪造成功

## 三、Task 1-23 追踪表

- [ ] **Task 1 初始化 Monorepo 与质量门禁**（TaskList #2）
  - 文件：`.gitignore`、`.env.example`、`Makefile`、`README.md`、`.github/workflows/ci.yml`、空模块目录
  - Makefile：`bootstrap/infra-up/infra-down/migrate/init-qdrant/test/lint/e2e`
  - 验收：`make help`、`git status --short`
  - 提交：`chore(infra): initialize zhipath monorepo`

- [ ] **Task 2 建立共享 API 契约**（TaskList #1）
  - 文件：`packages/api-contracts/openapi.yaml`、`schemas/*.json`、`packages/api-client/src/index.ts`、`tests/contracts.test.ts`
  - 统一响应 `code/message/data/trace_id`；错误码：认证失败/越权/参数错误/MBTI缺失/模型超时/OCR失败/危机转人工
  - 公网请求不含可信 `user_id`；生成 TS client + Go/Python 结构入 CI
  - 验收：`pnpm --dir packages/api-client test` + `npx @redocly/cli lint`
  - 提交：`feat(contract): define public and internal APIs`

- [ ] **Task 3 Docker-first 基础设施**（TaskList #3）
  - 文件：`infra/docker-compose.yml`、`infra/mysql/init/000-default-tenant.sql`、`infra/qdrant/init_collections.py`、`infra/scripts/wait-for-stack.sh`、各服务 `Dockerfile`
  - MySQL/Redis/Qdrant/MinIO/n8n/API/AI/OCR/Worker 固定版本 + 健康检查；MinIO 建 `zhipath-ai` bucket；Qdrant 幂等建 3 collection + payload index
  - 验收：`cp .env.example .env && docker compose config && up -d --build && wait-for-stack.sh` 全 healthy
  - 提交：`chore(infra): add docker development stack`

- [ ] **Task 4 MySQL 迁移与种子数据**（TaskList #4）
  - 文件：`migrations/000001_init_schema`、`000002_mbti_profile`、`000003_agent_memory`、`000004_order_payment`、`000005_risk_followup_outbox`（各含 up/down）、`migration_test.go`
  - 业务唯一键用租户复合键；`user_profiles.current_mbti_result_id` 指向当前确认结果；`agent_runs/decision_records` 存 `mbti_result_id+mbti_snapshot`；建 `outbox_events`+轮询索引；默认租户 `default_consumer`
  - 验收：`make migrate && make migration-test`（up/down/再up 一致）
  - 提交：`feat(database): add versioned mysql schema`

- [ ] **Task 5 Go Hertz 服务基础与资源适配器**（TaskList #5）
  - 文件：`cmd/server/main.go`、`internal/config`、`internal/platform/{mysql,redis,object_store}.go`、`internal/middleware/{trace,errors,recovery}.go` 及测试
  - 定义 `RelationalStore/CacheStore/ObjectStore` 接口；本地对象存储用 MinIO S3-compatible adapter；`/healthz`+`/readyz`（readiness 查 MySQL/Redis/AI）
  - 配置加载失败返回明确错误，禁默认连生产
  - 验收：`go test ./... && curl /healthz && curl /readyz`
  - 提交：`feat(api): bootstrap hertz gateway`

- [ ] **Task 6 认证授权与服务间鉴权**（TaskList #6）
  - 文件：`internal/auth/{service,token,wechat,mock}.go`、`internal/middleware/{auth,internal_auth}.go`、`internal/handler/auth.go` 及测试
  - 本地 Mock 微信登录；access/refresh token 分离，refresh 存哈希；公网接口从 token 取身份，忽略请求体同名字段；Go 调 AI/OCR 用内部 service token
  - 测试跨用户访问会话/文件/MBTI 返回 403
  - 提交：`feat(auth): add user and service authentication`

- [ ] **Task 7 用户画像与 MBTI 全流程**（TaskList #8）
  - 文件：`internal/domain/{profile,mbti}/*.go`、`internal/handler/{profile,mbti}.go`、`apps/miniapp/src/pages/{profile,mbti}/*` 及测试
  - MBTI 维度保存"极性+百分比"；截图识别结果必须用户确认才成当前结果；深度分析入口无 MBTI 时展示测试链接+上传入口；用户跳过写 `analysis_level=basic`
  - 验收：`go test ./internal/domain/profile/... ./internal/domain/mbti/... && pnpm --dir apps/miniapp test`
  - 提交：`feat(profile): add mbti profile lifecycle`

- [ ] **Task 8 会话消息与幂等**（TaskList #7）
  - 文件：`internal/domain/conversation/*.go`、`internal/handler/conversation.go`、`internal/middleware/idempotency.go` 及测试
  - 创建会话/发消息/游标分页；`request_id` 唯一约束，重试返回相同结果；事务内写消息+更新会话时间+建 Agent run；失败不残留
  - 提交：`feat(conversation): add idempotent messaging`

- [ ] **Task 9 Python AI Service 基础与 Provider 接口**（TaskList #9）
  - 文件：`app/{main,config}.py`、`app/providers/{base,mock,openai_compatible,anthropic}.py`、`app/stores/{cache,vector,checkpoint}.py`、`tests/providers/*`、`Dockerfile`
  - Mock provider 返回确定性结果；OpenAI-compatible adapter 支持 DeepSeek/OpenAI base URL；fallback 只处理超时/限流/5xx；Redis checkpoint 以 `tenant_id:user_id:conversation_id` 为 thread key+TTL；`/healthz`+`/readyz`+内部鉴权
  - 提交：`feat(ai): bootstrap langgraph service`

- [ ] **Task 10 LangGraph 主图与恢复能力**（TaskList #11）
  - 文件：`app/graph/{state,builder,routes}.py`、`app/graph/nodes/*.py`、`tests/graph/{test_routes,test_resume}.py`
  - 实现附件检查/上下文加载/意图路由/风险检测/MBTI检查/画像完整性/检索/Agent分派/质量检查/修复/最终化；深度分析校验 `current_mbti_result_id`；每次生成不可变 `mbti_snapshot`；修复循环最多2次；中断后同 thread key 恢复，不重复调已完成工具
  - 提交：`feat(ai): implement consultation graph`

- [ ] **Task 11 情感职业决策 Agent**（TaskList #13）
  - 文件：`app/agents/{emotion,career,decision_coach}.py`、`app/prompts/*.py`、`app/schemas/analysis.py`、`tests/agents/*`
  - 统一输出问题本质/现实情况/MBTI参考/选项/风险/推荐/行动计划；MBTI 是现实情况之后的主要分析依据；情感 Agent 禁医学诊断；职业 Agent 禁保证 offer/薪资/晋升；混合问题并行情感+职业，决策教练合并；Prompt 版本写入每次运行
  - 提交：`feat(ai): add consultation specialists`

- [ ] **Task 12 质量安全与危机响应**（TaskList #10）
  - 文件：`app/safety/{rules,classifier,response}.py`、`app/quality/{scorer,grounding}.py`、`internal/domain/risk/*.go`、`infra/config/emergency_resources.zh-CN.json` 及测试
  - 检测自伤/自杀/家暴/未成年人/暴力威胁；critical 风险不进普通生成链，返回安全响应+创建人工介入事件；人工值守 SLA/离线降级/地区化紧急联系方式；质量评分覆盖事实依据/可执行性/安全/结构完整性；≥50 条安全回归样本
  - 提交：`feat(safety): add crisis and quality gates`

- [ ] **Task 13 知识库导入与 RAG**（TaskList #12）
  - 文件：`worker-service/app/jobs/knowledge_ingest.py`、`ai-agent-service/app/rag/{splitter,retriever,reranker}.py`、`infra/qdrant/init_collections.py`、`tests/fixtures/knowledge/*`
  - 文档按领域/标题/段落切分，chunk 带版本和审核状态；embedding 可配置，本地用确定性 fake embedding；仅检索 `status=published`/正确 tenant/domain/version；MySQL 存 chunk 元数据，通过 Outbox 同步 Qdrant；重复消费不产生重复 point
  - 提交：`feat(rag): add versioned knowledge retrieval`

- [ ] **Task 14 文件上传与 OCR**（TaskList #15）
  - 文件：Go `internal/domain/file/*.go`、OCR `app/engine/{base,mock,paddle}.py`、`app/preprocess/image.py`、`tests/fixtures/ocr/*`
  - 预签名 URL 上传 MinIO，限类型/大小/哈希；OCR 只用短期签名 URL 读文件；MBTI 截图解析类型/A-T/带方向维度比例；低置信度必须用户确认；支持用户删除原始文件和 OCR 文本
  - 提交：`feat(ocr): add secure document extraction`

- [ ] **Task 15 Outbox Worker、Qdrant 同步与 n8n**（TaskList #14）
  - 文件：`worker-service/app/outbox/{poller,handlers}.py`、`app/clients/{qdrant,n8n}.py`、`infra/n8n/workflows/*.json`、`tests/test_outbox.py`
  - 行锁批量获取 pending 事件；基于 `event_id` 幂等；失败指数退避，达上限进 dead 状态并报警；实现用户记忆/会话摘要/知识 chunk 三类 Qdrant handler；实现 3 天回访/7 天复盘/报告完成 三个 n8n 工作流
  - 提交：`feat(worker): add reliable event processing`

- [ ] **Task 16 订单会员与支付 Adapter**（TaskList #16）
  - 文件：`internal/domain/order/*.go`、`internal/domain/payment/{base,mock,wechat}.go`、`internal/handler/order.go` 及测试
  - 产品支持会员/深度报告两类；本地 Mock 支付显式标记 `provider=mock`；微信回调验签/金额校验/幂等更新；付款成功发 Outbox 事件，不直接调下游
  - 提交：`feat(payment): add order and payment lifecycle`

- [ ] **Task 17 深度报告生成**（TaskList #18）
  - 文件：`worker-service/app/jobs/report.py`、`ai-agent-service/app/schemas/report.py`、Go `internal/domain/report/*.go`、小程序 `pages/report/*`
  - 报告记录现实依据/MBTI测试版本/MBTI分析/风险/30-90-180天行动；任务可重试且不重复扣权益；文件写 MinIO，用户只能访问自己的预签名 URL；完成后通过 Outbox 触发 n8n 通知
  - 提交：`feat(report): add deep consultation reports`

- [ ] **Task 18 微信小程序完整用户流程**（TaskList #22）
  - 文件：`apps/miniapp/src/pages/{index,login,chat,profile,mbti,report,orders,settings}/`、`services/api.ts`、`stores/*.ts` 及 Vitest
  - 登录/画像补全/MBTI测试引导+截图上传/基础咨询/深度分析入口/流式或轮询状态/风险场景专用安全页/历史会话/报告/订单/数据设置；网络重试复用相同 `request_id`
  - 提交：`feat(miniapp): implement consumer experience`

- [ ] **Task 19 隐私同意导出与删除**（TaskList #17）
  - 文件：migration `000006_privacy_consent.*.sql`、Go `internal/domain/privacy/*.go`、小程序 `pages/settings/privacy.tsx` 及测试
  - 用户单独同意存 MBTI/聊天截图/长期记忆；支持撤回长期记忆授权；支持导出画像/MBTI/会话/报告；注销时删或匿名化 MySQL 数据，并通过 Outbox 删 Qdrant point 和 MinIO 文件；审计日志不记正文和凭证
  - 提交：`feat(privacy): add consent and data lifecycle`

- [ ] **Task 20 可观测性限流与安全加固**（TaskList #20）
  - 文件：`infra/observability/*`、Go/Python metrics 和 tracing middleware、`infra/nginx/nginx.conf` 及测试
  - 指标覆盖 QPS/p95-p99/模型耗时/token/风险命中/Outbox延迟；日志同 `trace_id` 串联；用户/IP/接口三级限流；模型/OCR/n8n 超时/重试/熔断；上传文件扫描类型阻路径穿越
  - 提交：`feat(observability): add production guardrails`

- [ ] **Task 21 全链路 E2E 与 Agent 评估**（TaskList #21）
  - 文件：`tests/e2e/*.py`、`tests/evaluation/golden_cases.jsonl`、`run_eval.py`、`infra/scripts/e2e.sh`
  - 覆盖注册/MBTI/基础咨询/深度分析/OCR/RAG/报告/支付Mock/回访事件；验证无 MBTI 只得基础分析；切换 MBTI 后历史报告保留旧快照；危机文本不返回普通行动建议；黄金集至少情感50/职业50/混合20/安全50条
  - 验收：`make test && make e2e && make eval`
  - 提交：`test(e2e): cover complete consultation journey`

- [ ] **Task 22 云资源替换演练**（TaskList #19）
  - 文件：`docs/cloud-migration.md`、`infra/docker-compose.external.yml`、`infra/scripts/verify-external-services.sh`
  - 用 Compose override 停止本地 MySQL/Redis/Qdrant/MinIO/n8n；仅改环境变量连外部测试资源；执行 migration/Qdrant初始化/E2E；验证领域代码和 LangGraph 节点零修改；记录备份/回滚/DNS连接池/TLS/凭证轮换
  - 提交：`docs(infra): verify cloud service portability`

- [ ] **Task 23 发布验收**（TaskList #23）
  - 全部依赖和镜像固定版本，提交 `uv.lock/go.sum/pnpm-lock.yaml`；`docker compose config` 无错误；所有 migration 可升降级；所有健康检查通过；单元/集成/E2E/安全/评估测试通过；仓库无密钥/真实用户数据/上传文件；人工危机响应流程完成一次演练；发布文档含启动/备份/恢复/回滚/云资源替换
  - 提交：`chore(release): prepare zhipath v0.1.0`

## 四、最终完成标准（全部满足方可声明完成）

1. 从全新目录按文档可生成项目
2. 一条命令启动全部后端容器
3. 小程序可完成完整 C 端用户流程
4. 深度分析强提醒 MBTI，并保存当前结果、历史结果和分析快照
5. 情感、职业、混合分析都有结构化结果和安全控制
6. MySQL、Qdrant、MinIO、n8n 数据链路可验证
7. Outbox 重试不产生重复向量或重复通知
8. 用户可导出和删除自己的敏感数据
9. 断电或服务重启后，LangGraph 会话可恢复
10. 替换外部云资源只修改配置和 adapter，不修改领域逻辑
11. 所有自动化测试通过
12. 没有未完成占位符、静默降级和伪造成功

## 五、推进说明

- 顺序按 Task 1 → Task 23 推进，前序任务通过验收后再进入下一任务。
- 每个任务严格遵循 TDD：写失败测试 → 确认失败 → 最小实现 → 测试通过 → 提交。
- TaskList 已建对应 23 个任务，开发时逐个认领、完成后勾选本表 checkbox 并在 TaskList 标记 completed。
- 三份权威文档需放入项目 `docs/` 目录：`docs/langgraph_consult_agent_architecture.md`、`docs/consult_agent_project_bootstrap_guide.md`、`docs/zhipath_full_project_implementation_plan.md`。
