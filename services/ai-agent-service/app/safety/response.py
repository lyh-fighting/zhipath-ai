"""安全响应。critical 风险不进普通生成链，返回安全响应并转人工。"""
from __future__ import annotations

import json
from pathlib import Path

from .rules import RiskLevel, RiskType

_EMERGENCY: dict | None = None


def _load_emergency() -> dict:
    global _EMERGENCY
    if _EMERGENCY is None:
        # infra/config/emergency_resources.zh-CN.json
        p = Path(__file__).resolve().parents[4] / "infra" / "config" / "emergency_resources.zh-CN.json"
        _EMERGENCY = json.loads(p.read_text(encoding="utf-8")) if p.exists() else {}
    return _EMERGENCY


def safety_response(risk_type: RiskType | None, level: RiskLevel) -> str:
    """critical/high 风险返回安全响应，含地区化紧急联系方式。"""
    if level == RiskLevel.CRITICAL:
        res = _load_emergency().get("crisis_lines", [])
        lines = "\n".join(f"· {r['name']}：{r['number']}（{r['hours']}）" for r in res) or "· 心理援助热线：400-161-9995（24 小时）"
        return (
            "我注意到你提到了一些让我非常担心的内容。你的安全是最重要的。\n"
            "如果你正处于危险中或有过伤害自己的想法，请立即联系：\n"
            f"{lines}\n"
            "· 紧急情况请拨 110 或 120\n"
            "我已为你转接人工咨询师，请等待。"
        )
    if level == RiskLevel.HIGH:
        return "检测到你可能需要更专业的帮助，正在为你转接人工咨询师。"
    return ""
