"""知识导入测试。"""
from app.jobs.knowledge_ingest import ingest_document, is_duplicate


def test_ingest_document_splits():
    content = "第一段。\n\n第二段。"
    chunks = ingest_document(content, "doc_001", "emotion", "1")
    assert len(chunks) == 2
    assert chunks[0]["doc_id"] == "doc_001"
    assert chunks[0]["status"] == "published"
    assert chunks[0]["content_hash"]


def test_ingest_idempotent_by_event_id():
    seen: set[str] = set()
    assert not is_duplicate("evt_001", seen)
    seen.add("evt_001")
    assert is_duplicate("evt_001", seen)
