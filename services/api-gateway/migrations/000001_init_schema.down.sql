-- 回滚 000001：按依赖反序删除（tenants 保留，由 000-default-tenant.sql 维护）
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS prompt_versions;
DROP TABLE IF EXISTS agent_node_outputs;
DROP TABLE IF EXISTS agent_runs;
DROP TABLE IF EXISTS conversation_summaries;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversations;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS user_auth_identities;
DROP TABLE IF EXISTS users;
