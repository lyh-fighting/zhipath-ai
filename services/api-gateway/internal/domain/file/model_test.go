package file

import "testing"

func TestValidateUploadOK(t *testing.T) {
	req := &UploadRequest{FileType: "image", MimeType: "image/png", SizeBytes: 1024}
	if err := ValidateUpload(req); err != nil {
		t.Fatalf("合法文件应通过: %v", err)
	}
}

func TestValidateUploadBadMime(t *testing.T) {
	req := &UploadRequest{FileType: "image", MimeType: "application/exe", SizeBytes: 1024}
	if err := ValidateUpload(req); err == nil {
		t.Fatal("不支持的类型应拒绝")
	}
}

func TestValidateUploadTooLarge(t *testing.T) {
	req := &UploadRequest{FileType: "image", MimeType: "image/png", SizeBytes: MaxFileSize + 1}
	if err := ValidateUpload(req); err == nil {
		t.Fatal("超限文件应拒绝")
	}
}

func TestAllowedMimeTypes(t *testing.T) {
	if !AllowedMimeTypes["image/png"] {
		t.Error("image/png 应被允许")
	}
	if AllowedMimeTypes["application/exe"] {
		t.Error("exe 不应被允许")
	}
}
