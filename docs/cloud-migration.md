# 云资源替换指南

## 目标
仅修改配置和环境变量连接外部云资源，不修改领域代码和 LangGraph 节点。

## 替换步骤

### 1. 停止本地服务
```bash
docker compose -f infra/docker-compose.yml down
```

### 2. 配置外部资源环境变量
编辑 `.env`，填入外部连接信息：
- `DATABASE_URL`（外部 MySQL）
- `REDIS_URL`（外部 Redis）
- `QDRANT_URL`（外部 Qdrant）
- `OBJECT_STORAGE_ENDPOINT`（外部对象存储 OSS/COS/S3）

### 3. 启动应用服务（连外部资源）
```bash
docker compose -f infra/docker-compose.yml -f infra/docker-compose.external.yml --profile apps up -d
```

### 4. 执行 migration + Qdrant 初始化
```bash
make migrate
make init-qdrant
```

### 5. 运行 E2E 验证
```bash
make e2e
```

## 验证要点
- 领域代码（Go `internal/domain` + Python `app/agents`）零修改
- LangGraph 节点零修改
- 所有接口和 Agent 行为一致

## 备份与回滚
| 资源 | 备份方式 | 回滚方式 |
| --- | --- | --- |
| MySQL | mysqldump 定期 + binlog | 恢复 dump + binlog 回放 |
| Qdrant | snapshot API | snapshot 恢复 |
| MinIO | mc mirror | mc mirror 回拷 |
| Redis | RDB/AOF | 恢复 RDB |

## 连接池 / TLS / 凭证轮换
- MySQL：`max_open_conns=20`，生产启用 TLS
- Redis：TLS + 连接复用
- Qdrant：API key
- MinIO：access key/secret 定期轮换，应用通过适配接口无感切换
