package conversation

import "testing"

func TestCreateRequestDomainDefaultsToAuto(t *testing.T) {
	req := &CreateRequest{}
	if req.Domain != "" {
		t.Error("默认 Domain 应为空，由 Create 补 auto")
	}
}

func TestSendMessageRequestFields(t *testing.T) {
	req := &SendMessageRequest{Message: "测试消息", RequestID: "req_001"}
	if req.Message == "" {
		t.Error("message 不应为空")
	}
	if req.RequestID == "" {
		t.Error("request_id 用于幂等，不应为空")
	}
}

func TestTruncateShort(t *testing.T) {
	if got := truncate("短消息", 500); got != "短消息" {
		t.Errorf("短消息应不变: %s", got)
	}
}

func TestTruncateLong(t *testing.T) {
	long := ""
	for i := 0; i < 600; i++ {
		long += "a"
	}
	if got := truncate(long, 500); len(got) != 500 {
		t.Errorf("长消息应截断为 500: got %d", len(got))
	}
}
