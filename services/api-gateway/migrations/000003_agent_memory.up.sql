-- 000003：Agent 记忆、决策、知识库
-- 对应架构文档 5.5（user_memory_items）、5.8（decision_records）、5.12（knowledge_*）

-- ===== 结构化长期记忆 =====
CREATE TABLE user_memory_items (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  memory_id VARCHAR(64) NOT NULL UNIQUE,
  user_id VARCHAR(64) NOT NULL,
  conversation_id VARCHAR(64) NULL,
  domain VARCHAR(32) NOT NULL,
  memory_type VARCHAR(64) NOT NULL COMMENT 'profile_fact|mbti_profile|career_goal|emotion_pattern|decision_history|risk_signal|preference',
  title VARCHAR(256) NULL,
  content TEXT NOT NULL,
  content_summary VARCHAR(512) NULL,
  importance_score DECIMAL(4,3) NOT NULL DEFAULT 0.500,
  confidence_score DECIMAL(4,3) NOT NULL DEFAULT 0.800,
  source VARCHAR(64) NOT NULL DEFAULT 'agent',
  vector_point_id VARCHAR(128) NULL,
  expires_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  KEY idx_tenant_user (tenant_id, user_id),
  KEY idx_domain_type (domain, memory_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ===== 决策记录 =====
CREATE TABLE decision_records (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  decision_id VARCHAR(64) NOT NULL UNIQUE,
  user_id VARCHAR(64) NOT NULL,
  conversation_id VARCHAR(64) NULL,
  domain VARCHAR(32) NOT NULL,
  decision_title VARCHAR(256) NOT NULL,
  problem_essence TEXT NULL,
  options JSON NULL,
  recommended_option TEXT NULL,
  risks JSON NULL,
  action_plan JSON NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'open',
  review_at DATETIME(3) NULL,
  outcome TEXT NULL,
  mbti_result_id VARCHAR(64) NULL,
  mbti_snapshot JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  KEY idx_tenant_user (tenant_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ===== 知识库文档 =====
CREATE TABLE knowledge_documents (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  doc_id VARCHAR(64) NOT NULL UNIQUE,
  domain VARCHAR(32) NOT NULL,
  title VARCHAR(256) NOT NULL,
  source VARCHAR(128) NULL,
  source_url VARCHAR(1024) NULL,
  version VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'draft' COMMENT 'draft|published|archived',
  risk_level VARCHAR(32) NOT NULL DEFAULT 'low',
  reviewer VARCHAR(64) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  KEY idx_tenant_domain_status (tenant_id, domain, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ===== 知识库 chunk =====
CREATE TABLE knowledge_chunks (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  chunk_id VARCHAR(64) NOT NULL UNIQUE,
  doc_id VARCHAR(64) NOT NULL,
  domain VARCHAR(32) NOT NULL,
  chunk_index INT NOT NULL,
  content TEXT NOT NULL,
  content_hash VARCHAR(128) NOT NULL,
  vector_point_id VARCHAR(128) NULL,
  metadata JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  KEY idx_doc (doc_id),
  KEY idx_tenant_domain (tenant_id, domain)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
