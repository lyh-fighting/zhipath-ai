"""Anthropic Claude provider（httpx 封装，不引入 anthropic SDK）。"""
from __future__ import annotations

import httpx

from .base import ModelMessage, ModelResponse
from typing import Any


class AnthropicProvider:
    name = "anthropic"

    def __init__(self, api_key: str, model: str = "claude-3-5-sonnet-20241022"):
        self.api_key = api_key
        self.model = model

    def invoke(self, messages: list[ModelMessage], **kwargs: Any) -> ModelResponse:
        if not self.api_key:
            raise RuntimeError("anthropic API key 未配置")
        system = "\n".join(m.content for m in messages if m.role == "system")
        conv = [{"role": m.role, "content": m.content} for m in messages if m.role != "system"]
        with httpx.Client(timeout=60) as client:
            resp = client.post(
                "https://api.anthropic.com/v1/messages",
                headers={
                    "x-api-key": self.api_key,
                    "anthropic-version": "2023-06-01",
                    "content-type": "application/json",
                },
                json={
                    "model": kwargs.get("model", self.model),
                    "max_tokens": kwargs.get("max_tokens", 2000),
                    "system": system,
                    "messages": conv,
                },
            )
            resp.raise_for_status()
            data = resp.json()
        blocks = data.get("content", [])
        content = blocks[0].get("text", "") if blocks else ""
        usage = data.get("usage", {})
        return ModelResponse(
            content=content,
            model=data.get("model", self.model),
            input_tokens=usage.get("input_tokens", 0),
            output_tokens=usage.get("output_tokens", 0),
            raw=data,
        )
