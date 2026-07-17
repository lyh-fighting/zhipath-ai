"""LangGraph 图构建。注册全部节点、条件边、修复循环（最多 MAX_REPAIR_COUNT 次）。"""
from __future__ import annotations

from langgraph.graph import END, START, StateGraph

from .nodes import (
    career_agent,
    clean_extracted_text,
    emotion_agent,
    finalize_response,
    generate_clarification,
    human_handoff,
    intent_router,
    load_context,
    mbti_completeness_check,
    multi_agent_collaboration,
    ocr_extract,
    persist_memory,
    profile_completeness_check,
    quality_check,
    repair_answer,
    retrieve_knowledge,
    risk_detector,
    unsupported_response,
)
from .routes import (
    attachment_check_route,
    clarify_route,
    dispatch_route,
    quality_route,
    risk_route,
)
from .state import ConsultationState


def build_graph():
    """编译并返回咨询工作流图。"""
    g = StateGraph(ConsultationState)

    nodes = {
        "ocr_extract": ocr_extract,
        "clean_extracted_text": clean_extracted_text,
        "load_context": load_context,
        "intent_router": intent_router,
        "risk_detector": risk_detector,
        "mbti_completeness_check": mbti_completeness_check,
        "profile_completeness_check": profile_completeness_check,
        "generate_clarification": generate_clarification,
        "retrieve_knowledge": retrieve_knowledge,
        "emotion_agent": emotion_agent,
        "career_agent": career_agent,
        "multi_agent_collaboration": multi_agent_collaboration,
        "unsupported_response": unsupported_response,
        "quality_check": quality_check,
        "repair_answer": repair_answer,
        "human_handoff": human_handoff,
        "finalize_response": finalize_response,
        "persist_memory": persist_memory,
    }
    for name, fn in nodes.items():
        g.add_node(name, fn)

    # 附件条件分支
    g.add_conditional_edges(
        START,
        attachment_check_route,
        {"ocr": "ocr_extract", "skip": "load_context"},
    )
    g.add_edge("ocr_extract", "clean_extracted_text")
    g.add_edge("clean_extracted_text", "load_context")

    # 意图 -> 风险
    g.add_edge("load_context", "intent_router")
    g.add_edge("intent_router", "risk_detector")
    g.add_conditional_edges(
        "risk_detector",
        risk_route,
        {"high": "human_handoff", "low": "mbti_completeness_check"},
    )

    # MBTI -> 画像 -> 澄清
    g.add_edge("mbti_completeness_check", "profile_completeness_check")
    g.add_conditional_edges(
        "profile_completeness_check",
        clarify_route,
        {"clarify": "generate_clarification", "proceed": "retrieve_knowledge"},
    )
    g.add_edge("generate_clarification", "persist_memory")

    # 检索 -> 分派
    g.add_conditional_edges(
        "retrieve_knowledge",
        dispatch_route,
        {
            "emotion": "emotion_agent",
            "career": "career_agent",
            "mixed": "multi_agent_collaboration",
            "unsupported": "unsupported_response",
        },
    )
    g.add_edge("emotion_agent", "quality_check")
    g.add_edge("career_agent", "quality_check")
    g.add_edge("multi_agent_collaboration", "quality_check")
    g.add_edge("unsupported_response", "finalize_response")

    # 质量检查 -> 通过/修复/转人工
    g.add_conditional_edges(
        "quality_check",
        quality_route,
        {"pass": "finalize_response", "repair": "repair_answer", "unsafe": "human_handoff"},
    )
    g.add_edge("repair_answer", "quality_check")  # 循环，quality_route 控制上限

    g.add_edge("human_handoff", "persist_memory")
    g.add_edge("finalize_response", "persist_memory")
    g.add_edge("persist_memory", END)

    return g.compile()
