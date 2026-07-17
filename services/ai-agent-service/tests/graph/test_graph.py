"""图构建与端到端测试。"""
from app.graph import build_graph


def test_graph_compiles():
    graph = build_graph()
    assert graph is not None


def test_emotion_flow():
    graph = build_graph()
    result = graph.invoke({"message": "最近很难过，分手了很焦虑", "consultation_type": "emotion"})
    assert result.get("final_answer")
    assert result.get("intent") == "emotion"


def test_crisis_flow_human_handoff():
    """危机文本必须转人工，不返回普通行动建议。"""
    graph = build_graph()
    result = graph.invoke({"message": "我不想活了"})
    assert result.get("need_human_handoff") is True


def test_career_flow():
    graph = build_graph()
    result = graph.invoke({"message": "想换工作但不知道怎么跳槽"})
    assert result.get("final_answer")
    assert result.get("intent") == "career"


def test_mixed_flow():
    graph = build_graph()
    result = graph.invoke({"message": "工作压力大，很焦虑，想跳槽"})
    assert result.get("intent") == "mixed"
    assert result.get("final_answer")


def test_mbti_snapshot_on_finalize():
    """有 current_mbti_result_id 时 finalize 生成 mbti_snapshot。"""
    graph = build_graph()
    result = graph.invoke(
        {
            "message": "想换工作",
            "consultation_type": "career",
            "current_mbti_result_id": "mbti_001",
            "mbti_profile": {"result_type": "INFP", "assertiveness": "T"},
        }
    )
    assert result.get("mbti_snapshot") is not None
    assert result["mbti_snapshot"]["result_type"] == "INFP"
