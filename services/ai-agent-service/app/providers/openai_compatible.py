"""OpenAI-compatible provider（支持 DeepSeek / OpenAI base URL 切换）。

fallback 处理超时、限流、5xx、403 配额耗尽、404 模型名不存在；参数及安全错误向上抛出不吞掉。
支持多 model 列表：额度耗尽或模型名不可用时按顺序自动切换到下一个模型。
"""
from __future__ import annotations

import concurrent.futures
import logging
from typing import Any

from openai import APIConnectionError, APIError, APIStatusError, APITimeoutError, OpenAI, RateLimitError

from .base import ModelMessage, ModelResponse

logger = logging.getLogger(__name__)


class OpenAICompatibleProvider:
    def __init__(
        self,
        name: str,
        api_key: str,
        base_url: str | None = None,
        model: str = "",
        models: list[str] | None = None,
        timeout: float = 60.0,
    ):
        self.name = name
        self.timeout = timeout
        self.models = list(models) if models else ([model] if model else [name])
        self._last_working_model: str | None = None
        # max_retries=0 让我们自己在多模型间切换，避免 SDK 在单模型上重试浪费时间
        self._client = OpenAI(api_key=api_key, base_url=base_url, max_retries=0) if api_key else None
        if self._client and len(self.models) > 1:
            self._reorder_models()

    def _probe_one(self, model: str, timeout: float) -> tuple[str, bool]:
        """用极小请求探测模型是否可达；返回 (model, reachable)。"""
        if self._client is None:
            return model, False
        try:
            self._client.chat.completions.create(
                model=model,
                messages=[{"role": "user", "content": "hi"}],
                max_tokens=1,
                timeout=timeout,
            )
            return model, True
        except APIStatusError as e:
            if e.status_code == 404:
                logger.info("[provider] model %s not found in OpenAI-compatible endpoint, will skip", model)
                return model, False
            # 403/429/5xx 等说明模型名存在，只是额度/限流/服务端问题，仍保留
            return model, True
        except Exception as e:
            logger.info("[provider] model %s probe failed (%s), will still try later", model, e)
            return model, True

    def _reorder_models(self):
        """启动时并行探测，把可达模型排在前面（仍保持原过期时间顺序），404 放到最后。"""
        probe_timeout = min(5.0, self.timeout)
        with concurrent.futures.ThreadPoolExecutor(max_workers=8, thread_name_prefix="model_probe") as pool:
            futures = [pool.submit(self._probe_one, m, probe_timeout) for m in self.models]
            results = [f.result() for f in concurrent.futures.as_completed(futures)]
        reachable = [m for m, ok in results if ok]
        unreachable = [m for m, ok in results if not ok]
        if unreachable:
            logger.info("[provider] reordering models: reachable=%s unreachable=%s", reachable, unreachable)
        self.models = reachable + unreachable

    def _is_retryable(self, e: Exception) -> bool:
        if isinstance(e, (APITimeoutError, APIConnectionError, RateLimitError)):
            return True
        if isinstance(e, APIStatusError):
            # 404/403=模型不可用或无权访问，切换下一个模型；5xx=服务端问题，可重试
            if e.status_code in (404, 403, 500, 502, 503, 504):
                return True
        return False

    def invoke(self, messages: list[ModelMessage], **kwargs: Any) -> ModelResponse:
        if self._client is None:
            raise RuntimeError(f"{self.name} API key 未配置")
        last_error: Exception | None = None
        # 优先使用上一次成功调用的模型，降低每次请求都从头试 403/404 的开销
        ordered = self.models
        if self._last_working_model and self._last_working_model in self.models:
            ordered = [self._last_working_model] + [m for m in self.models if m != self._last_working_model]
        for idx, model in enumerate(ordered):
            try:
                resp = self._client.chat.completions.create(
                    model=model,
                    messages=[{"role": m.role, "content": m.content} for m in messages],
                    temperature=kwargs.get("temperature", 0.7),
                    max_tokens=kwargs.get("max_tokens", 2000),
                    timeout=self.timeout,
                )
            except (APITimeoutError, APIConnectionError, RateLimitError, APIStatusError, APIError) as e:
                if self._is_retryable(e) and idx < len(ordered) - 1:
                    last_error = e
                    continue
                raise
            choice = resp.choices[0]
            usage = resp.usage
            self._last_working_model = model
            return ModelResponse(
                content=choice.message.content or "",
                model=resp.model,
                input_tokens=usage.prompt_tokens if usage else 0,
                output_tokens=usage.completion_tokens if usage else 0,
                raw=resp.model_dump(),
            )
        raise last_error or RuntimeError(f"{self.name} 全部模型均调用失败")
