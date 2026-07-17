#!/usr/bin/env bash
# 验证外部服务连通性
set -u
echo "===== 验证外部服务连通性 ====="

echo -n "MySQL ($MYSQL_HOST): "
if command -v mysqladmin >/dev/null 2>&1; then
  mysqladmin ping -h "${MYSQL_HOST:-localhost}" -u "${MYSQL_USER:-consult}" -p"${MYSQL_PASSWORD:-consult}" 2>/dev/null && echo "" || echo "down"
else
  echo "跳过（未安装 mysql client）"
fi

echo -n "Redis ($REDIS_URL): "
if command -v redis-cli >/dev/null 2>&1; then
  redis-cli -u "${REDIS_URL:-redis://localhost:6379/0}" ping 2>/dev/null || echo "down"
else
  echo "跳过（未安装 redis client）"
fi

echo -n "Qdrant ($QDRANT_URL): "
curl -sf "${QDRANT_URL:-http://localhost:6333}/healthz" >/dev/null 2>&1 && echo "ok" || echo "down"

echo -n "MinIO ($OBJECT_STORAGE_ENDPOINT): "
curl -sf "http://${OBJECT_STORAGE_ENDPOINT:-localhost:9000}/minio/health/live" >/dev/null 2>&1 && echo "ok" || echo "down"
