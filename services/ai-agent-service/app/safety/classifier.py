"""风险分类器。"""
from __future__ import annotations

from dataclasses import dataclass

from .rules import RiskLevel, RiskType, detect


@dataclass
class RiskResult:
    risk_type: RiskType | None
    risk_level: RiskLevel
    need_human_handoff: bool
    matched_keyword: str = ""


def classify(text: str) -> RiskResult:
    """分类文本风险。HIGH/CRITICAL 需人工介入。"""
    rtype, level = detect(text)
    need_handoff = level in (RiskLevel.HIGH, RiskLevel.CRITICAL)
    return RiskResult(risk_type=rtype, risk_level=level, need_human_handoff=need_handoff)
