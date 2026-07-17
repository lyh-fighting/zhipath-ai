"""事实核查（骨架，Task 13 接 RAG 后完善）。"""
from __future__ import annotations


def check_grounding(answer: str, retrieved_docs: list[dict]) -> bool:
    """检查答案是否有事实依据。无检索结果时不扣分。"""
    if not retrieved_docs:
        return True
    return True
