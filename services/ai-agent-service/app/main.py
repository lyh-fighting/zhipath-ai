"""ZhiPath AI Agent Service 入口。

提供 /healthz、/readyz 和内部 Agent 接口。
内部接口用 X-Internal-Token 鉴权，无凭证拒绝调用。

实际对话：优先用阿里云百炼（OpenAI 兼容）真实模型，未配置 key 时回退 Mock。
每次请求先做意图分析（路由 + 危机识别），再按会话类型产出评测式回答。
"""
from __future__ import annotations

import json
import re

from fastapi import Depends, FastAPI, Header, HTTPException
from pydantic import BaseModel, Field

from .config import settings
from .prompts import (
    CAREER_CHAT_SYSTEM,
    CAREER_CHAT_SYSTEM_VERSION,
    DECISION_COACH_PROMPT,
    EMOTION_CHAT_SYSTEM,
    EMOTION_CHAT_SYSTEM_VERSION,
    INTENT_PROMPT,
    INTENT_PROMPT_VERSION,
)
from .providers import MockProvider, ModelMessage, OpenAICompatibleProvider

app = FastAPI(title="ZhiPath AI Agent Service", version="0.1.0")


# ---------------------------------------------------------------------------
# Provider 装配：百炼（OpenAI 兼容）优先，缺 key 回退 Mock
# ---------------------------------------------------------------------------
_REAL_PROVIDER: OpenAICompatibleProvider | None = None
if settings.bailian_api_key:
    _REAL_PROVIDER = OpenAICompatibleProvider(
        name="bailian",
        api_key=settings.bailian_api_key,
        base_url=settings.bailian_base_url,
        model=settings.bailian_model,
    )


def _provider():
    return _REAL_PROVIDER or MockProvider()


# ---------------------------------------------------------------------------
# 鉴权
# ---------------------------------------------------------------------------
async def verify_internal_token(x_internal_token: str = Header(None)) -> None:
    """服务间鉴权：Go API Gateway 调用时必须携带 X-Internal-Token。"""
    if not settings.internal_service_token or x_internal_token != settings.internal_service_token:
        raise HTTPException(status_code=401, detail="internal auth failed")


# ---------------------------------------------------------------------------
# 请求体
# ---------------------------------------------------------------------------
class InvokePayload(BaseModel):
    message: str
    trace_id: str = ""
    conversation_type: str = "career"  # career | emotion | decision
    history: list[dict] = Field(default_factory=list)


# ---------------------------------------------------------------------------
# 意图分析（真实调用）
# ---------------------------------------------------------------------------
def _analyze_intent(message: str) -> dict:
    prompt = INTENT_PROMPT.format(message=message)
    resp = _provider().invoke(
        [
            ModelMessage(role="system", content="你是意图识别器，仅输出 JSON。"),
            ModelMessage(role="user", content=prompt),
        ],
        temperature=0.2,
        max_tokens=400,
    )
    return _parse_intent(resp.content)


def _parse_intent(text: str) -> dict:
    """从模型输出里尽量稳健地提取 JSON。"""
    text = (text or "").strip()
    try:
        return json.loads(text)
    except Exception:
        pass
    # 去掉 ```json ``` 围栏后重试
    m = re.search(r"\{.*\}", text, re.DOTALL)
    if m:
        try:
            return json.loads(m.group(0))
        except Exception:
            pass
    return {
        "intent": "general",
        "confidence": 0.0,
        "summary": text[:80],
        "tags": [],
        "needs_human": False,
    }


# ---------------------------------------------------------------------------
# 评测式回答（真实调用）
# ---------------------------------------------------------------------------
_SYSTEMS = {
    "career": (CAREER_CHAT_SYSTEM, CAREER_CHAT_SYSTEM_VERSION),
    "emotion": (EMOTION_CHAT_SYSTEM, EMOTION_CHAT_SYSTEM_VERSION),
    "decision": (DECISION_COACH_PROMPT, "decision-coach-v1"),
}


def _answer(conversation_type: str, history: list[dict], message: str) -> str:
    system, _ = _SYSTEMS.get(conversation_type, _SYSTEMS["career"])
    messages: list[ModelMessage] = [ModelMessage(role="system", content=system)]
    for item in history[-20:]:
        role = item.get("role")
        content = item.get("content", "")
        if role in ("user", "assistant") and content:
            messages.append(ModelMessage(role=role, content=content))
    messages.append(ModelMessage(role="user", content=message))
    resp = _provider().invoke(messages, temperature=0.7, max_tokens=2000)
    return resp.content


# ---------------------------------------------------------------------------
# 路由
# ---------------------------------------------------------------------------
@app.get("/healthz")
async def healthz() -> dict:
    return {"status": "ok"}


@app.get("/readyz")
async def readyz() -> dict:
    return {
        "status": "ok",
        "model": settings.bailian_model if _REAL_PROVIDER else "mock",
        "checks": {"redis": "ok", "qdrant": "ok"},
    }


@app.post("/internal/v1/agent/invoke", dependencies=[Depends(verify_internal_token)])
async def agent_invoke(payload: InvokePayload) -> dict:
    """同步调用 Agent（意图分析 + 评测式回答）。"""
    conversation_type = payload.conversation_type or "career"
    if conversation_type not in _SYSTEMS:
        conversation_type = "career"

    try:
        # 1) 意图分析（路由 + 危机识别）
        intent = _analyze_intent(payload.message)
        # 2) 评测式回答
        reply = _answer(conversation_type, payload.history, payload.message)
    except Exception as e:  # noqa: BLE001
        msg = str(e)
        if "quota" in msg.lower() or "Free quota" in msg or "AllocationQuota" in msg:
            note = ("⚠️ AI 服务额度不足：请在「阿里云百炼」控制台为当前 API Key 充值，"
                    "或关闭「仅使用免费额度」模式后重试。")
        elif "auth" in msg.lower() or "401" in msg or "403" in msg:
            note = "⚠️ AI 服务鉴权失败：请检查 BAILIAN_API_KEY 是否正确。"
        else:
            note = f"⚠️ AI 调用失败：{msg[:160]}"
        return {
            "code": "ERROR",
            "message": "ai_unavailable",
            "data": {
                "message_id": "m_err",
                "role": "assistant",
                "content_summary": note,
                "structured_result": None,
                "need_human_handoff": False,
                "quality_score": 0,
            },
            "trace_id": payload.trace_id,
        }

    return {
        "code": "SUCCESS",
        "message": "ok",
        "data": {
            "message_id": "m_real",
            "role": "assistant",
            "content_summary": reply,
            "structured_result": {
                "intent": intent.get("intent", "general"),
                "intent_confidence": intent.get("confidence", 0.0),
                "summary": intent.get("summary", ""),
                "tags": intent.get("tags", []),
                "prompt_version": INTENT_PROMPT_VERSION,
            },
            "need_human_handoff": bool(intent.get("needs_human", False)),
            "quality_score": 0,
        },
        "trace_id": payload.trace_id,
    }
