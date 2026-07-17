package migration_test

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMigrationFilesPaired 验证 5 个 migration 的 up/down 文件均存在且配对。
// 实际 up/down/up 回滚测试在 Docker MySQL 就绪后由 make migration-test 覆盖。
func TestMigrationFilesPaired(t *testing.T) {
	migrations := []string{
		"000001_init_schema",
		"000002_mbti_profile",
		"000003_agent_memory",
		"000004_order_payment",
		"000005_risk_followup_outbox",
		"000006_privacy_consent",
	}
	dir := filepath.Join("..", "migrations")
	for _, m := range migrations {
		for _, suffix := range []string{"up.sql", "down.sql"} {
			p := filepath.Join(dir, m+"."+suffix)
			if _, err := os.Stat(p); err != nil {
				t.Errorf("migration 文件缺失: %s", p)
			}
		}
	}
}
