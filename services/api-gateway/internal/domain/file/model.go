package file

import "time"

// File 文件记录。
type File struct {
	FileID         string    `json:"file_id"`
	TenantID       string    `json:"tenant_id"`
	UserID         string    `json:"user_id"`
	ConversationID string    `json:"conversation_id,omitempty"`
	FileType       string    `json:"file_type"` // image|pdf|document
	MimeType       string    `json:"mime_type"`
	StorageURL     string    `json:"storage_url"`
	SHA256         string    `json:"sha256"`
	SizeBytes      int64     `json:"size_bytes"`
	UploadClient   string    `json:"upload_client,omitempty"`
	Status         string    `json:"status"` // uploaded|deleted
	CreatedAt      time.Time `json:"created_at"`
}

// UploadRequest 申请预签名上传 URL。
type UploadRequest struct {
	FileType  string `json:"file_type" vd:"required"`
	MimeType  string `json:"mime_type" vd:"required"`
	SizeBytes int64  `json:"size_bytes" vd:"required"`
	SHA256    string `json:"sha256"`
}

// UploadResponse 预签名上传 URL。
type UploadResponse struct {
	FileID    string    `json:"file_id"`
	UploadURL string    `json:"upload_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AllowedMimeTypes 允许的文件类型。
var AllowedMimeTypes = map[string]bool{
	"image/jpeg":       true,
	"image/png":        true,
	"image/webp":       true,
	"application/pdf":  true,
}

// MaxFileSize 文件大小上限 20MB。
const MaxFileSize = 20 * 1024 * 1024
