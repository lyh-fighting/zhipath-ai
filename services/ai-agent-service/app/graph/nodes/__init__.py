"""LangGraph 节点实现。

每个节点接收 state，返回需更新的字段 dict。
关键节点（intent_router/risk/mbti/quality/finalize）含实际逻辑，其余为骨架（Task 11/12/13/14 完善）。
"""
from __future__ import annotations

from typing import Any

from ..state import ConsultationState

# ===== 危机/意图关键词 =====
_CRISIS_KW = ["不想活", "自杀", "自伤", "家暴", "打死", "活不下去"]
_CAREER_KW = ["工作", "职业", "跳槽", "换工作", "面试", "薪资", "晋升", "转行", "offer"]
_EMOTION_KW = ["难过", "焦虑", "分手", "吵架", "孤独", "迷茫", "压力", "崩溃", "抑郁"]


def ocr_extract(state: ConsultationState) -> dict[str, Any]:
    # TODO Task 14: 调用 OCR Service
    return {"extracted_text": "", "ocr_status": "completed"}


def clean_extracted_text(state: ConsultationState) -> dict[str, Any]:
    return {"extracted_text": state.get("extracted_text", "").strip()}


def load_context(state: ConsultationState) -> dict[str, Any]:
    # TODO: 读用户画像、最近对话、长期记忆、会员权益
    return {"user_profile": {}, "short_term_memory": [], "long_term_memory": []}


def intent_router(state: ConsultationState) -> dict[str, Any]:
    """意图识别。关键词规则，Task 11 接模型。"""
    msg = state.get("message", "")
    if any(k in msg for k in _CRISIS_KW):
        return {"intent": "crisis", "intent_confidence": 0.9, "safety_risk_level": "critical"}
    is_career = any(k in msg for k in _CAREER_KW)
    is_emotion = any(k in msg for k in _EMOTION_KW)
    if is_career and is_emotion:
        return {"intent": "mixed", "intent_confidence": 0.7}
    if is_career:
        return {"intent": "career", "intent_confidence": 0.8}
    if is_emotion:
        return {"intent": "emotion", "intent_confidence": 0.8}
    return {"intent": "unsupported", "intent_confidence": 0.3}


def risk_detector(state: ConsultationState) -> dict[str, Any]:
    """风险检测。intent_router 已初检，此处补充并标记人工介入。"""
    if state.get("safety_risk_level") in ("high", "critical"):
        return {"need_human_handoff": True}
    return {}


def mbti_completeness_check(state: ConsultationState) -> dict[str, Any]:
    """深度分析必须校验 current_mbti_result_id。缺失时提示测试。"""
    mbti = state.get("mbti_profile") or {}
    missing = not mbti.get("result_type")
    should_prompt = missing and state.get("consultation_type") in ("emotion", "career")
    return {"mbti_missing": missing, "should_prompt_mbti": should_prompt}


def profile_completeness_check(state: ConsultationState) -> dict[str, Any]:
    return {"need_clarification": False}


def generate_clarification(state: ConsultationState) -> dict[str, Any]:
    return {"clarification_questions": []}


def retrieve_knowledge(state: ConsultationState) -> dict[str, Any]:
    # TODO Task 13: 按 domain 检索 Qdrant
    return {"retrieved_docs": []}


def emotion_agent(state: ConsultationState) -> dict[str, Any]:
    # TODO Task 11: 调模型 + 禁医学诊断
    return {
        "draft_answer": f"[emotion mock] 我理解你的感受：{state.get('message', '')[:50]}",
        "agent_outputs": {"emotion": True},
    }


def career_agent(state: ConsultationState) -> dict[str, Any]:
    # TODO Task 11: 调模型 + 禁保证 offer/薪资/晋升
    return {
        "draft_answer": f"[career mock] 职业建议：{state.get('message', '')[:50]}",
        "agent_outputs": {"career": True},
    }


def multi_agent_collaboration(state: ConsultationState) -> dict[str, Any]:
    e = emotion_agent(state)
    c = career_agent(state)
    return {"draft_answer": f"{e['draft_answer']}\n{c['draft_answer']}", "agent_outputs": {"mixed": True}}


def unsupported_response(state: ConsultationState) -> dict[str, Any]:
    return {"draft_answer": "暂时无法识别你的问题类型，请补充更多细节。"}


def quality_check(state: ConsultationState) -> dict[str, Any]:
    # TODO Task 12: 事实核查/安全合规/可执行性评分
    score = 70 if state.get("draft_answer") else 0
    return {"quality_score": score}


def repair_answer(state: ConsultationState) -> dict[str, Any]:
    count = state.get("repair_count", 0) + 1
    return {"repair_count": count, "draft_answer": state.get("draft_answer", "") + "（已修复）"}


def human_handoff(state: ConsultationState) -> dict[str, Any]:
    return {
        "need_human_handoff": True,
        "final_answer": "检测到你需要更专业的帮助，正在为你转接人工。",
    }


def finalize_response(state: ConsultationState) -> dict[str, Any]:
    """生成最终结构化响应。深度分析生成不可变 mbti_snapshot。"""
    answer = state.get("draft_answer", "")
    snapshot = None
    if state.get("current_mbti_result_id"):
        snapshot = state.get("mbti_profile")
    return {
        "final_answer": answer,
        "structured_result": {
            "problem_essence": "",
            "options": [],
            "recommended_plan": [],
            "next_actions": [],
        },
        "mbti_snapshot": snapshot,
    }


def persist_memory(state: ConsultationState) -> dict[str, Any]:
    # TODO: 写消息/摘要/画像变化/decision_records/agent_runs
    # 通过 Outbox 同步 Qdrant，不直接双写
    return {}
