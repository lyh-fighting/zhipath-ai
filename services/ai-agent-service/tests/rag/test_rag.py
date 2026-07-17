"""RAG 模块测试。"""
from app.rag import fake_embedding, rerank, split_document
from app.rag.retriever import RetrievedDoc, Retriever


def test_split_document_by_paragraph():
    content = "第一段内容。\n\n第二段内容。\n\n第三段内容。"
    chunks = split_document(content, "doc_001", "emotion", "1")
    assert len(chunks) == 3
    assert chunks[0].doc_id == "doc_001"
    assert chunks[0].domain == "emotion"
    assert chunks[0].version == "1"
    assert chunks[0].status == "published"
    assert chunks[0].content_hash


def test_split_long_paragraph():
    content = "a" * 1200
    chunks = split_document(content, "doc_002", "career", "1", max_length=500, overlap=50)
    assert len(chunks) >= 3
    for c in chunks:
        assert len(c.content) <= 500


def test_fake_embedding_deterministic():
    v1 = fake_embedding("test", 128)
    v2 = fake_embedding("test", 128)
    assert v1 == v2
    assert len(v1) == 128


def test_rerank_desc_by_score():
    docs = [
        RetrievedDoc("c1", "d1", "emotion", "x", 0.5, "1", "published"),
        RetrievedDoc("c2", "d1", "emotion", "y", 0.9, "1", "published"),
        RetrievedDoc("c3", "d1", "emotion", "z", 0.3, "1", "published"),
    ]
    ranked = rerank(docs)
    assert ranked[0].score == 0.9
    assert ranked[-1].score == 0.3


def test_retriever_filter_published_domain_version():
    r = Retriever("http://localhost:6333")
    f = r.filter("emotion", "1")
    must = f["must"]
    assert must[0]["match"]["value"] == "published"
    assert must[1]["match"]["value"] == "emotion"
    assert must[2]["match"]["value"] == "1"
