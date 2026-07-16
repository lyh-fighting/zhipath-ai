-- 000002：MBTI 画像
-- user_profiles 增加指向当前生效结果的指针；新建完整 MBTI 测试结果表
-- 对应架构文档 5.5 + 实施计划 Task 4/7

ALTER TABLE user_profiles
  ADD COLUMN current_mbti_result_id VARCHAR(64) NULL COMMENT '当前生效的 MBTI 结果 ID（用户确认）';

CREATE TABLE user_mbti_results (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_consumer',
  mbti_result_id VARCHAR(64) NOT NULL UNIQUE,
  user_id VARCHAR(64) NOT NULL,
  conversation_id VARCHAR(64) NULL,
  source VARCHAR(32) NOT NULL COMMENT 'manual|ocr|imported|agent_extracted',
  test_url VARCHAR(1024) NULL,
  result_type VARCHAR(8) NOT NULL COMMENT '16 型，如 INFP',
  assertiveness VARCHAR(1) NULL COMMENT 'A 坚决 / T 谨慎',
  energy_score INT NULL COMMENT 'E/I 维度百分比',
  mind_score INT NULL COMMENT 'S/N 维度百分比',
  nature_score INT NULL COMMENT 'T/F 维度百分比',
  tactics_score INT NULL COMMENT 'J/P 维度百分比',
  identity_score INT NULL COMMENT 'A/T 维度百分比',
  raw_text MEDIUMTEXT NULL,
  raw_payload JSON NULL,
  file_id VARCHAR(64) NULL,
  ocr_id VARCHAR(64) NULL,
  confidence_score DECIMAL(4,3) NOT NULL DEFAULT 0.800,
  confirmed_by_user BOOLEAN NOT NULL DEFAULT FALSE,
  tested_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  KEY idx_tenant_user (tenant_id, user_id),
  KEY idx_user_tested (user_id, tested_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
