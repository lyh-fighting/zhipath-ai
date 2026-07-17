"""重排序（骨架，接 cross-encoder 模型后完善）。"""
from __future__ import annotations

from collections.abc import Sequence

from .retriever import RetrievedDoc


def rerank(docs: Sequence[RetrievedDoc], query: str = "") -> list[RetrievedDoc]:
    """按 score 降序重排。"""
    return sorted(docs, key=lambda d: d.score, reverse=True)
