"""Agent 测试。"""
from app.agents import CareerAgent, DecisionCoachAgent, EmotionAgent


def test_emotion_agent():
    agent = EmotionAgent()
    out = agent.analyze("最近很难过", "INFP")
    assert out.prompt_version == "emotion-v1"
    assert out.mbti_reference == "INFP"
    assert out.problem_essence


def test_career_agent():
    agent = CareerAgent()
    out = agent.analyze("想换工作", "ENTJ")
    assert out.prompt_version == "career-v1"
    assert out.mbti_reference == "ENTJ"


def test_decision_coach_merges_emotion_and_career():
    coach = DecisionCoachAgent()
    out = coach.analyze("工作压力很大想跳槽", "INFP")
    assert out.prompt_version == "decision-coach-v1"
    assert "综合" in out.problem_essence
    assert "情感" in out.recommended_option
    assert "职业" in out.recommended_option


def test_emotion_output_structure():
    agent = EmotionAgent()
    out = agent.analyze("焦虑", "")
    assert hasattr(out, "problem_essence")
    assert hasattr(out, "reality")
    assert hasattr(out, "mbti_reference")
    assert hasattr(out, "prompt_version")
