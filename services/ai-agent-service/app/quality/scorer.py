"""质量评分：事实依据 / 可执行性 / 安全 / 结构完整性，满分 100，>=60 通过。"""
from __future__ import annotations

from dataclasses import dataclass


@dataclass
class QualityScore:
    grounding: int = 0
    actionability: int = 0
    safety: int = 0
    structure: int = 0
    total: int = 0
    passed: bool = False


def score(draft_answer: str, structured_result: dict | None = None, safety_ok: bool = True) -> QualityScore:
    """计算质量评分。"""
    grounding = 20 if draft_answer else 0
    actionability = 20 if len(draft_answer) > 50 else 10
    safety = 25 if safety_ok else 0
    structure = 20 if structured_result else 10
    total = grounding + actionability + safety + structure
    return QualityScore(
        grounding=grounding,
        actionability=actionability,
        safety=safety,
        structure=structure,
        total=total,
        passed=total >= 60,
    )
