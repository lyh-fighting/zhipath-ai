package risk

import "time"

// Event 风险事件。
type Event struct {
	RiskID         string    `json:"risk_id"`
	TenantID       string    `json:"tenant_id"`
	UserID         string    `json:"user_id"`
	ConversationID string    `json:"conversation_id,omitempty"`
	MessageID      string    `json:"message_id,omitempty"`
	RiskType       string    `json:"risk_type"`  // self_harm|suicide|domestic_violence|minor|violence
	RiskLevel      string    `json:"risk_level"` // low|medium|high|critical
	Detector       string    `json:"detector"`
	Status         string    `json:"status"` // open|resolved
	CreatedAt      time.Time `json:"created_at"`
}

// Handoff 人工介入工单。
type Handoff struct {
	HandoffID      string    `json:"handoff_id"`
	TenantID       string    `json:"tenant_id"`
	UserID         string    `json:"user_id"`
	ConversationID string    `json:"conversation_id"`
	RiskID         string    `json:"risk_id,omitempty"`
	Reason         string    `json:"reason"`
	Status         string    `json:"status"` // pending|assigned|resolved
	AssignedTo     string    `json:"assigned_to,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
