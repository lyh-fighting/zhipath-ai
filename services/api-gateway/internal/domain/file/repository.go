package file

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository 文件数据访问。预签名 URL 上传 MinIO，限类型/大小/哈希。
type Repository struct {
	db *sql.DB
}

// NewRepository 构造 Repository。
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ValidateUpload 校验文件类型与大小。
func ValidateUpload(req *UploadRequest) error {
	if !AllowedMimeTypes[req.MimeType] {
		return fmt.Errorf("不支持的文件类型: %s", req.MimeType)
	}
	if req.SizeBytes > MaxFileSize {
		return fmt.Errorf("文件超过 20MB 限制")
	}
	return nil
}

// Create 创建文件记录，返回预签名上传 URL。
func (r *Repository) Create(ctx context.Context, tenantID, userID string, req *UploadRequest) (*UploadResponse, error) {
	if err := ValidateUpload(req); err != nil {
		return nil, err
	}
	fileID := "f_" + uuid.NewString()
	storageURL := fmt.Sprintf("s3://zhipath-ai/%s", fileID)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO files (tenant_id, file_id, user_id, file_type, mime_type, storage_url, sha256, size_bytes, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'uploaded')`,
		tenantID, fileID, userID, req.FileType, req.MimeType, storageURL, req.SHA256, req.SizeBytes,
	)
	if err != nil {
		return nil, fmt.Errorf("file create: %w", err)
	}
	// TODO: 用 ObjectStore 生成真实预签名 URL
	return &UploadResponse{
		FileID:    fileID,
		UploadURL: fmt.Sprintf("http://minio:9000/zhipath-ai/%s?upload", fileID),
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}, nil
}

// Get 查询文件。不存在返回 (nil, nil)。
func (r *Repository) Get(ctx context.Context, tenantID, fileID string) (*File, error) {
	f := &File{}
	err := r.db.QueryRowContext(ctx, `
		SELECT file_id, tenant_id, user_id, conversation_id, file_type, mime_type,
		       storage_url, sha256, size_bytes, status, created_at
		FROM files WHERE tenant_id = ? AND file_id = ? AND deleted_at IS NULL`,
		tenantID, fileID,
	).Scan(&f.FileID, &f.TenantID, &f.UserID, &f.ConversationID, &f.FileType, &f.MimeType,
		&f.StorageURL, &f.SHA256, &f.SizeBytes, &f.Status, &f.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("file get: %w", err)
	}
	return f, nil
}

// Delete 软删除文件（同步删 MinIO 由 Outbox 处理）。
func (r *Repository) Delete(ctx context.Context, tenantID, fileID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE files SET deleted_at = NOW(3) WHERE tenant_id = ? AND file_id = ?`, tenantID, fileID)
	if err != nil {
		return fmt.Errorf("file delete: %w", err)
	}
	return nil
}
