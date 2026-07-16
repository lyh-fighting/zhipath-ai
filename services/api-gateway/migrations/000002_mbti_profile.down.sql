-- 回滚 000002
DROP TABLE IF EXISTS user_mbti_results;
ALTER TABLE user_profiles DROP COLUMN current_mbti_result_id;
