"""ConsultationState：LangGraph 咨询工作流状态。

对应架构文档 3.1 节。深度分析必须校验 current_mbti_result_id，每次执行生成不可变 mbti_snapshot。
"""
from __future__ import annotations

from typing import Any, Literal, TypedDict


class ConsultationState(TypedDict, total=False):
    # ===== 请求上下文 =====
    trace_id: str
    user_id: str
    conversation_id: str
    message: str
    client_type: Literal["wechat_miniapp", "ios", "android", "web"]
    consultation_type: Literal["auto", "emotion", "career"]

    # ===== 附件与 OCR =====
    attachments: list[dict[str, Any]]
    extracted_text: str
    ocr_status: Literal["none", "pending", "completed", "failed"]

    # ===== 用户画像与 MBTI =====
    user_profile: dict[str, Any]
    mbti_profile: dict[str, Any]
    mbti_missing: bool
    should_prompt_mbti: bool
    current_mbti_result_id: str
    mbti_snapshot: dict[str, Any]  # 每次执行不可变快照
    short_term_memory: list[dict[str, Any]]
    long_term_memory: list[dict[str, Any]]

    # ===== 意图与澄清 =====
    intent: Literal["emotion", "career", "mixed", "crisis", "unsupported"]
    intent_confidence: float
    missing_fields: list[str]
    need_clarification: bool
    clarification_questions: list[str]

    # ===== 检索与 Agent =====
    retrieved_docs: list[dict[str, Any]]
    agent_outputs: dict[str, Any]
    draft_answer: str
    structured_result: dict[str, Any]

    # ===== 质量与安全 =====
    quality_score: int
    repair_count: int  # 修复循环计数，最多 2 次
    safety_risk_level: Literal["none", "low", "medium", "high", "critical"]
    need_human_handoff: bool

    # ===== 最终输出 =====
    final_answer: str
    next_actions: list[dict[str, Any]]


# 修复循环上限
MAX_REPAIR_COUNT = 2
