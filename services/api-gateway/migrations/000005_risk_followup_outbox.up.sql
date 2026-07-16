-- 000005：风险、人工介入、反馈、回访、Outbox
-- 对应架构文档 5.13 + 实施计划 Task 4/15（outbox_events）

CREATE TABLE risk_events (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  risk_id VARCHAR(64) NOT NULL UNIQUE,
  user_id VARCHAR(64) NOT NULL,
  conversation_id VARCHAR(64) NULL,
  message_id VARCHAR(64) NULL,
  risk_type VARCHAR(64) NOT NULL COMMENT 'self_harm|suicide|domestic_violence|minor|violence',
  risk_level VARCHAR(32) NOT NULL COMMENT 'low|medium|high|critical',
  detector VARCHAR(64) NOT NULL,
  detail_encrypted TEXT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'open',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  KEY idx_tenant_user (tenant_id, user_id),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE human_handoffs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  handoff_id VARCHAR(64) NOT NULL UNIQUE,
  user_id VARCHAR(64) NOT NULL,
  conversation_id VARCHAR(64) NOT NULL,
  risk_id VARCHAR(64) NULL,
  reason VARCHAR(256) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT 'pending|assigned|resolved',
  assigned_to VARCHAR(64) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  KEY idx_tenant_user (tenant_id, user_id),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE feedbacks (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  feedback_id VARCHAR(64) NOT NULL UNIQUE,
  user_id VARCHAR(64) NOT NULL,
  conversation_id VARCHAR(64) NULL,
  message_id VARCHAR(64) NULL,
  rating INT NULL,
  reason VARCHAR(128) NULL,
  comment TEXT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  KEY idx_tenant_user (tenant_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE followup_tasks (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  task_id VARCHAR(64) NOT NULL UNIQUE,
  user_id VARCHAR(64) NOT NULL,
  conversation_id VARCHAR(64) NULL,
  task_type VARCHAR(64) NOT NULL COMMENT 'followup_3d|review_7d|report_done|conversion',
  scheduled_at DATETIME(3) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  n8n_workflow_id VARCHAR(128) NULL,
  payload JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  KEY idx_status_scheduled (status, scheduled_at),
  KEY idx_tenant_user (tenant_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ===== Transactional Outbox（实施计划 Task 4/15）=====
-- MySQL 与异步系统（Qdrant/MinIO/n8n）间的可靠事件桥
CREATE TABLE outbox_events (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  event_id VARCHAR(64) NOT NULL UNIQUE,
  event_type VARCHAR(64) NOT NULL COMMENT 'memory_upserted|summary_upserted|chunk_upserted|memory_deleted|file_deleted|report_done|n8n_trigger',
  aggregate_type VARCHAR(64) NOT NULL,
  aggregate_id VARCHAR(64) NOT NULL,
  payload JSON NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT 'pending|processed|dead',
  retry_count INT NOT NULL DEFAULT 0,
  max_retries INT NOT NULL DEFAULT 5,
  next_attempt_at DATETIME(3) NULL,
  last_error TEXT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  processed_at DATETIME(3) NULL,
  KEY idx_status_next (status, next_attempt_at),
  KEY idx_event_type (event_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
