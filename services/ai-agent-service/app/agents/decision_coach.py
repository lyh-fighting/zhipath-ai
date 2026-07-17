"""决策教练。混合问题并行执行情感+职业，由决策教练合并。"""
from __future__ import annotations

from ..prompts import DECISION_COACH_PROMPT_VERSION
from ..schemas.analysis import AnalysisOutput
from .career import CareerAgent
from .emotion import EmotionAgent


class DecisionCoachAgent:
    def __init__(self, emotion: EmotionAgent | None = None, career: CareerAgent | None = None):
        self.emotion = emotion or EmotionAgent()
        self.career = career or CareerAgent()
        self.version = DECISION_COACH_PROMPT_VERSION

    def analyze(self, message: str, mbti: str = "") -> AnalysisOutput:
        """并行执行情感与职业分析，合并为综合决策。"""
        e = self.emotion.analyze(message, mbti)
        c = self.career.analyze(message, mbti)
        return AnalysisOutput(
            problem_essence=f"综合：{e.problem_essence} | {c.problem_essence}",
            reality="",
            mbti_reference=mbti,
            options=e.options + c.options,
            risks=e.risks + c.risks,
            recommended_option=f"情感：{e.recommended_option}；职业：{c.recommended_option}",
            action_plan=e.action_plan + c.action_plan,
            prompt_version=self.version,
        )
