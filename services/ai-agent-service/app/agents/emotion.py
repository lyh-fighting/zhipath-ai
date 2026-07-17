"""情感 Agent。禁止医学诊断、药物建议、绝对化判断。"""
from __future__ import annotations

from typing import Any

from ..prompts import EMOTION_PROMPT, EMOTION_PROMPT_VERSION
from ..providers import MockProvider, ModelMessage, ModelProvider
from ..schemas.analysis import AnalysisOutput


class EmotionAgent:
    def __init__(self, provider: ModelProvider | None = None):
        self.provider = provider or MockProvider()
        self.version = EMOTION_PROMPT_VERSION

    def analyze(self, message: str, mbti: str = "") -> AnalysisOutput:
        prompt = EMOTION_PROMPT.format(message=message, mbti=mbti or "未知")
        resp = self.provider.invoke(
            [
                ModelMessage(role="system", content="你是专业情感咨询师，禁止医学诊断与药物建议。"),
                ModelMessage(role="user", content=prompt),
            ]
        )
        return AnalysisOutput(
            problem_essence=resp.content[:200],
            reality="",
            mbti_reference=mbti,
            recommended_option=resp.content,
            prompt_version=self.version,
        )
