"""Mock provider 测试。"""
from app.providers import MockProvider, ModelMessage


def test_mock_provider_deterministic():
    provider = MockProvider()
    msgs = [ModelMessage(role="user", content="我想换工作")]
    r1 = provider.invoke(msgs)
    r2 = provider.invoke(msgs)
    assert r1.content == r2.content  # 确定性
    assert "换工作" in r1.content
    assert r1.model == "mock-1"
    assert r1.input_tokens > 0


def test_mock_provider_empty_messages():
    provider = MockProvider()
    r = provider.invoke([])
    assert r.content  # 即使无 user 消息也返回确定性结果
