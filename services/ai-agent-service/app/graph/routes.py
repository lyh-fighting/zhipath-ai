"""条件边路由函数。"""
from __future__ import annotations

from .state import MAX_REPAIR_COUNT, ConsultationState


def attachment_check_route(state: ConsultationState) -> str:
    """有附件走 OCR，否则跳过。"""
    return "ocr" if state.get("attachments") else "skip"


def risk_route(state: ConsultationState) -> str:
    """高风险直接转人工，否则继续。"""
    return "high" if state.get("safety_risk_level") in ("high", "critical") else "low"


def clarify_route(state: ConsultationState) -> str:
    """缺关键字段时生成追问，否则进入检索。"""
    return "clarify" if state.get("need_clarification") else "proceed"


def dispatch_route(state: ConsultationState) -> str:
    """按意图分派 Agent。"""
    return state.get("intent", "unsupported")


def quality_route(state: ConsultationState) -> str:
    """质量检查路由：通过则最终化，可修复则修复，不安全转人工。

    修复循环最多 MAX_REPAIR_COUNT 次，超过转人工。
    """
    if state.get("need_human_handoff"):
        return "unsafe"
    if state.get("quality_score", 0) >= 60:
        return "pass"
    if state.get("repair_count", 0) >= MAX_REPAIR_COUNT:
        return "unsafe"
    return "repair"
