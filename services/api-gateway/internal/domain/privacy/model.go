package privacy

import "time"

// Consent 用户同意（单独同意存储 MBTI/聊天截图/长期记忆）。
type Consent struct {
	TenantID    string     `json:"tenant_id"`
	UserID      string     `json:"user_id"`
	ConsentType string     `json:"consent_type"` // store_mbti|store_screenshot|long_term_memory|marketing
	Granted     bool       `json:"granted"`
	GrantedAt   *time.Time `json:"granted_at,omitempty"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
}

// ExportRequest 数据导出请求。
type ExportRequest struct {
	IncludeProfile bool `json:"include_profile"`
	IncludeMBTI    bool `json:"include_mbti"`
	IncludeChats   bool `json:"include_chats"`
	IncludeReports bool `json:"include_reports"`
}

// ConsentType 常量。
const (
	ConsentStoreMBTI        = "store_mbti"
	ConsentStoreScreenshot  = "store_screenshot"
	ConsentLongTermMemory   = "long_term_memory"
	ConsentMarketing        = "marketing"
)
