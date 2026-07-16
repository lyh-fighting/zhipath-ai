#!/usr/bin/env bash
# 等待 zhipath 基础设施全部就绪
set -u

# host:port 列表（与 docker-compose 基础设施服务对应）
SERVICES=(
  "mysql:3306"
  "redis:6379"
  "qdrant:6333"
  "minio:9000"
  "n8n:5678"
)

MAX_WAIT=120
for svc in "${SERVICES[@]}"; do
  host="${svc%%:*}"
  port="${svc##*:}"
  echo -n "等待 $host:$port ..."
  waited=0
  while [ $waited -lt $MAX_WAIT ]; do
    if bash -c "</dev/tcp/$host/$port" 2>/dev/null; then
      echo " ✓"
      break
    fi
    sleep 2
    waited=$((waited + 2))
  done
  if [ $waited -ge $MAX_WAIT ]; then
    echo " ✗ 超时"
    exit 1
  fi
done

echo ""
echo "✓ 全部基础设施依赖就绪"
echo "  MySQL    : mysql:3306"
echo "  Redis    : redis:6379"
echo "  Qdrant   : qdrant:6333"
echo "  MinIO    : minio:9000 (console :9001)"
echo "  n8n      : n8n:5678"
