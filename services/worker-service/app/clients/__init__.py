"""Qdrant 客户端封装。upsert 幂等（同 point_id 覆盖，不产生重复）。"""
from __future__ import annotations

from typing import Any


class QdrantClient:
    def __init__(self, url: str):
        self.url = url

    def upsert(self, collection: str, point_id: str, vector: list, payload: dict) -> None:
        # TODO: qdrant_client.upsert（幂等）
        pass

    def delete(self, collection: str, point_id: str) -> None:
        pass

    def ping(self) -> bool:
        return True
