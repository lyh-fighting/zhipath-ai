"""Outbox 与通知测试。"""
from app.outbox import OutboxEvent, backoff_seconds, dispatch, is_dead
from app.jobs.notification import notify_report_done, schedule_followup, schedule_review


def test_backoff_exponential():
    assert backoff_seconds(0) == 1
    assert backoff_seconds(1) == 2
    assert backoff_seconds(2) == 4
    assert backoff_seconds(3) == 8


def test_backoff_capped_300():
    assert backoff_seconds(20) == 300


def test_is_dead_at_limit():
    assert not is_dead(0)
    assert not is_dead(4)
    assert is_dead(5)
    assert is_dead(10)


def test_dispatch_report_done():
    event = OutboxEvent(
        event_id="e1",
        event_type="report_done",
        aggregate_type="report",
        aggregate_id="r1",
        payload={"user_id": "u1"},
    )
    dispatch(event, {})


def test_dispatch_followup_trigger():
    event = OutboxEvent(
        event_id="e2",
        event_type="followup_trigger",
        aggregate_type="conversation",
        aggregate_id="c1",
        payload={"conversation_id": "c1", "task_type": "followup_3d"},
    )
    dispatch(event, {})


def test_schedule_followup_3d():
    task = schedule_followup({"conversation_id": "c1", "user_id": "u1"})
    assert task.task_type == "followup_3d"
    assert task.conversation_id == "c1"


def test_schedule_review_7d():
    task = schedule_review({"conversation_id": "c1", "user_id": "u1"})
    assert task.task_type == "review_7d"


def test_notify_report_done():
    result = notify_report_done({"user_id": "u1"})
    assert result["notified"] is True
