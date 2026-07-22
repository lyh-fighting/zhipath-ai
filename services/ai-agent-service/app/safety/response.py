"""安全响应。critical 风险不进普通生成链，返回安全响应并转人工。"""
from __future__ import annotations

import json
from pathlib import Path

from .rules import RiskLevel, RiskType

# 内嵌兜底紧急资源：容器镜像内不保证挂载 infra/config，故内置一份，
# 保证安全响应在任何环境都可用（地区化联系方式优先取自同级 JSON）。
_DEFAULT_EMERGENCY: dict = {
    "crisis_lines": [
        {"name": "心理援助热线", "number": "400-161-9995", "hours": "24 小时"},
        {"name": "北京心理危机研究与干预中心", "number": "010-82951332", "hours": "24 小时"},
        {"name": "希望24热线", "number": "400-161-9995", "hours": "24 小时"},
        {"name": "紧急求助", "number": "110", "hours": "24 小时"},
        {"name": "医疗急救", "number": "120", "hours": "24 小时"},
    ],
    "regions": {
        "北京": {"crisis": "010-82951332"},
        "上海": {"crisis": "021-12320-5"},
        "广州": {"crisis": "020-81899120"},
        "成都": {"crisis": "028-87577510"},
    },
}

_EMERGENCY: dict | None = None


def _load_emergency() -> dict:
    """加载紧急资源。本地开发若挂载了 infra/config 则优先使用；否则回退内嵌兜底。"""
    global _EMERGENCY
    if _EMERGENCY is not None:
        return _EMERGENCY
    candidates = [
        # 本地开发路径（services/ai-agent-service/app/safety -> 上溯 3 级到仓库根）
        Path(__file__).resolve().parents[3] / "infra" / "config" / "emergency_resources.zh-CN.json",
        # 容器内若通过挂载提供
        Path("/app/infra/config/emergency_resources.zh-CN.json"),
    ]
    data: dict = {}
    for p in candidates:
        try:
            if p.exists():
                data = json.loads(p.read_text(encoding="utf-8"))
                break
        except Exception:
            continue
    _EMERGENCY = data or _DEFAULT_EMERGENCY
    return _EMERGENCY


def safety_response(risk_type: RiskType | None, level: RiskLevel) -> str:
    """critical/high 风险返回安全响应，含紧急联系方式。"""
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
