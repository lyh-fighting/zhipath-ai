package platform

import "context"

// ObjectStore 抽象对象存储。
// 本地使用 S3-compatible MinIO adapter；云上通过实现此接口替换为 OSS/COS。
type ObjectStore interface {
	Ping(ctx context.Context) error
	BucketExists(ctx context.Context, bucket string) (bool, error)
}

// minioObjectStore MinIO 适配器，Task 14 接入 minio-go 完善真实实现。
type minioObjectStore struct {
	endpoint  string
	accessKey string
	secretKey string
	useSSL    bool
}

// NewObjectStore 构造 MinIO 对象存储适配器。
func NewObjectStore(endpoint, accessKey, secretKey string, useSSL bool) ObjectStore {
	return &minioObjectStore{endpoint, accessKey, secretKey, useSSL}
}

func (o *minioObjectStore) Ping(_ context.Context) error {
	// TODO Task 14: minio-go client Ping
	return nil
}

func (o *minioObjectStore) BucketExists(_ context.Context, _ string) (bool, error) {
	// TODO Task 14
	return false, nil
}
