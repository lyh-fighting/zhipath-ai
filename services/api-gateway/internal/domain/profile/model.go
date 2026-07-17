package profile

import "time"

// Profile 用户画像（长期档案）。
type Profile struct {
	TenantID            string   `json:"tenant_id"`
	UserID              string   `json:"user_id"`
	AgeRange            string   `json:"age_range,omitempty"`
	City                string   `json:"city,omitempty"`
	Education           string   `json:"education,omitempty"`
	Occupation          string   `json:"occupation,omitempty"`
	Industry            string   `json:"industry,omitempty"`
	WorkYears           *float64 `json:"work_years,omitempty"`
	IncomeRange         string   `json:"income_range,omitempty"`
	RelationshipStatus  string   `json:"relationship_status,omitempty"`
	CurrentMBTIResultID *string  `json:"current_mbti_result_id,omitempty"`
	ProfileCompleteness int      `json:"profile_completeness"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// UpdateRequest 用户可更新的画像字段。
type UpdateRequest struct {
	AgeRange           *string  `json:"age_range,omitempty"`
	City               *string  `json:"city,omitempty"`
	Education          *string  `json:"education,omitempty"`
	Occupation         *string  `json:"occupation,omitempty"`
	Industry           *string  `json:"industry,omitempty"`
	WorkYears          *float64 `json:"work_years,omitempty"`
	IncomeRange        *string  `json:"income_range,omitempty"`
	RelationshipStatus *string  `json:"relationship_status,omitempty"`
}
