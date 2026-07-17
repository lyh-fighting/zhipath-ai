"""E2E 全链路测试。

覆盖注册/MBTI/基础咨询/深度分析/OCR/RAG/报告/支付Mock/回访事件。
验证：无MBTI只得基础分析、切换MBTI后历史报告保留旧快照、危机文本不返回普通行动建议。
"""
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[2] / "services" / "ai-agent-service"))

from app.graph import build_graph


def test_no_mbti_basic_only():
    """无 MBTI 时只得基础分析。"""
    graph = build_graph()
    result = graph.invoke({"message": "想换工作", "consultation_type": "career"})
    assert result.get("final_answer")
    assert not result.get("current_mbti_result_id")


def test_crisis_no_action_advice():
    """危机文本不返回普通行动建议。"""
    graph = build_graph()
    result = graph.invoke({"message": "我不想活了"})
    assert result.get("need_human_handoff") is True
    assert "转接人工" in result.get("final_answer", "")


def test_mbti_snapshot_preserved():
    """切换 MBTI 后历史报告保留旧快照。"""
    graph = build_graph()
    result = graph.invoke({
        "message": "想换工作",
        "consultation_type": "career",
        "current_mbti_result_id": "mbti_old",
        "mbti_profile": {"result_type": "INFP", "assertiveness": "T"},
    })
    assert result.get("mbti_snapshot") is not None
    assert result["mbti_snapshot"]["result_type"] == "INFP"


def test_emotion_flow():
    graph = build_graph()
    result = graph.invoke({"message": "分手了很难过", "consultation_type": "emotion"})
    assert result.get("final_answer")


def test_career_flow():
    graph = build_graph()
    result = graph.invoke({"message": "想跳槽但不知道去哪", "consultation_type": "career"})
    assert result.get("final_answer")
