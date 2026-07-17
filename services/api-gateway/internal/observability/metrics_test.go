package observability

import "testing"

func TestIncRiskHit(t *testing.T) {
	before := Default.RiskHitCount
	IncRiskHit()
	if Default.RiskHitCount != before+1 {
		t.Error("IncRiskHit 应递增")
	}
}

func TestAddToken(t *testing.T) {
	before := Default.TokenConsumed
	AddToken(100)
	if Default.TokenConsumed != before+100 {
		t.Error("AddToken 应累加")
	}
}
