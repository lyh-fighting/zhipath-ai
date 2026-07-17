"""Outbox 轮询器与处理器。

行锁批量获取 pending 事件，event_id 幂等，失败指数退避，达上限进 dead 并报警。
三类 Qdrant handler：用户记忆/会话摘要/知识 chunk。
回访/复盘/通知用代码实现（非 n8n）。
"""
from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Callable


@dataclass
class OutboxEvent:
    event_id: str
    event_type: str
    aggregate_type: str
    aggregate_id: str
    payload: dict
    status: str = "pending"
    retry_count: int = 0
    max_retries: int = 5


def backoff_seconds(retry_count: int) -> int:
    """指数退避：2^retry 秒，上限 300。"""
    return min(2 ** retry_count, 300)


def is_dead(retry_count: int, max_retries: int = 5) -> bool:
    """达上限进 dead 状态并报警。"""
    return retry_count >= max_retries


# ===== handler 注册 =====
_HANDLERS: dict[str, Callable] = {}


def register(event_type: str):
    def deco(fn):
        _HANDLERS[event_type] = fn
        return fn
    return deco


@register("memory_upserted")
def handle_memory_upserted(event: OutboxEvent, qdrant: Any) -> None:
    """用户记忆同步 Qdrant（zhipath_user_memory_v1）。"""


@register("summary_upserted")
def handle_summary_upserted(event: OutboxEvent, qdrant: Any) -> None:
    """会话摘要同步 Qdrant（zhipath_conversation_summary_v1）。"""


@register("chunk_upserted")
def handle_chunk_upserted(event: OutboxEvent, qdrant: Any) -> None:
    """知识 chunk 同步 Qdrant（zhipath_kb_v1）。同一事件重复消费不产生重复 point。"""


@register("memory_deleted")
def handle_memory_deleted(event: OutboxEvent, qdrant: Any) -> None:
    """删除 Qdrant point。"""


@register("file_deleted")
def handle_file_deleted(event: OutboxEvent, object_store: Any) -> None:
    """删除 MinIO 文件。"""


@register("report_done")
def handle_report_done(event: OutboxEvent, notifier: Any) -> None:
    """报告完成通知（代码实现，非 n8n）。"""
    from ..jobs.notification import notify_report_done
    notify_report_done(event.payload)


@register("followup_trigger")
def handle_followup_trigger(event: OutboxEvent, scheduler: Any) -> None:
    """回访/复盘触发（代码实现定时任务，非 n8n）。"""
    from ..jobs.notification import schedule_followup, schedule_review
    task_type = event.payload.get("task_type", "followup_3d")
    if task_type == "review_7d":
        schedule_review(event.payload)
    else:
        schedule_followup(event.payload)


def dispatch(event: OutboxEvent, deps: dict) -> None:
    """分派事件到对应 handler。"""
    handler = _HANDLERS.get(event.event_type)
    if handler:
        handler(event, deps)
