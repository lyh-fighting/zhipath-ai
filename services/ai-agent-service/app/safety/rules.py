"""安全规则：风险类型、等级与关键词检测。

检测自伤、自杀、家暴、未成年人、暴力威胁。
critical 风险不进普通生成链，返回安全响应并创建人工介入事件。
"""
from __future__ import annotations

from enum import Enum


class RiskType(str, Enum):
    SELF_HARM = "self_harm"
    SUICIDE = "suicide"
    DOMESTIC_VIOLENCE = "domestic_violence"
    MINOR = "minor"
    VIOLENCE = "violence"


class RiskLevel(str, Enum):
    NONE = "none"
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


# 风险关键词（中文）
RISK_KEYWORDS: dict[RiskType, list[str]] = {
    RiskType.SUICIDE: ["自杀", "不想活", "活不下去", "想死", "结束生命", "了结自己", "跳楼", "喝药", "割腕", "轻生"],
    RiskType.SELF_HARM: ["自伤", "自残", "伤害自己", "划自己", "烫自己"],
    RiskType.DOMESTIC_VIOLENCE: ["家暴", "打我", "被打", "家暴我", "施暴"],
    RiskType.MINOR: ["未成年", "14岁", "15岁", "16岁", "17岁", "初中生", "小学生"],
    RiskType.VIOLENCE: ["杀人", "打人", "打他", "暴力", "报复", "伤害他", "揍他"],
}

# 风险类型 -> 等级
RISK_LEVEL_MAP: dict[RiskType, RiskLevel] = {
    RiskType.SUICIDE: RiskLevel.CRITICAL,
    RiskType.SELF_HARM: RiskLevel.HIGH,
    RiskType.DOMESTIC_VIOLENCE: RiskLevel.HIGH,
    RiskType.MINOR: RiskLevel.HIGH,
    RiskType.VIOLENCE: RiskLevel.MEDIUM,
}


def detect(text: str) -> tuple[RiskType | None, RiskLevel]:
    """检测文本中的风险。返回 (风险类型, 风险等级)。"""
    for rtype, keywords in RISK_KEYWORDS.items():
        for kw in keywords:
            if kw in text:
                return rtype, RISK_LEVEL_MAP[rtype]
    return None, RiskLevel.NONE
