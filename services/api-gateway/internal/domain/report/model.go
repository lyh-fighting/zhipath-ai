package report

import "time"

// Report 深度报告。
// 记录现实依据、MBTI 测试版本、MBTI 分析、风险、30/90/180 天行动。
type Report struct {
	ReportID       string            `json:"report_id"`
	TenantID       string            `json:"tenant_id"`
	UserID         string            `json:"user_id"`
	ConversationID string            `json:"conversation_id"`
	Status         string            `json:"status"` // pending|generating|completed|failed
	MBTIResultID   string            `json:"mbti_result_id"`
	MBTISnapshot   map[string]any    `json:"mbti_snapshot"` // 不可变快照
	Reality        string            `json:"reality"`
	MBTIAnalysis   string            `json:"mbti_analysis"`
	Risks          []map[string]any  `json:"risks"`
	ActionPlan30   []map[string]any  `json:"action_plan_30d"`
	ActionPlan90   []map[string]any  `json:"action_plan_90d"`
	ActionPlan180  []map[string]any  `json:"action_plan_180d"`
	DownloadURL    string            `json:"download_url,omitempty"`
	FileID         string            `json:"file_id,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}
