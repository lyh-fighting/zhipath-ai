package profile

import (
	"context"
	"database/sql"
	"fmt"
)

// Repository 用户画像数据访问。
type Repository struct {
	db *sql.DB
}

// NewRepository 构造 Repository。
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Get 查询用户画像。不存在返回 (nil, nil)。
func (r *Repository) Get(ctx context.Context, tenantID, userID string) (*Profile, error) {
	p := &Profile{}
	var currentMBTI sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT tenant_id, user_id, age_range, city, education, occupation, industry,
		       work_years, income_range, relationship_status, current_mbti_result_id,
		       profile_completeness, created_at, updated_at
		FROM user_profiles WHERE tenant_id = ? AND user_id = ?`,
		tenantID, userID,
	).Scan(&p.TenantID, &p.UserID, &p.AgeRange, &p.City, &p.Education, &p.Occupation,
		&p.Industry, &p.WorkYears, &p.IncomeRange, &p.RelationshipStatus, &currentMBTI,
		&p.ProfileCompleteness, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("profile get: %w", err)
	}
	if currentMBTI.Valid {
		id := currentMBTI.String
		p.CurrentMBTIResultID = &id
	}
	return p, nil
}

// Upsert 创建或更新用户画像。
func (r *Repository) Upsert(ctx context.Context, p *Profile) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_profiles (tenant_id, user_id, age_range, city, education, occupation,
		    industry, work_years, income_range, relationship_status, profile_completeness)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		    age_range=VALUES(age_range), city=VALUES(city), education=VALUES(education),
		    occupation=VALUES(occupation), industry=VALUES(industry), work_years=VALUES(work_years),
		    income_range=VALUES(income_range), relationship_status=VALUES(relationship_status),
		    profile_completeness=VALUES(profile_completeness)`,
		p.TenantID, p.UserID, p.AgeRange, p.City, p.Education, p.Occupation, p.Industry,
		p.WorkYears, p.IncomeRange, p.RelationshipStatus, p.ProfileCompleteness,
	)
	if err != nil {
		return fmt.Errorf("profile upsert: %w", err)
	}
	return nil
}

// SetCurrentMBTI 设置当前生效的 MBTI 结果 ID，并刷新 mbti_updated_at。
func (r *Repository) SetCurrentMBTI(ctx context.Context, tenantID, userID, mbtiResultID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_profiles SET current_mbti_result_id = ?, mbti_updated_at = NOW(3)
		WHERE tenant_id = ? AND user_id = ?`,
		mbtiResultID, tenantID, userID,
	)
	if err != nil {
		return fmt.Errorf("profile set current mbti: %w", err)
	}
	return nil
}
