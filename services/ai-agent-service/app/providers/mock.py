"""Mock provider，返回确定性结果，供本地 E2E 使用。

显式标记 provider=mock，不静默伪造成功。
"""
from __future__ import annotations

from typing import Any

from .base import ModelMessage, ModelResponse


class MockProvider:
    name = "mock"

    def invoke(self, messages: list[ModelMessage], **kwargs: Any) -> ModelResponse:
        user_msg = next((m for m in messages if m.role == "user"), None)
        prompt = user_msg.content if user_msg else ""
        return ModelResponse(
            content=f"[mock] 针对「{prompt[:80]}」的确定性回复",
            model="mock-1",
            input_tokens=max(1, len(prompt) // 4),
            output_tokens=20,
        )
