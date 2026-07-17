"""知识库导入任务。

MySQL 存 chunk 元数据，通过 Outbox 同步 Qdrant。
同一事件重复消费不产生重复 point（基于 event_id 幂等）。
"""
from __future__ import annotations

import hashlib


def ingest_document(content: str, doc_id: str, domain: str, version: str = "1") -> list[dict]:
    """切分文档 -> 写 MySQL chunk 元数据 -> 发 Outbox 事件同步 Qdrant。

    返回 chunk 列表（含 chunk_id/doc_id/domain/version/status）。
    """
    chunks = _split(content, doc_id, domain, version)
    # TODO: 事务内写 knowledge_chunks 表 + 发 outbox_events(chunk_upserted)
    return chunks


def _split(content: str, doc_id: str, domain: str, version: str, max_length: int = 500, overlap: int = 50) -> list[dict]:
    paragraphs = [p.strip() for p in content.split("\n\n") if p.strip()]
    chunks: list[dict] = []
    idx = 0
    for para in paragraphs:
        if len(para) > max_length:
            for i in range(0, len(para), max_length - overlap):
                chunks.append(_make_chunk(doc_id, domain, version, idx, para[i : i + max_length]))
                idx += 1
        else:
            chunks.append(_make_chunk(doc_id, domain, version, idx, para))
            idx += 1
    return chunks


def _make_chunk(doc_id: str, domain: str, version: str, idx: int, text: str) -> dict:
    return {
        "chunk_id": f"{doc_id}_c{idx}",
        "doc_id": doc_id,
        "domain": domain,
        "chunk_index": idx,
        "content": text,
        "content_hash": hashlib.sha256(text.encode()).hexdigest(),
        "version": version,
        "status": "published",
    }


def is_duplicate(event_id: str, seen: set[str]) -> bool:
    """幂等：同一 event_id 重复消费不产生重复 point。"""
    return event_id in seen
