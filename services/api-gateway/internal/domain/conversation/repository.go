package conversation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// Repository 会话与消息数据访问。
type Repository struct {
	db *sql.DB
}

// NewRepository 构造 Repository。
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建会话。
func (r *Repository) Create(ctx context.Context, tenantID, userID string, req *CreateRequest) (*Conversation, error) {
	conv := &Conversation{
		ConversationID: "c_" + uuid.NewString(),
		TenantID:        tenantID,
		UserID:          userID,
		Domain:          req.Domain,
		Title:           req.Title,
		Status:          "active",
		RiskLevel:       "none",
	}
	if conv.Domain == "" {
		conv.Domain = "auto"
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO conversations (tenant_id, conversation_id, user_id, domain, title, status, risk_level)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		conv.TenantID, conv.ConversationID, conv.UserID, conv.Domain, conv.Title, conv.Status, conv.RiskLevel,
	)
	if err != nil {
		return nil, fmt.Errorf("conversation create: %w", err)
	}
	return conv, nil
}

// Get 查询会话并校验归属。不存在返回 (nil, nil)。
func (r *Repository) Get(ctx context.Context, tenantID, conversationID string) (*Conversation, error) {
	conv := &Conversation{}
	err := r.db.QueryRowContext(ctx, `
		SELECT conversation_id, tenant_id, user_id, domain, title, status, risk_level,
		       last_message_at, created_at, updated_at
		FROM conversations
		WHERE tenant_id = ? AND conversation_id = ? AND deleted_at IS NULL`,
		tenantID, conversationID,
	).Scan(&conv.ConversationID, &conv.TenantID, &conv.UserID, &conv.Domain, &conv.Title,
		&conv.Status, &conv.RiskLevel, &conv.LastMessageAt, &conv.CreatedAt, &conv.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("conversation get: %w", err)
	}
	return conv, nil
}

// ListByUser 用户会话列表（游标分页，按 last_message_at 倒序）。
func (r *Repository) ListByUser(ctx context.Context, tenantID, userID string, limit int, cursor *string) ([]*Conversation, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	q := `SELECT conversation_id, tenant_id, user_id, domain, title, status, risk_level,
	             last_message_at, created_at, updated_at
	      FROM conversations
	      WHERE tenant_id = ? AND user_id = ? AND deleted_at IS NULL`
	args := []any{tenantID, userID}
	if cursor != nil && *cursor != "" {
		q += ` AND last_message_at < ?`
		args = append(args, *cursor)
	}
	q += ` ORDER BY last_message_at DESC LIMIT ?`
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("conversation list: %w", err)
	}
	defer rows.Close()
	var convs []*Conversation
	for rows.Next() {
		conv := &Conversation{}
		if err := rows.Scan(&conv.ConversationID, &conv.TenantID, &conv.UserID, &conv.Domain,
			&conv.Title, &conv.Status, &conv.RiskLevel, &conv.LastMessageAt, &conv.CreatedAt, &conv.UpdatedAt); err != nil {
			return nil, err
		}
		convs = append(convs, conv)
	}
	return convs, rows.Err()
}

// SendMessage 事务内写用户消息 + 更新会话时间。失败回滚不残留半完成消息。
func (r *Repository) SendMessage(ctx context.Context, tenantID, userID, conversationID string, req *SendMessageRequest) (*Message, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	msg := &Message{
		MessageID:      "m_" + uuid.NewString(),
		ConversationID: conversationID,
		TenantID:       tenantID,
		UserID:         userID,
		Role:           "user",
		MessageType:    "text",
		Attachments:    req.Attachments,
	}
	// attachments 是 JSON 列：有附件时传 JSON 文本，无附件时传 NULL（string([]byte(nil)) 会变成 "" 触发 MySQL 3140）。
	var attsParam any
	if len(req.Attachments) > 0 {
		if b, err := json.Marshal(req.Attachments); err == nil {
			attsParam = string(b)
		}
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO messages (tenant_id, message_id, conversation_id, user_id, role, message_type, content_summary, attachments)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		msg.TenantID, msg.MessageID, msg.ConversationID, msg.UserID, msg.Role, msg.MessageType,
		truncate(req.Message, 500), attsParam,
	)
	if err != nil {
		return nil, fmt.Errorf("insert message: %w", err)
	}
	_, err = tx.ExecContext(ctx, `UPDATE conversations SET last_message_at = NOW(3) WHERE conversation_id = ?`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("update conversation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return msg, nil
}

// History 取会话最近 limit 条消息（按时间升序），用于拼装 AI 上下文历史。
// 先取最近 limit 条（倒序），再包一层升序返回，避免一次性加载全量历史。
func (r *Repository) History(ctx context.Context, tenantID, conversationID string, limit int) ([]*Message, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT message_id, conversation_id, tenant_id, user_id, role, message_type,
		       content_summary, token_count, created_at
		FROM (
			SELECT message_id, conversation_id, tenant_id, user_id, role, message_type,
			       content_summary, token_count, created_at
			FROM messages
			WHERE tenant_id = ? AND conversation_id = ? AND deleted_at IS NULL
			ORDER BY created_at DESC LIMIT ?
		) t
		ORDER BY created_at ASC`,
		tenantID, conversationID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("message history: %w", err)
	}
	defer rows.Close()
	var msgs []*Message
	for rows.Next() {
		m := &Message{}
		if err := rows.Scan(&m.MessageID, &m.ConversationID, &m.TenantID, &m.UserID,
			&m.Role, &m.MessageType, &m.ContentSummary, &m.TokenCount, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

// SaveAssistantMessage 落库 AI 助手消息，并更新会话最近消息时间。
// content 过长时截断到 content_summary 列上限(512)；如需完整文本，后续可落 content_encrypted。
func (r *Repository) SaveAssistantMessage(ctx context.Context, tenantID, userID, conversationID, content string) (*Message, error) {
	msg := &Message{
		MessageID:      "m_" + uuid.NewString(),
		ConversationID: conversationID,
		TenantID:       tenantID,
		UserID:         userID,
		Role:           "assistant",
		MessageType:    "text",
		ContentSummary: truncate(content, 512),
	}
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO messages (tenant_id, message_id, conversation_id, user_id, role, message_type, content_summary)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		msg.TenantID, msg.MessageID, msg.ConversationID, msg.UserID, msg.Role, msg.MessageType, msg.ContentSummary,
	); err != nil {
		return nil, fmt.Errorf("insert assistant message: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, `UPDATE conversations SET last_message_at = NOW(3) WHERE conversation_id = ?`, conversationID); err != nil {
		return nil, fmt.Errorf("update conversation last_message_at: %w", err)
	}
	return msg, nil
}

// ListMessages 消息列表（游标分页，按 created_at 升序）。
func (r *Repository) ListMessages(ctx context.Context, tenantID, conversationID string, limit int, before *string) ([]*Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	q := `SELECT message_id, conversation_id, tenant_id, user_id, role, message_type,
	             content_summary, token_count, created_at
	      FROM messages
	      WHERE tenant_id = ? AND conversation_id = ? AND deleted_at IS NULL`
	args := []any{tenantID, conversationID}
	if before != nil && *before != "" {
		q += ` AND created_at < ?`
		args = append(args, *before)
	}
	q += ` ORDER BY created_at ASC LIMIT ?`
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("message list: %w", err)
	}
	defer rows.Close()
	var msgs []*Message
	for rows.Next() {
		m := &Message{}
		if err := rows.Scan(&m.MessageID, &m.ConversationID, &m.TenantID, &m.UserID,
			&m.Role, &m.MessageType, &m.ContentSummary, &m.TokenCount, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func truncate(s string, n int) string {
	// 按 Unicode 字符（rune）截断，避免按字节截断时切坏多字节 UTF-8 字符导致 MySQL 拒绝写入。
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}
