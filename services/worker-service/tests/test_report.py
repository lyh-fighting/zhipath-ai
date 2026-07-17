"""报告生成测试。"""
from app.jobs.report import ReportJob, can_retry, generate_report


def test_can_retry():
    assert can_retry("pending")
    assert can_retry("failed")
    assert not can_retry("completed")


def test_generate_report():
    job = ReportJob(report_id="r1", user_id="u1", conversation_id="c1", mbti_result_id="m1")
    result = generate_report(job, {"result_type": "INFP"})
    assert result["report_id"] == "r1"
    assert result["status"] == "completed"
    assert result["mbti_snapshot"]["result_type"] == "INFP"
