"""API 入口测试。"""
from fastapi.testclient import TestClient

from app.main import app


def test_healthz():
    client = TestClient(app)
    r = client.get("/healthz")
    assert r.status_code == 200
    assert r.json()["status"] == "ok"


def test_readyz():
    client = TestClient(app)
    r = client.get("/readyz")
    assert r.status_code == 200


def test_agent_invoke_no_token_rejected():
    """无凭证调用必须被拒绝。"""
    client = TestClient(app)
    r = client.post("/internal/v1/agent/invoke", json={"message": "test"})
    assert r.status_code == 401


def test_agent_invoke_mock():
    """带正确 token 返回 mock 结果。"""
    from app.config import settings

    settings.internal_service_token = "test-token"
    client = TestClient(app)
    r = client.post(
        "/internal/v1/agent/invoke",
        json={"message": "我想换工作", "trace_id": "t1"},
        headers={"X-Internal-Token": "test-token"},
    )
    assert r.status_code == 200
    body = r.json()
    assert body["code"] == "SUCCESS"
    assert body["trace_id"] == "t1"
