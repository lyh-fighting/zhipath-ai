package mbti

import "time"

// Dimension 单个维度，含极性与百分比（极性+百分比，不能仅存无方向数字）。
type Dimension struct {
	Pole       string `json:"pole"`       // E/I, S/N, T/F, J/P
	Percentage int    `json:"percentage"` // 0-100
}

// Dimensions 四维度。
type Dimensions struct {
	Energy  Dimension `json:"energy"`  // 外向 E / 内向 I
	Mind    Dimension `json:"mind"`     // 实感 S / 直觉 N
	Nature  Dimension `json:"nature"`   // 理性 T / 情感 F
	Tactics Dimension `json:"tactics"`  // 判断 J / 展望 P
}

// Result MBTI 测试结果。
type Result struct {
	MBTIResultID     string     `json:"mbti_result_id"`
	UserID           string     `json:"user_id"`
	ConversationID   string     `json:"conversation_id,omitempty"`
	Source           string     `json:"source"` // manual|ocr|imported|agent_extracted
	TestURL          string     `json:"test_url,omitempty"`
	ResultType       string     `json:"result_type"`       // 16 型，如 INFP
	Assertiveness    string     `json:"assertiveness,omitempty"` // A 坚决 / T 谨慎
	Dimensions       Dimensions `json:"dimensions"`
	EnergyScore      *int       `json:"energy_score,omitempty"`
	MindScore        *int       `json:"mind_score,omitempty"`
	NatureScore      *int       `json:"nature_score,omitempty"`
	TacticsScore     *int       `json:"tactics_score,omitempty"`
	IdentityScore    *int       `json:"identity_score,omitempty"`
	FileID           string     `json:"file_id,omitempty"`
	OCRID            string     `json:"ocr_id,omitempty"`
	ConfidenceScore  float64    `json:"confidence_score"`
	ConfirmedByUser  bool       `json:"confirmed_by_user"`
	IsCurrent        bool       `json:"is_current"` // 是否当前生效
	TestedAt         *time.Time `json:"tested_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// SubmitRequest 手动提交 MBTI。
type SubmitRequest struct {
	ResultType    string     `json:"result_type"`
	Assertiveness string     `json:"assertiveness"`
	Dimensions    Dimensions `json:"dimensions"`
	Source        string     `json:"source"`
	TestURL       string     `json:"test_url,omitempty"`
	TestedAt      *time.Time `json:"tested_at,omitempty"`
}
