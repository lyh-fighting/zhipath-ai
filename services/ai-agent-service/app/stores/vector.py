"""Qdrant 向量存储客户端封装。"""
from __future__ import annotations

from qdrant_client import QdrantClient


class VectorStore:
    """语义记忆、知识库检索。"""

    def __init__(self, url: str):
        self._client = QdrantClient(url=url)

    def ping(self) -> bool:
        try:
            self._client.get_collections()
            return True
        except Exception:
            return False
