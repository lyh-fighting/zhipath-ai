"""路由函数测试。"""
from app.graph.routes import (
    attachment_check_route,
    clarify_route,
    dispatch_route,
    quality_route,
    risk_route,
)


def test_attachment_with_attachments():
    assert attachment_check_route({"attachments": [{"file_id": "f1"}]}) == "ocr"


def test_attachment_without_attachments():
    assert attachment_check_route({}) == "skip"


def test_risk_high():
    assert risk_route({"safety_risk_level": "critical"}) == "high"
    assert risk_route({"safety_risk_level": "high"}) == "high"


def test_risk_low():
    assert risk_route({"safety_risk_level": "none"}) == "low"


def test_clarify():
    assert clarify_route({"need_clarification": True}) == "clarify"
    assert clarify_route({"need_clarification": False}) == "proceed"


def test_dispatch():
    assert dispatch_route({"intent": "emotion"}) == "emotion"
    assert dispatch_route({"intent": "career"}) == "career"
    assert dispatch_route({}) == "unsupported"


def test_quality_pass():
    assert quality_route({"quality_score": 70}) == "pass"


def test_quality_repair():
    assert quality_route({"quality_score": 30, "repair_count": 0}) == "repair"


def test_quality_over_limit_unsafe():
    assert quality_route({"quality_score": 30, "repair_count": 2}) == "unsafe"


def test_quality_human_handoff_unsafe():
    assert quality_route({"need_human_handoff": True, "quality_score": 90}) == "unsafe"
