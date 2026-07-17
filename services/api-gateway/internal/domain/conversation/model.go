package conversation

import "time"

// Conversation 会话。
type Conversation struct {
	ConversationID string     `json:"conversation_id"`
	TenantID       string     `json:"tenant_id"`
	UserID         string     `json:"user_id"`
	Domain         string     `json:"domain"`
	Title          string     `json:"title,omitempty"`
	Status         string     `json:"status"`
	RiskLevel      string     `json:"risk_level"`
	LastMessageAt  *time.Time `json:"last_message_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Message 消息。
type Message struct {
	MessageID      string       `json:"message_id"`
	ConversationID string       `json:"conversation_id"`
	TenantID       string       `json:"tenant_id"`
	UserID         string       `json:"user_id"`
	Role           string       `json:"role"` // user|assistant|system
	MessageType    string       `json:"message_type"`
	ContentSummary string       `json:"content_summary,omitempty"`
	Attachments    []Attachment `json:"attachments,omitempty"`
	TokenCount     int          `json:"token_count"`
	CreatedAt      time.Time    `json:"created_at"`
}

// Attachment 附件。
type Attachment struct {
	FileID   string `json:"file_id"`
	FileType string `json:"file_type"`
}

// CreateRequest 创建会话请求。
type CreateRequest struct {
	Domain string `json:"domain"`
	Title  string `json:"title"`
}

// SendMessageRequest 发送消息请求。
type SendMessageRequest struct {
	Message          string       `json:"message" vd:"required,max=3000"`
	ConsultationType string       `json:"consultation_type"`
	Attachments      []Attachment `json:"attachments"`
	RequestID        string       `json:"request_id"` // 幂等键
}
