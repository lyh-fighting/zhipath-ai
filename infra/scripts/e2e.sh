#!/usr/bin/env bash
set -e
echo "===== ZhiPath E2E ====="
cd "$(dirname "$0")/../.."

echo "[1/3] Go test..."
cd services/api-gateway && go test ./... && cd ../..

echo "[2/3] Python pytest..."
cd services/ai-agent-service && uv run pytest -q && cd ../..
cd services/ocr-service && uv run pytest -q && cd ../..
cd services/worker-service && uv run pytest -q && cd ../..

echo "[3/3] Agent 评估..."
cd services/ai-agent-service && uv run python ../../tests/evaluation/run_eval.py --eval && cd ../..

echo "✓ E2E 完成"
