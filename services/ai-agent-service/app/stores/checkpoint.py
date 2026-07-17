"""LangGraph checkpoint，以 tenant_id:user_id:conversation_id 为 thread key。

中断后用同一 thread key 恢复执行，不重复调用已完成工具。
"""
from __future__ import annotations


def thread_key(tenant_id: str, user_id: str, conversation_id: str) -> str:
    """断点续跑的 thread key。"""
    return f"{tenant_id}:{user_id}:{conversation_id}"


# TODO Task 10: 接入 langgraph.checkpoint.redis.RedisSaver，配置 TTL
