"""风险分类器。"""
from __future__ import annotations

from dataclasses import dataclass

from .rules import RISK_KEYWORDS, RISK_LEVEL_MAP, RiskLevel, RiskType, detect

# 高危类型（自伤/自杀/家暴/未成年人）：一旦命中即视为 critical，阻断生成并给完整安全响应
_SEVERE_TYPES = {
    RiskType.SUICIDE,
    RiskType.SELF_HARM,
    RiskType.DOMESTIC_VIOLENCE,
    RiskType.MINOR,
}

# 风险等级排序，用于取最高等级
_LEVEL_RANK = {
    RiskLevel.NONE: 0,
    RiskLevel.LOW: 1,
    RiskLevel.MEDIUM: 2,
    RiskLevel.HIGH: 3,
    RiskLevel.CRITICAL: 4,
}


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


def classify_risk(text: str) -> dict:
    """扫描全部风险关键词，返回三级简化判定，供 main 链路快速拦截/标记。

    Returns:
        {
            "level": "critical" | "warning" | "ok",
            "matched": [{"type": str, "keyword": str, "level": str}, ...],
        }
    """
    matched: list[dict] = []
    for rtype, keywords in RISK_KEYWORDS.items():
        for kw in keywords:
            if kw and kw in text:
                matched.append(
                    {
                        "type": rtype.value,
                        "keyword": kw,
                        "level": RISK_LEVEL_MAP[rtype].value,
                    }
                )

    if not matched:
        return {"level": "ok", "matched": []}

    top_level = max(
        (RiskLevel(m["level"]) for m in matched),
        key=lambda lv: _LEVEL_RANK[lv],
    )
    has_severe = any(RiskType(m["type"]) in _SEVERE_TYPES for m in matched)

    # critical：自杀（CRITICAL）或任一高危类型（自伤/家暴/未成年人）
    #   → 阻断生成、返回完整安全响应（含紧急联系方式）
    if top_level == RiskLevel.CRITICAL or has_severe:
        return {"level": "critical", "matched": matched}
    # warning：其余已识别风险（如暴力 MEDIUM）→ 仍生成但标记人工介入
    if top_level in (RiskLevel.HIGH, RiskLevel.MEDIUM):
        return {"level": "warning", "matched": matched}
    return {"level": "ok", "matched": matched}
