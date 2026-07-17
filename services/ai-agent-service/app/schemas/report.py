"""深度报告 schema。记录现实依据/MBTI测试版本/MBTI分析/风险/30-90-180天行动。"""
from __future__ import annotations

from dataclasses import dataclass, field


@dataclass
class ReportSchema:
    report_id: str = ""
    conversation_id: str = ""
    mbti_result_id: str = ""
    mbti_snapshot: dict = field(default_factory=dict)
    reality: str = ""
    mbti_analysis: str = ""
    risks: list = field(default_factory=list)
    action_plan_30d: list = field(default_factory=list)
    action_plan_90d: list = field(default_factory=list)
    action_plan_180d: list = field(default_factory=list)
    status: str = "pending"
