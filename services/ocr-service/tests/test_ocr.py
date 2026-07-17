"""OCR Service 测试。"""
from fastapi.testclient import TestClient

from app.engine import MockOCREngine
from app.main import app


def test_mock_engine_deterministic():
    engine = MockOCREngine()
    r1 = engine.extract("http://example.com/file1.png")
    r2 = engine.extract("http://example.com/file1.png")
    assert r1.clean_text == r2.clean_text
    assert r1.avg_confidence > 0.8


def test_healthz():
    client = TestClient(app)
    r = client.get("/healthz")
    assert r.status_code == 200


def test_ocr_no_token_rejected():
    client = TestClient(app)
    r = client.post("/internal/v1/ocr/extract", json={"file_url": "http://x"})
    assert r.status_code == 401


def test_ocr_extract_mock():
    from app.main import settings

    settings.internal_service_token = "test-token"
    client = TestClient(app)
    r = client.post(
        "/internal/v1/ocr/extract",
        json={"file_url": "http://example.com/test.png", "trace_id": "t1"},
        headers={"X-Internal-Token": "test-token"},
    )
    assert r.status_code == 200
    body = r.json()
    assert body["code"] == "SUCCESS"
    assert body["data"]["clean_text"]
    assert body["trace_id"] == "t1"
