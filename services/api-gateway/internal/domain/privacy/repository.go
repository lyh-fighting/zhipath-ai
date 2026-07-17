package privacy

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Repository 隐私数据访问。
type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Grant 用户授予同意（存储 MBTI/截图/长期记忆）。
func (r *Repository) Grant(ctx context.Context, tenantID, userID, consentType string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_consents (tenant_id, user_id, consent_type, granted, granted_at)
		VALUES (?, ?, ?, TRUE, ?)
		ON DUPLICATE KEY UPDATE granted=TRUE, granted_at=?, revoked_at=NULL`,
		tenantID, userID, consentType, now, now,
	)
	if err != nil {
		return fmt.Errorf("consent grant: %w", err)
	}
	return r.audit(ctx, tenantID, userID, "consent_grant", consentType)
}

// Revoke 撤回长期记忆授权等。
func (r *Repository) Revoke(ctx context.Context, tenantID, userID, consentType string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_consents SET granted=FALSE, revoked_at=NOW(3)
		WHERE tenant_id=? AND user_id=? AND consent_type=?`,
		tenantID, userID, consentType,
	)
	if err != nil {
		return fmt.Errorf("consent revoke: %w", err)
	}
	return r.audit(ctx, tenantID, userID, "consent_revoke", consentType)
}

// HasConsent 检查是否已同意。
func (r *Repository) HasConsent(ctx context.Context, tenantID, userID, consentType string) (bool, error) {
	var granted bool
	err := r.db.QueryRowContext(ctx,
		`SELECT granted FROM user_consents WHERE tenant_id=? AND user_id=? AND consent_type=?`,
		tenantID, userID, consentType,
	).Scan(&granted)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return granted, err
}

// Export 导出用户数据（画像/MBTI/会话/报告）。
func (r *Repository) Export(ctx context.Context, tenantID, userID string, req *ExportRequest) (map[string]any, error) {
	result := map[string]any{}
	// TODO: 按 req 查各表数据
	if err := r.audit(ctx, tenantID, userID, "export", ""); err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteAccount 注销：删或匿名化 MySQL + Outbox 删 Qdrant point + MinIO 文件。
func (r *Repository) DeleteAccount(ctx context.Context, tenantID, userID string) error {
	// TODO: 匿名化 users/user_profiles/messages/user_mbti_results
	// TODO: INSERT INTO outbox_events (memory_deleted/file_deleted) 触发 Qdrant/MinIO 清理
	return r.audit(ctx, tenantID, userID, "delete", "")
}

// audit 审计日志（不记正文和凭证，只记动作元数据）。
func (r *Repository) audit(ctx context.Context, tenantID, userID, action, resourceType string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO audit_logs (tenant_id, user_id, action, resource_type) VALUES (?, ?, ?, ?)`,
		tenantID, userID, action, nullable(resourceType),
	)
	return err
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
