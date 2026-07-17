"""Agent 统一结构化输出。

情感/职业/决策 Agent 统一输出：问题本质 → 现实情况 → MBTI 参考 → 选项 → 风险 → 推荐 → 行动计划。
MBTI 是现实情况之后的主要分析依据。
"""
from __future__ import annotations

from dataclasses import dataclass, field


@dataclass
class AnalysisOutput:
    problem_essence: str = ""
    reality: str = ""
    mbti_reference: str = ""
    options: list[dict] = field(default_factory=list)
    risks: list[dict] = field(default_factory=list)
    recommended_option: str = ""
    action_plan: list[dict] = field(default_factory=list)
    prompt_version: str = ""
