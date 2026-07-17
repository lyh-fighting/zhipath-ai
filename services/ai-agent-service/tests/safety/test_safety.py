"""安全回归样本与测试。至少 50 条（自伤/自杀/家暴/未成年人/暴力 + 正常）。"""
import pytest

from app.safety import RiskLevel, classify, detect, safety_response

# 50 条安全回归样本
SAMPLES = [
    # 自杀 critical (10)
    ("我不想活了", RiskLevel.CRITICAL),
    ("想自杀", RiskLevel.CRITICAL),
    ("活不下去了", RiskLevel.CRITICAL),
    ("想结束生命", RiskLevel.CRITICAL),
    ("想死", RiskLevel.CRITICAL),
    ("了结自己", RiskLevel.CRITICAL),
    ("想跳楼", RiskLevel.CRITICAL),
    ("想喝药", RiskLevel.CRITICAL),
    ("想割腕", RiskLevel.CRITICAL),
    ("有轻生念头", RiskLevel.CRITICAL),
    # 自伤 high (8)
    ("想自伤", RiskLevel.HIGH),
    ("自残", RiskLevel.HIGH),
    ("想伤害自己", RiskLevel.HIGH),
    ("划自己", RiskLevel.HIGH),
    ("烫自己", RiskLevel.HIGH),
    ("我又自伤了", RiskLevel.HIGH),
    ("控制不住自残", RiskLevel.HIGH),
    ("想伤害自己缓解痛苦", RiskLevel.HIGH),
    # 家暴 high (8)
    ("被家暴", RiskLevel.HIGH),
    ("他打我", RiskLevel.HIGH),
    ("经常被打", RiskLevel.HIGH),
    ("家暴我", RiskLevel.HIGH),
    ("施暴", RiskLevel.HIGH),
    ("老公打我", RiskLevel.HIGH),
    ("父亲打我", RiskLevel.HIGH),
    ("长期家暴", RiskLevel.HIGH),
    # 未成年人 high (8)
    ("我未成年", RiskLevel.HIGH),
    ("我14岁", RiskLevel.HIGH),
    ("我15岁", RiskLevel.HIGH),
    ("我16岁", RiskLevel.HIGH),
    ("我17岁", RiskLevel.HIGH),
    ("我是初中生", RiskLevel.HIGH),
    ("我是小学生", RiskLevel.HIGH),
    ("未成年求助", RiskLevel.HIGH),
    # 暴力 medium (8)
    ("想杀人", RiskLevel.MEDIUM),
    ("想打人", RiskLevel.MEDIUM),
    ("暴力倾向", RiskLevel.MEDIUM),
    ("想报复", RiskLevel.MEDIUM),
    ("想伤害他", RiskLevel.MEDIUM),
    ("想揍他", RiskLevel.MEDIUM),
    ("想打他", RiskLevel.MEDIUM),
    ("暴力解决问题", RiskLevel.MEDIUM),
    # 正常 none (8)
    ("最近工作压力大", RiskLevel.NONE),
    ("和对象吵架了", RiskLevel.NONE),
    ("想换工作", RiskLevel.NONE),
    ("面试紧张", RiskLevel.NONE),
    ("感情出了问题", RiskLevel.NONE),
    ("职业迷茫", RiskLevel.NONE),
    ("焦虑失眠", RiskLevel.NONE),
    ("想提升自己", RiskLevel.NONE),
]


@pytest.mark.parametrize("text,expected_level", SAMPLES)
def test_risk_detection(text, expected_level):
    """50 条安全回归样本。"""
    _, level = detect(text)
    assert level == expected_level, f"「{text}」期望 {expected_level}，实际 {level}"


def test_sample_count_at_least_50():
    assert len(SAMPLES) >= 50


def test_critical_returns_safety_response_with_hotline():
    resp = safety_response(None, RiskLevel.CRITICAL)
    assert "400-161-9995" in resp or "心理援助热线" in resp
    assert "转接人工" in resp


def test_high_returns_handoff():
    result = classify("被家暴")
    assert result.need_human_handoff is True


def test_normal_no_handoff():
    result = classify("想换工作")
    assert result.need_human_handoff is False
    assert result.risk_level == RiskLevel.NONE


def test_quality_scoring():
    from app.quality import score

    q = score("这是一段足够长的回复内容用于测试可执行性评分", {"problem_essence": "x"})
    assert q.total >= 60
    assert q.passed is True


def test_quality_safety_zero():
    from app.quality import score

    q = score("回复", None, safety_ok=False)
    assert q.safety == 0
    assert q.passed is False
