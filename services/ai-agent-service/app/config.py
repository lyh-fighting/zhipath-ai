"""ZhiPath AI Agent Service 配置。

关键凭证缺失时服务仍可启动（使用 Mock provider），但真实模型调用会失败。
"""
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    app_env: str = "local"

    # 模型 provider
    deepseek_api_key: str = ""
    openai_api_key: str = ""
    anthropic_api_key: str = ""
    primary_model: str = "deepseek:deepseek-chat"
    fallback_model: str = "openai:gpt-4o-mini"

    # 依赖
    redis_url: str = "redis://localhost:6379/0"
    qdrant_url: str = "http://localhost:6333"

    # 服务间鉴权
    internal_service_token: str = ""

    model_config = {"env_file": ".env", "extra": "ignore"}


settings = Settings()
