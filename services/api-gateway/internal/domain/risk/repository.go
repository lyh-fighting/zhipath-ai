package risk

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// Repository 风险事件与人工介入数据访问。
type Repository struct {
	db *sql.DB
}

// NewRepository 构造 Repository。
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建风险事件。
func (r *Repository) Create(ctx context.Context, tenantID string, e *Event) error {
	e.RiskID = "risk_" + uuid.NewString()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO risk_events (tenant_id, risk_id, user_id, conversation_id, message_id, risk_type, risk_level, detector, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'open')`,
		tenantID, e.RiskID, e.UserID, nullable(e.ConversationID), nullable(e.MessageID),
		e.RiskType, e.RiskLevel, e.Detector,
	)
	if err != nil {
		return fmt.Errorf("risk create: %w", err)
	}
	return nil
}

// CreateHandoff 创建人工介入工单。
func (r *Repository) CreateHandoff(ctx context.Context, tenantID string, h *Handoff) error {
	h.HandoffID = "ho_" + uuid.NewString()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO human_handoffs (tenant_id, handoff_id, user_id, conversation_id, risk_id, reason, status)
		VALUES (?, ?, ?, ?, ?, ?, 'pending')`,
		tenantID, h.HandoffID, h.UserID, h.ConversationID, nullable(h.RiskID), h.Reason,
	)
	if err != nil {
		return fmt.Errorf("handoff create: %w", err)
	}
	return nil
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
