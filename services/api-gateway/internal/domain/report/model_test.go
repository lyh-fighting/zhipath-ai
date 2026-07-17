package report

import "testing"

func TestCanRetry(t *testing.T) {
	if !CanRetry("pending") {
		t.Error("pending 应可重试")
	}
	if !CanRetry("failed") {
		t.Error("failed 应可重试")
	}
	if CanRetry("completed") {
		t.Error("completed 不应重试")
	}
}

func TestReportFields(t *testing.T) {
	r := Report{
		ReportID:     "rep_001",
		MBTIResultID: "mbti_001",
		Status:       "pending",
	}
	if r.ReportID != "rep_001" {
		t.Error("ReportID 不匹配")
	}
	if r.MBTIResultID == "" {
		t.Error("MBTIResultID 应记录")
	}
}
