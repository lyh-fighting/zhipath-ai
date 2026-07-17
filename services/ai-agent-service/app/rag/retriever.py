"""检索：仅检索 status=published 且 domain/version 匹配的 chunk。"""
from __future__ import annotations

from dataclasses import dataclass


@dataclass
class RetrievedDoc:
    chunk_id: str
    doc_id: str
    domain: str
    content: str
    score: float
    version: str
    status: str


class Retriever:
    """向量检索。embedding provider 可配置，本地用确定性 fake embedding。"""

    def __init__(self, qdrant_url: str, collection: str = "zhipath_kb_v1"):
        self.qdrant_url = qdrant_url
        self.collection = collection

    def retrieve(
        self, query_vector: list[float], domain: str, version: str, limit: int = 5
    ) -> list[RetrievedDoc]:
        """检索。filter: status=published AND domain AND version。"""
        # TODO: qdrant_client.search with filter
        return []

    def filter(self, domain: str, version: str) -> dict:
        """构建 Qdrant 过滤条件。"""
        return {
            "must": [
                {"key": "status", "match": {"value": "published"}},
                {"key": "domain", "match": {"value": domain}},
                {"key": "version", "match": {"value": version}},
            ]
        }


def fake_embedding(text: str, dim: int = 1536) -> list[float]:
    """确定性 fake embedding，供本地测试。"""
    import hashlib

    h = hashlib.sha256(text.encode()).digest()
    vec = []
    for i in range(dim):
        vec.append((h[i % len(h)] / 255.0) * 2 - 1)
    return vec
