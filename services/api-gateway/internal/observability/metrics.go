package observability

// 指标定义（骨架，接 Prometheus 后完善）。
// 覆盖：QPS / p95 / p99 / 模型耗时 / token / 风险命中 / Outbox 延迟。
// 日志通过同一 trace_id 串联（见 middleware/trace.go）。
type Metrics struct {
	RequestCount     int64
	RequestLatencyMs int64
	ModelCallMs      int64
	TokenConsumed    int64
	RiskHitCount     int64
	OutboxLagMs      int64
}

// Default 全局指标实例。
var Default = &Metrics{}

// IncRiskHit 风险命中计数。
func IncRiskHit() { Default.RiskHitCount++ }

// AddToken token 消耗累加。
func AddToken(n int64) { Default.TokenConsumed += n }
