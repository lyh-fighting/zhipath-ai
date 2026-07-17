package mbti

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// Repository MBTI 数据访问。
type Repository struct {
	db *sql.DB
}

// NewRepository 构造 Repository。
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建 MBTI 结果。截图识别（source=ocr）默认 confirmed=false，需用户确认才成当前。
func (r *Repository) Create(ctx context.Context, tenantID string, res *Result) error {
	res.MBTIResultID = "mbti_" + uuid.NewString()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_mbti_results
			(tenant_id, mbti_result_id, user_id, conversation_id, source, test_url,
			 result_type, assertiveness, energy_score, mind_score, nature_score,
			 tactics_score, identity_score, file_id, ocr_id, confidence_score,
			 confirmed_by_user, tested_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		tenantID, res.MBTIResultID, res.UserID, nullable(res.ConversationID), res.Source, nullable(res.TestURL),
		res.ResultType, nullable(res.Assertiveness), nullInt(res.EnergyScore), nullInt(res.MindScore),
		nullInt(res.NatureScore), nullInt(res.TacticsScore), nullInt(res.IdentityScore),
		nullable(res.FileID), nullable(res.OCRID), res.ConfidenceScore, res.ConfirmedByUser, res.TestedAt,
	)
	if err != nil {
		return fmt.Errorf("mbti create: %w", err)
	}
	return nil
}

// Get 查询单个结果。不存在返回 (nil, nil)。
func (r *Repository) Get(ctx context.Context, tenantID, mbtiResultID string) (*Result, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT mbti_result_id, user_id, conversation_id, source, test_url, result_type,
		       assertiveness, energy_score, mind_score, nature_score, tactics_score, identity_score,
		       file_id, ocr_id, confidence_score, confirmed_by_user, tested_at, created_at
		FROM user_mbti_results
		WHERE tenant_id = ? AND mbti_result_id = ? AND deleted_at IS NULL`,
		tenantID, mbtiResultID,
	)
	res, err := scanResult(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return res, err
}

// ListByUser 用户历史结果（倒序）。
func (r *Repository) ListByUser(ctx context.Context, tenantID, userID string, limit int) ([]*Result, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT mbti_result_id, user_id, conversation_id, source, test_url, result_type,
		       assertiveness, energy_score, mind_score, nature_score, tactics_score, identity_score,
		       file_id, ocr_id, confidence_score, confirmed_by_user, tested_at, created_at
		FROM user_mbti_results
		WHERE tenant_id = ? AND user_id = ? AND deleted_at IS NULL
		ORDER BY tested_at DESC, created_at DESC LIMIT ?`,
		tenantID, userID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("mbti list: %w", err)
	}
	defer rows.Close()
	var results []*Result
	for rows.Next() {
		res, err := scanResult(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	return results, rows.Err()
}

// Confirm 用户确认 MBTI 结果。
func (r *Repository) Confirm(ctx context.Context, tenantID, mbtiResultID string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE user_mbti_results SET confirmed_by_user = TRUE
		WHERE tenant_id = ? AND mbti_result_id = ? AND deleted_at IS NULL`,
		tenantID, mbtiResultID,
	)
	if err != nil {
		return fmt.Errorf("mbti confirm: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("mbti confirm: 结果不存在或越权")
	}
	return nil
}

// scanner 兼容 *sql.Row 和 *sql.Rows。
type scanner interface {
	Scan(dest ...any) error
}

func scanResult(s scanner) (*Result, error) {
	res := &Result{}
	var convID, testURL, assert, fileID, ocrID sql.NullString
	var energy, mind, nature, tactics, identity sql.NullInt64
	var testedAt sql.NullTime
	err := s.Scan(&res.MBTIResultID, &res.UserID, &convID, &res.Source, &testURL, &res.ResultType,
		&assert, &energy, &mind, &nature, &tactics, &identity, &fileID, &ocrID,
		&res.ConfidenceScore, &res.ConfirmedByUser, &testedAt, &res.CreatedAt)
	if err != nil {
		return nil, err
	}
	res.ConversationID = convID.String
	res.TestURL = testURL.String
	res.Assertiveness = assert.String
	res.FileID = fileID.String
	res.OCRID = ocrID.String
	if energy.Valid {
		v := int(energy.Int64)
		res.EnergyScore = &v
	}
	if mind.Valid {
		v := int(mind.Int64)
		res.MindScore = &v
	}
	if nature.Valid {
		v := int(nature.Int64)
		res.NatureScore = &v
	}
	if tactics.Valid {
		v := int(tactics.Int64)
		res.TacticsScore = &v
	}
	if identity.Valid {
		v := int(identity.Int64)
		res.IdentityScore = &v
	}
	if testedAt.Valid {
		res.TestedAt = &testedAt.Time
	}
	return res, nil
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullInt(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}
