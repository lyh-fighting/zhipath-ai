"""OpenAI-compatible provider（支持 DeepSeek / OpenAI base URL 切换）。

fallback 只处理超时、限流、5xx；参数及安全错误向上抛出不吞掉。
"""
from __future__ import annotations

from typing import Any

from openai import APIError, APIStatusError, APITimeoutError, OpenAI, RateLimitError

from .base import ModelMessage, ModelResponse


class OpenAICompatibleProvider:
    def __init__(self, name: str, api_key: str, base_url: str | None = None, model: str = ""):
        self.name = name
        self.model = model or name
        self._client = OpenAI(api_key=api_key, base_url=base_url) if api_key else None

    def invoke(self, messages: list[ModelMessage], **kwargs: Any) -> ModelResponse:
        if self._client is None:
            raise RuntimeError(f"{self.name} API key 未配置")
        try:
            resp = self._client.chat.completions.create(
                model=kwargs.get("model", self.model),
                messages=[{"role": m.role, "content": m.content} for m in messages],
                temperature=kwargs.get("temperature", 0.7),
                max_tokens=kwargs.get("max_tokens", 2000),
            )
        except (APITimeoutError, RateLimitError):
            raise  # 可重试
        except APIStatusError as e:
            if e.status_code >= 500:
                raise  # 可重试
            raise  # 4xx 参数/安全错误不重试
        except APIError:
            raise
        choice = resp.choices[0]
        usage = resp.usage
        return ModelResponse(
            content=choice.message.content or "",
            model=resp.model,
            input_tokens=usage.prompt_tokens if usage else 0,
            output_tokens=usage.completion_tokens if usage else 0,
            raw=resp.model_dump(),
        )
