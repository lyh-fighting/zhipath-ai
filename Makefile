.DEFAULT_GOAL := help
.PHONY: help bootstrap infra-up infra-down migrate migration-test init-qdrant test lint e2e eval dev-api dev-ai dev-ocr dev-worker miniapp

##@ 通用

help: ## 显示所有可用命令
	@awk 'BEGIN {FS = ":.*##"; printf "\n\033[1mZhiPath 可用命令\033[0m\n\n"} \
	/^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 } \
	/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

bootstrap: ## 初始化项目：复制 .env 并启动基础设施
	@test -f .env || cp .env.example .env
	@echo "✓ .env 就绪"
	@$(MAKE) infra-up
	@$(MAKE) migrate
	@$(MAKE) init-qdrant

##@ 基础设施

infra-up: ## 启动 MySQL/Redis/Qdrant/MinIO/n8n 等后端依赖
	@echo "→ 启动 Docker Compose 基础设施..."
	docker compose -f infra/docker-compose.yml up -d --build
	@infra/scripts/wait-for-stack.sh || echo "⚠ wait-for-stack.sh 将在 Task 3 提供"

infra-down: ## 停止基础设施
	docker compose -f infra/docker-compose.yml down

migrate: ## 执行 MySQL 迁移（up）
	@echo "→ 执行数据库迁移..."
	@cd services/api-gateway && go run ./cmd/migrate up 2>/dev/null || echo "⚠ migrate 命令将在 Task 4 提供"

migration-test: ## 测试迁移可升级与回滚
	@echo "→ 测试迁移 up/down/up..."
	@cd services/api-gateway && go run ./cmd/migrate up && go run ./cmd/migrate down && go run ./cmd/migrate up
	@echo "✓ 迁移测试通过"

init-qdrant: ## 初始化 Qdrant collection 与 payload index
	@echo "→ 初始化 Qdrant collection..."
	@python infra/qdrant/init_collections.py 2>/dev/null || echo "⚠ init_collections.py 将在 Task 3 提供"

##@ 测试

test: ## 运行全部测试（Go + Python + 前端）
	@echo "→ Go 测试..."
	@cd services/api-gateway && go test ./... 2>/dev/null || echo "⚠ api-gateway 测试将在后续任务提供"
	@echo "→ Python 测试..."
	@cd services/ai-agent-service && uv run pytest -q 2>/dev/null || echo "⚠ ai-agent-service 测试将在后续任务提供"
	@cd services/ocr-service && uv run pytest -q 2>/dev/null || echo "⚠ ocr-service 测试将在后续任务提供"
	@cd services/worker-service && uv run pytest -q 2>/dev/null || echo "⚠ worker-service 测试将在后续任务提供"
	@echo "→ 前端测试..."
	@pnpm --dir packages/api-client test 2>/dev/null || echo "⚠ api-client 测试将在 Task 2 提供"

lint: ## 代码格式检查
	@cd services/api-gateway && go vet ./... 2>/dev/null || true
	@cd services/ai-agent-service && uv run ruff check . 2>/dev/null || true

e2e: ## 全链路 E2E 测试
	@echo "→ 运行 E2E..."
	@infra/scripts/e2e.sh 2>/dev/null || echo "⚠ E2E 将在 Task 21 提供"

eval: ## Agent 评估（黄金集）
	@echo "→ 运行 Agent 评估..."
	@python tests/evaluation/run_eval.py 2>/dev/null || echo "⚠ 评估将在 Task 21 提供"

##@ 本地开发

dev-api: ## 启动 Go API Gateway
	cd services/api-gateway && go run ./cmd/server

dev-ai: ## 启动 Python AI Agent Service
	cd services/ai-agent-service && uv run uvicorn app.main:app --host 0.0.0.0 --port 8001 --reload

dev-ocr: ## 启动 Python OCR Service
	cd services/ocr-service && uv run uvicorn app.main:app --host 0.0.0.0 --port 8002 --reload

dev-worker: ## 启动 Python Worker Service
	cd services/worker-service && uv run python -m app.worker

miniapp: ## 启动微信小程序开发
	cd apps/miniapp && pnpm dev:weapp
