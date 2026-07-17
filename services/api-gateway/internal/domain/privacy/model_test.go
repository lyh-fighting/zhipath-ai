package privacy

import "testing"

func TestConsentTypeConstants(t *testing.T) {
	types := []string{ConsentStoreMBTI, ConsentStoreScreenshot, ConsentLongTermMemory, ConsentMarketing}
	for _, ct := range types {
		if ct == "" {
			t.Error("consent type 不应为空")
		}
	}
}

func TestExportRequestIndependent(t *testing.T) {
	req := &ExportRequest{IncludeProfile: true, IncludeMBTI: false, IncludeChats: true, IncludeReports: false}
	if !req.IncludeProfile {
		t.Error("IncludeProfile 应可独立设置")
	}
	if req.IncludeMBTI {
		t.Error("IncludeMBTI 应为 false")
	}
}
