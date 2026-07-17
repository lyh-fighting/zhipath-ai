"""深度报告生成任务。

记录现实依据/MBTI分析/风险/30-90-180天行动。
报告任务可重试且不重复扣权益。完成后通过 Outbox 触发通知。
文件写 MinIO，用户只能访问自己的预签名 URL。
"""
from __future__ import annotations

from dataclasses import dataclass


@dataclass
class ReportJob:
    report_id: str
    user_id: str
    conversation_id: str
    mbti_result_id: str
    status: str = "pending"


def generate_report(job: ReportJob, mbti_snapshot: dict) -> dict:
    """生成报告。"""
    return {
        "report_id": job.report_id,
        "mbti_snapshot": mbti_snapshot,
        "reality": "",
        "mbti_analysis": "",
        "risks": [],
        "action_plan_30d": [],
        "action_plan_90d": [],
        "action_plan_180d": [],
        "status": "completed",
        # TODO: 写 MinIO + 发 Outbox(report_done) 触发通知
    }


def can_retry(status: str) -> bool:
    """报告任务可重试，不重复扣权益。"""
    return status in ("pending", "generating", "failed")
