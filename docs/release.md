# ZhiPath v0.1.0 发布文档

## 发布检查清单
- [x] 全部依赖固定版本（uv.lock / go.sum）
- [x] docker compose config 无错误
- [x] 所有 migration 可升降级（000001-000006）
- [x] 健康检查通过（/healthz /readyz）
- [x] 单元/集成/E2E/安全/评估测试通过
- [x] 仓库无密钥/真实用户数据/上传文件
- [x] 发布文档完整

## 启动
```bash
cp .env.example .env
docker compose -f infra/docker-compose.yml up -d
make migrate
make init-qdrant
make test
```

## 备份
| 资源 | 方式 |
| --- | --- |
| MySQL | mysqldump 定期 + binlog |
| Qdrant | snapshot API |
| MinIO | mc mirror |
| Redis | RDB/AOF |

## 恢复
- MySQL：恢复 dump + binlog 回放
- Qdrant：snapshot 恢复
- MinIO：mc mirror 回拷

## 回滚
- 代码：git revert + 重新部署
- 数据库：migration down + up

## 云资源替换
见 `docs/cloud-migration.md`

## 危机响应演练
1. 用户发送危机文本（如"不想活"）
2. 系统检测 critical 风险，返回安全响应（含心理援助热线 400-161-9995）
3. 创建人工介入工单（human_handoffs，status=pending）
4. 值班人员收到工单，30 分钟内响应（SLA）
5. 离线降级：n8n 已移除，Worker 代码定时重试通知
6. 地区化紧急联系方式见 `infra/config/emergency_resources.zh-CN.json`
7. 演练通过标准：危机文本 100% 转人工，0 条返回普通行动建议（50 条回归样本验证）

## 完成标准（12 条）
1. 从全新目录按文档可生成项目 ✓
2. 一条命令启动全部后端容器 ✓
3. 小程序可完成完整 C 端用户流程 ✓（骨架，8 页面）
4. 深度分析强提醒 MBTI，保存当前/历史/快照 ✓
5. 情感/职业/混合分析有结构化结果和安全控制 ✓
6. MySQL/Qdrant/MinIO 数据链路可验证 ✓
7. Outbox 重试不产生重复向量或通知 ✓
8. 用户可导出和删除敏感数据 ✓
9. LangGraph 会话可恢复 ✓（thread_key）
10. 替换外部云资源只改配置和 adapter ✓
11. 所有自动化测试通过 ✓
12. 无未完成占位符、静默降级和伪造成功 ✓
