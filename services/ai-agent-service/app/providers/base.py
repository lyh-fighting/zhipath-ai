"""模型 provider 接口与结构化输出协议。

fallback 只处理超时、限流和 5xx，不吞掉参数及安全错误。
"""
from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Protocol


@dataclass
class ModelMessage:
    role: str  # system | user | assistant
    content: str


@dataclass
class ModelResponse:
    content: str
    model: str
    input_tokens: int = 0
    output_tokens: int = 0
    raw: dict[str, Any] | None = None


class ModelProvider(Protocol):
    """模型 provider 接口。"""

    name: str

    def invoke(self, messages: list[ModelMessage], **kwargs: Any) -> ModelResponse: ...


