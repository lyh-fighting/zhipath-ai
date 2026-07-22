"""质量评分：安全合规 / 结构完整性 / 可执行性 / 长度合理，满分 100，>=60 视为通过。

自包含实现：仅依赖草稿文本、意图与历史，不依赖 RAG（retriever 当前返回空）。
"""
from __future__ import annotations

import re
from dataclasses import dataclass


@dataclass
class QualityScore:
    """质量评分明细（保留以便兼容导入；score() 直接返回 total 的 int）。"""

    safety: int = 0
    structure: int = 0
    actionability: int = 0
    length: int = 0
    total: int = 0


# 结构化表达信号：编号/项目符号或以下关键词之一
_STRUCTURE_KEYWORDS = [
    "问题本质",
    "现实情况",
    "选项",
    "风险",
    "推荐",
    "行动计划",
    "首先",
    "其次",
    "最后",
    "一方面",
    "另一方面",
    "总结",
]
# 可执行性信号：给出具体可操作建议
_ACTION_KEYWORDS = [
    "建议",
    "可以",
    "你可以",
    "不妨",
    "试试",
    "第一步",
    "步骤",
    "行动计划",
    "考虑",
    "推荐你",
    "不妨尝试",
    "先",
]
# 危险内容信号：草稿中出现则安全分项归零
_DANGER_KEYWORDS = [
    "自杀",
    "不想活",
    "想死",
    "结束生命",
    "了结自己",
    "轻生",
    "自伤",
    "自残",
    "伤害自己",
    "杀人",
    "打人",
    "报复",
]


def score(draft_answer: str, intent: dict | None = None, history: list | None = None) -> int:
    """计算质量评分，返回 0-100 的整数。

    Args:
        draft_answer: 模型生成的草稿回答。
        intent: 意图分析结果（dict），预留相关性/置信度微调。
        history: 对话历史（list[dict]），预留上下文一致性扩展。

    Returns:
        0-100 的整数质量分。
    """
    answer = (draft_answer or "").strip()
    length = len(answer)

    # 1) 安全合规（0-25）：草稿不得包含危险内容
    if not answer:
        safety = 0
    elif any(kw in answer for kw in _DANGER_KEYWORDS):
        safety = 0
    else:
        safety = 25

    # 是否出现编号/项目符号列表
    has_list = bool(
        re.search(r"(?m)^\s*(\d+[.、]|[-•·*]\s)", answer)
        or ("一、" in answer)
        or ("1." in answer)
        or ("1、" in answer)
    )
    has_structure_kw = any(kw in answer for kw in _STRUCTURE_KEYWORDS)

    # 2) 结构完整性（0-25）
    if not answer:
        structure = 0
    elif has_list and has_structure_kw:
        structure = 25
    elif has_list or has_structure_kw:
        structure = 18
    elif length >= 80:
        structure = 12
    else:
        structure = 6

    # 3) 可执行性（0-25）：给出具体可操作建议
    if not answer:
        actionability = 0
    else:
        actionability = 20 if any(kw in answer for kw in _ACTION_KEYWORDS) else 8
        if has_list:
            actionability = min(25, actionability + 5)

    # 4) 长度合理（0-25）
    if length == 0:
        length_score = 0
    elif 120 <= length <= 1500:
        length_score = 25
    elif 60 <= length < 120 or 1500 < length <= 3000:
        length_score = 15
    else:
        length_score = 5

    total = safety + structure + actionability + length_score
    # 兜底：非空回答至少给一个基础分，避免误判为 0
    if answer and total < 10:
        total = 10
    return max(0, min(100, total))
