"""回访/复盘/通知（代码实现，非 n8n）。

3 天回访、7 天复盘、报告完成通知，均由 Worker 直接调度与发送。
"""
from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime, timedelta


@dataclass
class FollowupTask:
    task_id: str
    user_id: str
    conversation_id: str
    task_type: str  # followup_3d|review_7d|report_done
    scheduled_at: datetime
    status: str = "pending"


def schedule_followup(payload: dict) -> FollowupTask:
    """3 天回访。"""
    return FollowupTask(
        task_id=f"fu_{payload.get('conversation_id', '')}",
        user_id=payload.get("user_id", ""),
        conversation_id=payload.get("conversation_id", ""),
        task_type="followup_3d",
        scheduled_at=datetime.now() + timedelta(days=3),
    )


def schedule_review(payload: dict) -> FollowupTask:
    """7 天复盘。"""
    return FollowupTask(
        task_id=f"rv_{payload.get('conversation_id', '')}",
        user_id=payload.get("user_id", ""),
        conversation_id=payload.get("conversation_id", ""),
        task_type="review_7d",
        scheduled_at=datetime.now() + timedelta(days=7),
    )


def notify_report_done(payload: dict) -> dict:
    """报告完成通知（微信订阅消息/站内信，非 n8n）。"""
    # TODO: 调微信 subscribeMessage.send / 站内信
    return {"notified": True, "user_id": payload.get("user_id", "")}
