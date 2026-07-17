"""职业 Agent。禁止保证 offer、薪资、晋升。"""
from __future__ import annotations

from ..prompts import CAREER_PROMPT, CAREER_PROMPT_VERSION
from ..providers import MockProvider, ModelMessage, ModelProvider
from ..schemas.analysis import AnalysisOutput


class CareerAgent:
    def __init__(self, provider: ModelProvider | None = None):
        self.provider = provider or MockProvider()
        self.version = CAREER_PROMPT_VERSION

    def analyze(self, message: str, mbti: str = "") -> AnalysisOutput:
        prompt = CAREER_PROMPT.format(message=message, mbti=mbti or "未知")
        resp = self.provider.invoke(
            [
                ModelMessage(role="system", content="你是职业规划师，禁止保证 offer/薪资/晋升。"),
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
