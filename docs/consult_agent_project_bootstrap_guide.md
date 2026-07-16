# 知途 AI 项目创建与迁移执行手册

本文档用于指导开发者或 AI Coding Agent 在一个空目录中完成智能咨询 Agent 项目的初始化、目录创建、参考代码拉取、迁移到自有 Git 仓库、服务骨架生成和后续开发。

推荐品牌名：

```text
中文名：知途
英文名：ZhiPath
仓库名：zhipath-ai
项目代号：zhipath
```

目标项目采用 `monorepo + 多服务独立启动` 架构：

```text
一个仓库统一维护
多个服务独立启动
多个服务独立部署
```

核心服务：

```text
api-gateway：Go Hertz，统一入口、认证、用户、订单、会话、多端适配
ai-agent-service：Python FastAPI + LangGraph，负责 Agent 工作流
ocr-service：Python FastAPI + PaddleOCR，负责图片/文档 OCR
worker-service：异步任务、报告生成、回访触发
miniapp：微信小程序入口
```

完整实现时必须同时读取：

```text
docs/consult_agent_project_bootstrap_guide.md
docs/langgraph_consult_agent_architecture.md
docs/zhipath_full_project_implementation_plan.md
```

其中启动手册定义如何创建项目，架构文档定义业务与数据规则，实施计划定义逐项开发和验收顺序。

## 1. 前置准备

### 1.1 本地开发工具

开发机需要安装：

```text
Git
Go 1.23
Python 3.12
uv
Node.js 20+
pnpm
Docker
Docker Compose
```

可选工具：

```text
make
just
direnv
pre-commit
```

### 1.2 检查命令

```bash
git --version
go version
python3 --version
uv --version
node --version
pnpm --version
docker --version
docker compose version
```

### 1.3 需要准备的远程仓库

建议先在 GitHub 创建一个空仓库，例如：

```text
git@github.com:<your-name>/zhipath-ai.git
```

注意：

- 创建仓库时可以不要初始化 README。
- 不要勾选 `.gitignore` 和 license，后续由项目统一生成。
- 本文档里的 `<your-name>` 替换成你的 GitHub 用户名或组织名。

## 2. 从空目录创建项目

### 2.1 创建项目目录

```bash
mkdir zhipath-ai
cd zhipath-ai
```

### 2.2 初始化 Git

```bash
git init
git branch -M main
```

### 2.3 创建基础目录

```bash
mkdir -p apps/miniapp
mkdir -p apps/mobile-app
mkdir -p apps/admin-web

mkdir -p services/api-gateway
mkdir -p services/api-gateway/migrations
mkdir -p services/ai-agent-service
mkdir -p services/ocr-service
mkdir -p services/worker-service

mkdir -p packages/shared-types
mkdir -p packages/api-client

mkdir -p infra/nginx
mkdir -p infra/k8s
mkdir -p infra/scripts
mkdir -p infra/mysql/init
mkdir -p infra/qdrant
mkdir -p infra/n8n/workflows
mkdir -p docs
mkdir -p references
```

### 2.4 推荐目录结构

最终结构应为：

```text
zhipath-ai/
├── apps/
│   ├── miniapp/
│   ├── mobile-app/
│   └── admin-web/
├── services/
│   ├── api-gateway/
│   │   └── migrations/
│   ├── ai-agent-service/
│   ├── ocr-service/
│   └── worker-service/
├── packages/
│   ├── shared-types/
│   └── api-client/
├── infra/
│   ├── docker-compose.yml
│   ├── nginx/
│   ├── k8s/
│   ├── mysql/init/
│   ├── qdrant/
│   ├── n8n/workflows/
│   └── scripts/
├── docs/
├── references/
├── .env.example
├── .gitignore
├── README.md
└── Makefile
```

## 3. 基础配置文件

### 3.1 `.gitignore`

在项目根目录创建 `.gitignore`：

```gitignore
# OS
.DS_Store

# env
.env
.env.*
!.env.example

# logs
*.log
logs/

# Python
__pycache__/
*.py[cod]
.pytest_cache/
.mypy_cache/
.ruff_cache/
.venv/
dist/
build/

# Go
bin/
*.test
coverage.out

# Node
node_modules/
dist/
.turbo/
.next/

# IDE
.idea/
.vscode/

# data
data/
storage/
uploads/
tmp/

# secrets
*.pem
*.key
*.crt
```

### 3.2 `.env.example`

```bash
APP_ENV=local

MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_USER=consult
MYSQL_PASSWORD=consult
MYSQL_DATABASE=zhipath
MYSQL_DSN=consult:consult@tcp(localhost:3306)/zhipath?parseTime=true
DATABASE_URL=mysql+mysqlconnector://consult:consult@localhost:3306/zhipath

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_URL=redis://localhost:6379/0

QDRANT_URL=http://localhost:6333
QDRANT_API_KEY=
QDRANT_COLLECTION=zhipath_kb_v1

S3_ENDPOINT=http://localhost:9000
S3_REGION=us-east-1
S3_BUCKET=zhipath-ai
S3_ACCESS_KEY=zhipath
S3_SECRET_KEY=zhipath-local-secret
S3_FORCE_PATH_STYLE=true

N8N_BASE_URL=http://localhost:5678
N8N_WEBHOOK_TOKEN=local-webhook-token

API_GATEWAY_PORT=8080
AI_AGENT_SERVICE_URL=http://localhost:8001
OCR_SERVICE_URL=http://localhost:8002

PRIMARY_MODEL=deepseek:deepseek-chat
FALLBACK_MODEL_1=openai:gpt-5.5
EMBEDDING_MODEL=text-embedding-3-large

DEEPSEEK_API_KEY=
OPENAI_API_KEY=
ANTHROPIC_API_KEY=

INFRA_MODE=docker
```

切换外部云资源时只替换以上连接配置。例如：

```text
MYSQL_DSN -> 云 MySQL 地址
REDIS_URL -> 云 Redis 地址
QDRANT_URL/QDRANT_API_KEY -> Qdrant Cloud
S3_ENDPOINT/S3_* -> S3、OSS、COS 或 TOS
N8N_BASE_URL -> 外部 n8n
```

业务代码必须通过 `RelationalStore`、`CacheStore`、`VectorStore`、`ObjectStore`、`AutomationClient` 接口访问资源，禁止直接依赖 Docker 主机名或云厂商 SDK。

### 3.3 根目录 `README.md`

```md
# ZhiPath

知途 AI 是基于 LangGraph 的智能咨询 Agent 平台，支持情感咨询、职业规划咨询、OCR 图片识别、多端接入和长期记忆。

## 服务

- `services/api-gateway`: Go Hertz，统一 API 入口
- `services/ai-agent-service`: Python FastAPI + LangGraph
- `services/ocr-service`: Python FastAPI + PaddleOCR
- `services/worker-service`: 异步任务
- `apps/miniapp`: 微信小程序

## 本地启动

```bash
cp .env.example .env
docker compose -f infra/docker-compose.yml up -d
make dev
```
```

### 3.4 根目录 `Makefile`

```makefile
.PHONY: dev infra-up infra-down api ai ocr worker

infra-up:
	docker compose -f infra/docker-compose.yml up -d

infra-down:
	docker compose -f infra/docker-compose.yml down

api:
	cd services/api-gateway && go run ./cmd/server

ai:
	cd services/ai-agent-service && uv run uvicorn app.main:app --host 0.0.0.0 --port 8001 --reload

ocr:
	cd services/ocr-service && uv run uvicorn app.main:app --host 0.0.0.0 --port 8002 --reload

worker:
	cd services/worker-service && uv run python worker.py

dev:
	@echo "请分别在多个终端运行: make api / make ai / make ocr / make worker"
```

## 4. 拉取参考代码

本项目的 Agent 核心采用 LangGraph，不使用 LangChain 作为项目框架，也不拉取 LangChain 源码。第三方大型项目源码只作为参考放到 `references/`，业务服务通过依赖包或服务接口使用相关能力。

### 4.1 LangGraph 使用方式

AI 服务只安装 LangGraph 和必要的模型/向量库 SDK：

```bash
uv add langgraph
uv add openai httpx qdrant-client
```

建议：

- LangGraph 负责状态流转、节点编排和条件分支。
- 模型调用通过 OpenAI-compatible SDK 或 `httpx` 直接封装 DeepSeek、OpenAI、Claude 等 provider。
- Qdrant 通过 `qdrant-client` 直接访问。
- 文档切分、Prompt 管理、RAG 检索逻辑在 `ai-agent-service` 内自研轻量实现。
- 不引入 `langchain`、`langchain-openai`、`langchain-qdrant`、`langchain-text-splitters` 等依赖，避免框架边界混乱。

### 4.2 拉取 PaddleOCR 参考代码

```bash
cd references
git clone git@github.com:PaddlePaddle/PaddleOCR.git PaddleOCR
cd ..
```

建议：

- `references/PaddleOCR` 只作为 OCR 引擎参考代码。
- 业务代码放在 `services/ocr-service`。
- OCR 服务通过安装 PaddleOCR 依赖或封装命令调用实现识别。
- 不要把 PaddleOCR 的完整源码混入 `services/ocr-service`。

## 5. 迁移到自己的 Git 仓库

### 5.1 绑定远程仓库

```bash
git remote add origin git@github.com:<your-name>/zhipath-ai.git
```

### 5.2 检查远程仓库

```bash
git remote -v
```

应看到：

```text
origin  git@github.com:<your-name>/zhipath-ai.git (fetch)
origin  git@github.com:<your-name>/zhipath-ai.git (push)
```

### 5.3 首次提交

```bash
git add .
git commit -m "chore: initialize consult agent platform"
git push -u origin main
```

### 5.4 如果已经从别的目录复制了文档

在空目录中可以这样操作：

```bash
mkdir zhipath-ai
cd zhipath-ai
cp /path/to/consult_agent_project_bootstrap_guide.md ./docs-bootstrap.md
```

然后让 AI Coding Agent 读取 `docs-bootstrap.md` 并执行本文档步骤。

## 6. 初始化 Go API Gateway

### 6.1 创建 Go module

```bash
cd services/api-gateway
go mod init github.com/<your-name>/zhipath-ai/services/api-gateway
go get github.com/cloudwego/hertz
go get github.com/redis/go-redis/v9
go get github.com/go-sql-driver/mysql
go get github.com/google/uuid
```

### 6.2 推荐目录

```text
services/api-gateway/
├── cmd/server/main.go
├── internal/config/
├── internal/middleware/
├── internal/handler/
├── internal/service/
├── internal/repository/
├── internal/client/
│   ├── aiagent/
│   └── ocr/
├── internal/model/
├── go.mod
└── go.sum
```

### 6.3 API Gateway 职责

第一版至少实现：

- `GET /healthz`
- `POST /api/v1/consultations`
- `POST /api/v1/consultations/{conversation_id}/messages`
- `GET /api/v1/conversations/{conversation_id}/messages`
- `POST /api/v1/files`
- `POST /api/v1/files/{file_id}/ocr`

## 7. 初始化 Python AI Agent Service

### 7.1 创建 Python 项目

```bash
cd services/ai-agent-service
uv init --package
uv add fastapi uvicorn pydantic pydantic-settings python-dotenv
uv add langgraph
uv add openai httpx qdrant-client redis
uv add pytest pytest-asyncio httpx ruff mypy --dev
```

AI Service 不直接写 MySQL 业务表。用户画像、MBTI、会话和消息由 Go API Gateway 读取并组装后传入；AI Service 返回结构化结果和待发布事件，由 Go 服务在同一事务中写入 MySQL 与 `outbox_events`。

### 7.2 推荐目录

```text
services/ai-agent-service/
├── app/
│   ├── main.py
│   ├── config.py
│   ├── api/
│   ├── graph/
│   │   ├── state.py
│   │   ├── builder.py
│   │   ├── nodes/
│   │   └── edges.py
│   ├── agents/
│   │   ├── emotion_agent.py
│   │   ├── career_agent.py
│   │   └── decision_coach_agent.py
│   ├── prompts/
│   ├── rag/
│   ├── tools/
│   ├── schemas/
│   └── storage/
├── tests/
├── pyproject.toml
└── README.md
```

### 7.3 第一版 LangGraph 节点

第一版实现这些节点即可：

```text
attachment_check
ocr_extract，可先调用 OCR Service mock
clean_extracted_text
load_context
intent_router
risk_detector
profile_completeness_check
retrieve_knowledge
emotion_agent
career_agent
multi_agent_collaboration
quality_check
repair_answer
human_handoff
finalize_response
persist_memory
```

### 7.4 AI Service API

```http
GET /healthz
POST /internal/v1/agent/invoke
POST /internal/v1/agent/stream
```

`POST /internal/v1/agent/invoke` 请求：

```json
{
  "trace_id": "trace_xxx",
  "user_id": "u_001",
  "conversation_id": "c_001",
  "message": "我想换工作但又很焦虑",
  "attachments": [],
  "client_context": {
    "client_type": "wechat_miniapp"
  }
}
```

## 8. 初始化 Python OCR Service

### 8.1 创建 Python 项目

```bash
cd services/ocr-service
uv init --package
uv add fastapi uvicorn pydantic pydantic-settings python-dotenv
uv add pillow opencv-python
uv add pytest httpx ruff mypy --dev
```

PaddleOCR 依赖可在第二阶段接入：

```bash
uv add paddleocr
```

如果 PaddleOCR 安装依赖较重，先使用 mock OCR 接口完成主链路开发，等服务边界稳定后再接入真实识别。

### 8.2 推荐目录

```text
services/ocr-service/
├── app/
│   ├── main.py
│   ├── config.py
│   ├── api/
│   ├── engine/
│   │   ├── base.py
│   │   ├── mock_engine.py
│   │   └── paddle_engine.py
│   ├── preprocess/
│   ├── schemas/
│   └── storage/
├── tests/
├── pyproject.toml
└── README.md
```

### 8.3 OCR Service API

```http
GET /healthz
POST /internal/v1/ocr/extract
```

请求：

```json
{
  "trace_id": "trace_xxx",
  "file_id": "file_xxx",
  "file_url": "s3://bucket/path/image.png",
  "file_type": "image"
}
```

响应：

```json
{
  "file_id": "file_xxx",
  "ocr_status": "completed",
  "clean_text": "识别后的文本",
  "avg_confidence": 0.91,
  "blocks": [
    {
      "type": "paragraph",
      "text": "岗位要求：3年以上前端经验...",
      "confidence": 0.94
    }
  ]
}
```

## 9. 初始化 Worker Service

### 9.1 创建 Python 项目

```bash
cd services/worker-service
uv init --package
uv add redis pydantic pydantic-settings python-dotenv
uv add httpx mysql-connector-python qdrant-client boto3
uv add pytest ruff mypy --dev
```

### 9.2 Worker 职责

第一版可以处理：

- 深度报告异步生成。
- OCR 异步任务。
- n8n 回访触发。
- 用户长期记忆摘要生成。
- 知识库导入任务。

## 10. 初始化微信小程序

### 10.1 Taro 方案

```bash
cd apps
pnpm create taro@latest miniapp
```

创建时选择：

```text
框架：React
语言：TypeScript
包管理器：pnpm
目标：微信小程序
```

### 10.2 小程序首批页面

```text
pages/index/index：入口页
pages/chat/index：咨询聊天页
pages/report/index：报告页
pages/profile/index：用户档案页
pages/orders/index：订单页
```

### 10.3 小程序调用链路

```text
小程序
  -> Go Hertz API Gateway
  -> Python AI Agent Service
  -> LangGraph
  -> 模型 / RAG / OCR / DB
```

小程序不能直接调用 Python AI Service 或 OCR Service。

## 11. 本地基础设施

### 11.1 `infra/docker-compose.yml`

```yaml
services:
  mysql:
    image: mysql:8.4
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: zhipath
      MYSQL_USER: consult
      MYSQL_PASSWORD: consult
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./mysql/init:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-uconsult", "-pconsult"]
      interval: 5s
      timeout: 3s
      retries: 20

  redis:
    image: redis:7.4-alpine
    restart: unless-stopped
    command: ["redis-server", "--appendonly", "yes"]
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 20

  qdrant:
    image: qdrant/qdrant:v1.15.4
    restart: unless-stopped
    ports:
      - "6333:6333"
      - "6334:6334"
    volumes:
      - qdrant_data:/qdrant/storage
    healthcheck:
      test: ["CMD-SHELL", "bash -c '</dev/tcp/127.0.0.1/6333'"]
      interval: 5s
      timeout: 3s
      retries: 20

  minio:
    image: minio/minio:RELEASE.2025-04-22T22-12-26Z
    restart: unless-stopped
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: zhipath
      MINIO_ROOT_PASSWORD: zhipath-local-secret
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - minio_data:/data

  minio-init:
    image: minio/mc:RELEASE.2025-04-16T18-13-26Z
    depends_on:
      - minio
    entrypoint:
      - /bin/sh
      - -c
      - >
        until mc alias set local http://minio:9000 zhipath zhipath-local-secret; do sleep 2; done;
        mc mb --ignore-existing local/zhipath-ai;

  n8n:
    image: n8nio/n8n:1.102.4
    restart: unless-stopped
    environment:
      N8N_HOST: localhost
      N8N_PORT: 5678
      N8N_PROTOCOL: http
      N8N_ENCRYPTION_KEY: local-n8n-encryption-key
      WEBHOOK_URL: http://localhost:5678/
    ports:
      - "5678:5678"
    volumes:
      - n8n_data:/home/node/.n8n

  api-gateway:
    build:
      context: ../services/api-gateway
    env_file: ../.env
    environment:
      MYSQL_DSN: consult:consult@tcp(mysql:3306)/zhipath?parseTime=true
      REDIS_URL: redis://redis:6379/0
      AI_AGENT_SERVICE_URL: http://ai-agent-service:8001
      OCR_SERVICE_URL: http://ocr-service:8002
    ports:
      - "8080:8080"
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
      ai-agent-service:
        condition: service_started

  ai-agent-service:
    build:
      context: ../services/ai-agent-service
    env_file: ../.env
    environment:
      REDIS_URL: redis://redis:6379/0
      QDRANT_URL: http://qdrant:6333
      S3_ENDPOINT: http://minio:9000
      OCR_SERVICE_URL: http://ocr-service:8002
    ports:
      - "8001:8001"
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
      qdrant:
        condition: service_healthy
      minio-init:
        condition: service_completed_successfully

  ocr-service:
    build:
      context: ../services/ocr-service
    env_file: ../.env
    environment:
      S3_ENDPOINT: http://minio:9000
    ports:
      - "8002:8002"
    depends_on:
      minio-init:
        condition: service_completed_successfully

  worker-service:
    build:
      context: ../services/worker-service
    env_file: ../.env
    environment:
      DATABASE_URL: mysql+mysqlconnector://consult:consult@mysql:3306/zhipath
      REDIS_URL: redis://redis:6379/0
      QDRANT_URL: http://qdrant:6333
      S3_ENDPOINT: http://minio:9000
      N8N_BASE_URL: http://n8n:5678
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
      qdrant:
        condition: service_healthy

volumes:
  mysql_data:
  redis_data:
  qdrant_data:
  minio_data:
  n8n_data:
```

### 11.2 数据库迁移和初始化

MySQL 表结构必须使用版本化 migration：

```text
services/api-gateway/migrations/
├── 000001_init_schema.up.sql
├── 000001_init_schema.down.sql
├── 000002_mbti_profile.up.sql
├── 000002_mbti_profile.down.sql
├── 000003_agent_memory.up.sql
├── 000003_agent_memory.down.sql
├── 000004_order_payment.up.sql
├── 000004_order_payment.down.sql
├── 000005_risk_followup_outbox.up.sql
└── 000005_risk_followup_outbox.down.sql
```

使用 `golang-migrate` 执行：

```bash
docker run --rm --network host \
  -v "$PWD/services/api-gateway/migrations:/migrations" \
  migrate/migrate:v4.18.3 \
  -path=/migrations \
  -database "mysql://consult:consult@tcp(localhost:3306)/zhipath" up
```

Qdrant 初始化脚本：

```text
infra/qdrant/init_collections.py
```

该脚本必须幂等创建：

```text
zhipath_kb_v1
zhipath_user_memory_v1
zhipath_conversation_summary_v1
```

并创建架构文档中声明的 payload indexes。脚本从 `EMBEDDING_DIMENSION` 读取向量维度，不得硬编码。

n8n 工作流以 JSON 形式纳入版本控制：

```text
infra/n8n/workflows/followup_3_days.json
infra/n8n/workflows/review_7_days.json
infra/n8n/workflows/report_completed.json
```

### 11.3 启动基础设施

```bash
docker compose -f infra/docker-compose.yml up -d
```

### 11.4 查看状态

```bash
docker compose -f infra/docker-compose.yml ps
```

所有容器健康后执行 migration 和 Qdrant 初始化，再运行端到端测试。

## 12. 本地启动顺序

推荐按以下顺序启动：

```text
1. MySQL / Redis / Qdrant / MinIO / n8n
2. OCR Service
3. AI Agent Service
4. Worker Service
5. API Gateway
6. 微信小程序
```

全栈 Docker 启动：

```bash
cp .env.example .env
docker compose -f infra/docker-compose.yml up -d --build
docker compose -f infra/docker-compose.yml ps
```

小程序仍使用微信开发者工具运行：

```bash
cd apps/miniapp
pnpm install
pnpm dev:weapp
```

## 13. 首批开发任务

### 13.1 第 1 阶段：项目骨架

目标：所有服务可以启动并返回 `/healthz`。

任务：

```text
创建 monorepo 目录
创建 .env.example、.gitignore、README.md、Makefile
创建 docker-compose
初始化 api-gateway
初始化 ai-agent-service
初始化 ocr-service
初始化 worker-service
提交首个 commit
```

验收：

```bash
curl http://localhost:8080/healthz
curl http://localhost:8001/healthz
curl http://localhost:8002/healthz
```

### 13.2 第 2 阶段：主链路打通

目标：小程序或 HTTP 请求可以通过 Go Hertz API Gateway 调到 AI Service 并返回 mock 咨询结果。

任务：

```text
实现 POST /api/v1/consultations
实现 POST /api/v1/consultations/{conversation_id}/messages
Go Hertz API Gateway 调用 AI Service
AI Service 返回 mock Agent 结果
保存 conversation、message、用户画像和决策轨迹到 MySQL
```

### 13.3 第 3 阶段：MBTI 画像 MVP

目标：深度咨询前能够引导用户完成 MBTI 测试，并保存完整人格测试结果。

任务：

```text
在 user_profiles 中保存 mbti_type、mbti_assertiveness、mbti_source、mbti_confidence、mbti_updated_at
创建 user_mbti_results 表，保存用户完整 MBTI 测试结果
实现 GET /api/v1/users/{user_id}/mbti
实现 POST /api/v1/users/{user_id}/mbti
实现 POST /api/v1/users/{user_id}/mbti/ocr
在深度职业规划和深度情感咨询前，如果用户没有 MBTI，提示用户先完成测试或上传结果
测试链接使用：http://16personalities.com/ch/%E4%BA%BA%E6%A0%BC%E6%B5%8B%E8%AF%95
将用户确认后的 MBTI 写入 user_profiles、user_memory_items 和 Qdrant 用户记忆
```

### 13.4 第 4 阶段：LangGraph MVP

目标：实现情感咨询、职业规划、混合咨询的基础路由。

任务：

```text
定义 ConsultationState
增加 mbti_profile、mbti_missing、should_prompt_mbti 状态字段
实现 intent_router
实现 risk_detector
实现 mbti_completeness_check
实现 emotion_agent
实现 career_agent
实现 quality_check
实现 persist_memory
```

### 13.5 第 5 阶段：OCR MVP

目标：上传图片后可以识别文本，并把文本传给 Agent。

任务：

```text
实现 file upload
实现 OCR Service mock
实现 POST /internal/v1/ocr/extract
实现 ai-agent-service 调用 ocr-service
接入 PaddleOCR
```

### 13.6 第 6 阶段：知识库和 RAG

目标：Agent 能基于咨询知识库回答。

任务：

```text
设计知识库文档格式
实现文档切分
生成 embedding
写入 Qdrant
实现 retrieve_knowledge
在情感/职业 Agent 中使用 retrieved_docs
```

## 14. AI Coding Agent 执行提示词

如果你把本文档放到空目录，可以直接给 AI Coding Agent 下面的提示词：

```text
请读取以下三份文档，并以实施计划的任务顺序逐项执行：
- docs/consult_agent_project_bootstrap_guide.md
- docs/langgraph_consult_agent_architecture.md
- docs/zhipath_full_project_implementation_plan.md

目标：
1. 在当前空目录中创建 zhipath-ai 项目结构。
2. 采用 monorepo + 多服务独立启动架构。
3. 创建 api-gateway、ai-agent-service、ocr-service、worker-service、apps/miniapp、infra、docs 等目录。
4. 生成 .gitignore、.env.example、README.md、Makefile、infra/docker-compose.yml。
5. 初始化 Go Hertz API Gateway，提供 /healthz。
6. 初始化 Python AI Agent Service，提供 /healthz 和 /internal/v1/agent/invoke。
7. 初始化 Python OCR Service，提供 /healthz 和 /internal/v1/ocr/extract，先用 mock engine。
8. 初始化 Worker Service。
9. 增加 MBTI 画像能力：user_mbti_results 表、MBTI 提交接口、MBTI OCR 识别入口、深度咨询前置提醒。
10. 不要拉取、复制或安装 LangChain；PaddleOCR 如需参考源码，放入 references/。
11. 使用版本化 MySQL migration、Qdrant 初始化脚本和 transactional outbox。
12. 本地用 Docker Compose 启动 MySQL、Redis、Qdrant、MinIO、n8n 和全部后端服务。
13. 每个服务必须能独立启动。
14. 按实施计划运行单元、集成、E2E、安全和 Agent 评估。
15. 只有满足实施计划“最终完成标准”才能声明项目完成。

约束：
- 不要提交 .env、密钥、node_modules、.venv、data、uploads。
- Go Hertz 服务负责业务入口和多端适配。
- Python AI 服务负责 LangGraph 工作流。
- OCR 服务负责 PaddleOCR，不能嵌入 AI 服务。
- 小程序不能直接调用 AI 服务或 OCR 服务，只能调用 Go Hertz API Gateway。
```

## 15. 代码拉取与迁移注意事项

### 15.1 第三方源码使用边界

本项目不使用 LangChain 作为项目框架，也不需要拉取 LangChain 源码。业务项目中应该：

```text
LangGraph：通过依赖包使用
模型 provider：通过 SDK 或 httpx 封装
Qdrant：通过 qdrant-client 使用
PaddleOCR：可通过 references/ 阅读源码，也可通过依赖包接入
```

不建议：

```text
把 PaddleOCR 源码复制到 services/ocr-service
把 LangChain 源码复制到 services/ai-agent-service
在 ai-agent-service 中安装 langchain、langchain-openai、langchain-qdrant 等依赖
直接修改第三方源码作为业务逻辑
```

### 15.2 推荐迁移方式

如果 AI 需要参考第三方代码：

```bash
mkdir -p references
cd references
git clone git@github.com:PaddlePaddle/PaddleOCR.git PaddleOCR
```

然后在业务代码中：

```text
ai-agent-service 使用 langgraph、openai/httpx、qdrant-client
ocr-service 使用 paddleocr 依赖包
```

### 15.3 自有仓库提交策略

首次提交：

```bash
git add .
git commit -m "chore: initialize consult agent platform"
git push -u origin main
```

后续建议分支：

```bash
git checkout -b feature/api-gateway-bootstrap
git checkout -b feature/ai-agent-langgraph-mvp
git checkout -b feature/ocr-service-bootstrap
git checkout -b feature/miniapp-chat-entry
```

## 16. 完成后的验证清单

项目创建完成后，应满足：

```text
[ ] 根目录有 README.md、Makefile、.env.example、.gitignore
[ ] infra/docker-compose.yml 可以启动 MySQL、Redis、Qdrant、MinIO、n8n 和全部后端服务
[ ] MySQL migrations 可以升级、回滚、再次升级
[ ] Qdrant collection 和 payload indexes 可以幂等初始化
[ ] MinIO bucket 可以幂等初始化
[ ] n8n 工作流 JSON 已纳入版本控制并可导入
[ ] services/api-gateway 可以独立启动
[ ] services/ai-agent-service 可以独立启动
[ ] services/ocr-service 可以独立启动
[ ] services/worker-service 可以独立启动
[ ] api-gateway /healthz 正常
[ ] ai-agent-service /healthz 正常
[ ] ocr-service /healthz 正常
[ ] api-gateway 能调用 ai-agent-service
[ ] ai-agent-service 能调用 ocr-service mock
[ ] MBTI 当前结果、历史结果和分析快照可以追溯
[ ] 深度分析缺少 MBTI 时强提醒，跳过后只能生成基础分析
[ ] Outbox 重试不会重复写向量或重复触发通知
[ ] 用户数据导出、撤回授权和注销删除链路通过测试
[ ] 使用外部测试资源时只需修改配置和 adapter
[ ] 单元、集成、E2E、安全和 Agent 黄金集评估通过
[ ] 代码已经提交到你自己的远程仓库
```

## 17. 最终建议

第一版不要追求所有能力一次做完。最合理的顺序是：

```text
先搭骨架
再打通 API 主链路
再实现 LangGraph MVP
再接知识库
再接 OCR
最后接支付、报告、会员、运营回访
```

这样能保证项目从第一天起就是可运行、可验证、可迭代的，而不是堆很多目录但主链路跑不通。
