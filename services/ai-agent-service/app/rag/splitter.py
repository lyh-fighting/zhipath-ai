"""文档切分：按段落切分，chunk 带版本和审核状态。"""
from __future__ import annotations

import hashlib
from dataclasses import dataclass, field


@dataclass
class Chunk:
    chunk_id: str
    doc_id: str
    domain: str
    chunk_index: int
    content: str
    content_hash: str
    version: str
    status: str = "published"  # draft|published|archived
    metadata: dict = field(default_factory=dict)


def split_document(
    content: str,
    doc_id: str,
    domain: str,
    version: str = "1",
    max_length: int = 500,
    overlap: int = 50,
) -> list[Chunk]:
    """按段落切分，超长再切。chunk 带 version 和 status。"""
    paragraphs = [p.strip() for p in content.split("\n\n") if p.strip()]
    chunks: list[Chunk] = []
    idx = 0
    for para in paragraphs:
        if len(para) > max_length:
            for i in range(0, len(para), max_length - overlap):
                text = para[i : i + max_length]
                chunks.append(_make_chunk(doc_id, domain, version, idx, text))
                idx += 1
        else:
            chunks.append(_make_chunk(doc_id, domain, version, idx, para))
            idx += 1
    return chunks


def _make_chunk(doc_id: str, domain: str, version: str, idx: int, text: str) -> Chunk:
    return Chunk(
        chunk_id=f"{doc_id}_c{idx}",
        doc_id=doc_id,
        domain=domain,
        chunk_index=idx,
        content=text,
        content_hash=hashlib.sha256(text.encode()).hexdigest(),
        version=version,
        status="published",
    )
