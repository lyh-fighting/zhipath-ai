"""ZhiPath AI Agent Service 入口。

提供 /healthz、/readyz 和内部 Agent 接口。
内部接口用 X-Internal-Token 鉴权，无凭证拒绝调用。
"""
from fastapi import Depends, FastAPI, Header, HTTPException

from .config import settings

app = FastAPI(title="ZhiPath AI Agent Service", version="0.1.0")


async def verify_internal_token(x_internal_token: str = Header(None)) -> None:
    """服务间鉴权：Go API Gateway 调用时必须携带 X-Internal-Token。"""
    if not settings.internal_service_token or x_internal_token != settings.internal_service_token:
        raise HTTPException(status_code=401, detail="internal auth failed")


@app.get("/healthz")
async def healthz() -> dict:
    return {"status": "ok"}


@app.get("/readyz")
async def readyz() -> dict:
    # TODO Task 10: 检查 Redis/Qdrant 真实连通性
    return {"status": "ok", "checks": {"redis": "ok", "qdrant": "ok"}}


@app.post("/internal/v1/agent/invoke", dependencies=[Depends(verify_internal_token)])
async def agent_invoke(payload: dict) -> dict:
    """同步调用 Agent（Task 10 接入 LangGraph）。

    Mock 阶段返回确定性结果，供本地 E2E 使用。
    """
    trace_id = payload.get("trace_id", "")
    message = payload.get("message", "")
    # TODO Task 10: 初始化 LangGraph state 并执行
    return {
        "code": "SUCCESS",
        "message": "ok",
        "data": {
            "message_id": "m_mock",
            "role": "assistant",
            "content_summary": f"mock 回复: {message[:50]}",
            "structured_result": None,
            "need_human_handoff": False,
            "quality_score": 0,
        },
        "trace_id": trace_id,
    }
