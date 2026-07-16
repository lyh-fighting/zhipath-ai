# ZhiPath（知途 AI）

> 知己，知路，走好下一步。

基于 LangGraph 的智能咨询 Agent 平台，支持情感咨询、职业规划咨询、MBTI 画像、OCR 识别、多端接入与长期记忆。

## 服务

| 服务 | 目录 | 技术栈 | 端口 |
| --- | --- | --- | --- |
| API Gateway | `services/api-gateway` | Go 1.23 + Hertz | 8080 |
| AI Agent Service | `services/ai-agent-service` | Python 3.12 + FastAPI + LangGraph | 8001 |
| OCR Service | `services/ocr-service` | Python 3.12 + FastAPI + PaddleOCR | 8002 |
| Worker Service | `services/worker-service` | Python 3.12 | 8003 |
| 微信小程序 | `apps/miniapp` | Taro + React + TypeScript | - |

## 架构

- Go Hertz 是公网入口和 MySQL 业务事实的唯一写入者
- Python LangGraph 服务只执行分析并返回结构化结果
- OCR 服务只识别文件
- Worker 通过 transactional outbox 幂等同步 Qdrant、触发 n8n 和生成报告
- 本地依赖全部使用 Docker Compose，云上通过存储与服务适配接口替换

## 本地启动

```bash
cp .env.example .env
make help          # 查看所有命令
make infra-up      # 启动 MySQL/Redis/Qdrant/MinIO/n8n
make migrate       # 执行数据库迁移
make init-qdrant   # 初始化 Qdrant collection
make test          # 运行全部测试
```

## 文档

- `docs/langgraph_consult_agent_architecture.md`：架构与数据规则
- `docs/consult_agent_project_bootstrap_guide.md`：项目创建与初始化
- `docs/zhipath_full_project_implementation_plan.md`：全项目实施计划（Task 1-23）

## 开发约束

1. 不使用 LangChain，不安装任何 `langchain-*` 包
2. 不把客户端传入的 `user_id` 当可信身份；身份必须来自登录态
3. AI/OCR/Worker 不得直接修改用户、订单、画像、消息等业务事实
4. MySQL 是事实源；Qdrant/Redis/MinIO 索引均可重建
5. MySQL 与异步系统间用 `outbox_events`，禁止请求内双写
6. 深度分析必须读取当前有效 MBTI；缺失时只生成基础分析
7. 敏感文本、OCR 结果、风险详情必须加密存储
8. 无真实凭证时用明确 Mock Adapter，不静默伪造成功
