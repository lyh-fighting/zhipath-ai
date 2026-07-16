-- ZhiPath 初始 schema：租户、用户、画像、会话消息、Agent 运行、Prompt、文件
-- 对应架构文档第 5 节

-- ===== 租户（与 000-default-tenant.sql 兼容，IF NOT EXISTS）=====
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

-- ===== 用户 =====
CREATE TABLE users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  user_id VARCHAR(64) NOT NULL UNIQUE,
  nickname VARCHAR(128) NULL,
  avatar_url VARCHAR(512) NULL,
  gender VARCHAR(16) NULL,
  birth_year INT NULL,
  city VARCHAR(64) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  registered_channel VARCHAR(64) NULL,
  last_login_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  KEY idx_tenant_user (tenant_id, user_id),
  KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ===== 登录身份 =====
CREATE TABLE user_auth_identities (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  user_id VARCHAR(64) NOT NULL,
  provider VARCHAR(32) NOT NULL,
  provider_user_id VARCHAR(128) NOT NULL,
  union_id VARCHAR(128) NULL,
  phone_hash VARCHAR(128) NULL,
  credential_encrypted TEXT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_provider_user (provider, provider_user_id),
  KEY idx_tenant_user (tenant_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ===== 用户画像 =====
CREATE TABLE user_profiles (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  user_id VARCHAR(64) NOT NULL,
  age_range VARCHAR(32) NULL,
  city VARCHAR(64) NULL,
  education VARCHAR(64) NULL,
  occupation VARCHAR(128) NULL,
  industry VARCHAR(128) NULL,
  work_years DECIMAL(4,1) NULL,
  income_range VARCHAR(64) NULL,
  skills JSON NULL,
  career_goal TEXT NULL,
  relationship_status VARCHAR(64) NULL,
  mbti_type VARCHAR(8) NULL,
  mbti_assertiveness VARCHAR(1) NULL,
  mbti_source VARCHAR(32) NULL,
  mbti_confidence DECIMAL(4,3) NULL,
  mbti_updated_at DATETIME(3) NULL,
  personality_tags JSON NULL,
  current_challenges JSON NULL,
  risk_flags JSON NULL,
  profile_completeness INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_user (tenant_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ===== 会话 =====
CREATE TABLE conversations (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  conversation_id VARCHAR(64) NOT NULL UNIQUE,
  user_id VARCHAR(64) NOT NULL,
  domain VARCHAR(32) NOT NULL DEFAULT 'auto',
  title VARCHAR(256) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  summary TEXT NULL,
  risk_level VARCHAR(32) NOT NULL DEFAULT 'none',
  last_message_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  KEY idx_tenant_user_lastmsg (tenant_id, user_id, last_message_at),
  KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ===== 消息 =====
CREATE TABLE messages (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  message_id VARCHAR(64) NOT NULL UNIQUE,
  conversation_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  role VARCHAR(32) NOT NULL,
  message_type VARCHAR(32) NOT NULL DEFAULT 'text',
  content_encrypted MEDIUMTEXT NULL,
  content_summary VARCHAR(512) NULL,
  attachments JSON NULL,
  metadata JSON NULL,
  token_count INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  KEY idx_conv_created (conversation_id, created_at),
  KEY idx_tenant_user (tenant_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ===== 会话摘要 =====
CREATE TABLE conversation_summaries (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  summary_id VARCHAR(64) NOT NULL UNIQUE,
  conversation_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  summary_type VARCHAR(32) NOT NULL,
  summary_text TEXT NOT NULL,
  covered_message_start_id VARCHAR(64) NULL,
  covered_message_end_id VARCHAR(64) NULL,
  vector_point_id VARCHAR(128) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  KEY idx_tenant_user (tenant_id, user_id),
  KEY idx_conv (conversation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ===== Agent 运行记录 =====
CREATE TABLE agent_runs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  run_id VARCHAR(64) NOT NULL UNIQUE,
  trace_id VARCHAR(128) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  conversation_id VARCHAR(64) NULL,
  graph_version VARCHAR(64) NOT NULL,
  intent VARCHAR(32) NULL,
  state_snapshot JSON NULL,
  status VARCHAR(32) NOT NULL,
  started_at DATETIME(3) NOT NULL,
  finished_at DATETIME(3) NULL,
  error_code VARCHAR(64) NULL,
  error_message TEXT NULL,
  mbti_result_id VARCHAR(64) NULL,
  mbti_snapshot JSON NULL,
  KEY idx_trace (trace_id),
  KEY idx_tenant_user (tenant_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ===== Agent 节点输出 =====
CREATE TABLE agent_node_outputs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  run_id VARCHAR(64) NOT NULL,
  node_name VARCHAR(128) NOT NULL,
  input_snapshot JSON NULL,
  output_snapshot JSON NULL,
  latency_ms INT NULL,
  model_used VARCHAR(128) NULL,
  prompt_version VARCHAR(64) NULL,
  input_tokens INT NOT NULL DEFAULT 0,
  output_tokens INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  KEY idx_run (run_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ===== Prompt 版本 =====
CREATE TABLE prompt_versions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  prompt_key VARCHAR(128) NOT NULL,
  version VARCHAR(64) NOT NULL,
  domain VARCHAR(32) NOT NULL,
  content MEDIUMTEXT NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'draft',
  created_by VARCHAR(64) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_prompt_version (tenant_id, prompt_key, version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ===== 文件 =====
CREATE TABLE files (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  file_id VARCHAR(64) NOT NULL UNIQUE,
  user_id VARCHAR(64) NOT NULL,
  conversation_id VARCHAR(64) NULL,
  file_type VARCHAR(32) NOT NULL,
  mime_type VARCHAR(128) NULL,
  storage_url VARCHAR(1024) NOT NULL,
  sha256 VARCHAR(128) NOT NULL,
  size_bytes BIGINT NOT NULL,
  upload_client VARCHAR(64) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'uploaded',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  KEY idx_tenant_user (tenant_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
