-- ZhiPath 默认租户种子数据
-- 由 docker-entrypoint-initdb.d 在 MySQL 首次启动时执行
-- 使用 CREATE TABLE IF NOT EXISTS 与 INSERT IGNORE 保证幂等，与 Task 4 migration 兼容

CREATE TABLE IF NOT EXISTS tenants (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL UNIQUE,
  name VARCHAR(128) NOT NULL,
  type VARCHAR(32) NOT NULL DEFAULT 'consumer',
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  settings JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT IGNORE INTO tenants (tenant_id, name, type, status)
VALUES ('default_consumer', '知途 C 端应用', 'consumer', 'active');

-- 验证
SELECT tenant_id, name, type, status FROM tenants WHERE tenant_id = 'default_consumer';
