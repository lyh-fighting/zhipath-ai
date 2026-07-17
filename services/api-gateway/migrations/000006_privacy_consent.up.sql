-- 000006：隐私同意与审计日志

-- 用户同意（单独同意存储 MBTI/聊天截图/长期记忆）
CREATE TABLE user_consents (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  user_id VARCHAR(64) NOT NULL,
  consent_type VARCHAR(64) NOT NULL COMMENT 'store_mbti|store_screenshot|long_term_memory|marketing',
  granted BOOLEAN NOT NULL DEFAULT FALSE,
  granted_at DATETIME(3) NULL,
  revoked_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_consent (tenant_id, user_id, consent_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 审计日志（不记正文和凭证，只记动作元数据）
CREATE TABLE audit_logs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  user_id VARCHAR(64) NOT NULL,
  action VARCHAR(64) NOT NULL COMMENT 'export|delete|consent_grant|consent_revoke|login',
  resource_type VARCHAR(64) NULL,
  resource_id VARCHAR(64) NULL,
  ip_hash VARCHAR(128) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  KEY idx_tenant_user (tenant_id, user_id),
  KEY idx_action (action)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
