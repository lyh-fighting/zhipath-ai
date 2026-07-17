package report

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// Repository 报告数据访问。
type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建报告任务。可重试，不重复扣权益。
func (r *Repository) Create(
	ctx context.Context, tenantID, userID, conversationID, mbtiResultID string,
	mbtiSnapshot map[string]any,
) (*Report, error) {
	rep := &Report{
		ReportID:       "rep_" + uuid.NewString(),
		TenantID:       tenantID,
		UserID:         userID,
		ConversationID: conversationID,
		Status:         "pending",
		MBTIResultID:   mbtiResultID,
		MBTISnapshot:   mbtiSnapshot,
	}
	// TODO: INSERT INTO reports
	return rep, nil
}

// MarkCompleted 标记完成，写 MinIO 文件 URL，发 Outbox 通知。
func (r *Repository) MarkCompleted(ctx context.Context, tenantID, reportID, fileID, downloadURL string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE reports SET status='completed', file_id=?, download_url=? WHERE report_id=? AND status IN ('pending','generating','failed')`,
		fileID, downloadURL, reportID,
	)
	if err != nil {
		return fmt.Errorf("report complete: %w", err)
	}
	// TODO: INSERT INTO outbox_events (event_type='report_done', aggregate_id=reportID)
	return nil
}

// Get 查询报告。用户只能访问自己的预签名 URL。
func (r *Repository) Get(ctx context.Context, tenantID, reportID string) (*Report, error) {
	// TODO: SELECT
	return nil, nil
}

// CanRetry 报告任务可重试且不重复扣权益。
func CanRetry(status string) bool {
	return status == "pending" || status == "generating" || status == "failed"
}
