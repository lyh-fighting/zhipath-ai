"""Redis 短期记忆（当前会话上下文）。"""
from __future__ import annotations

import redis.asyncio as redis


class CacheStore:
    """当前对话上下文、限流、幂等键。"""

    def __init__(self, url: str):
        self._rdb = redis.from_url(url)

    async def get(self, key: str) -> str | None:
        return await self._rdb.get(key)

    async def set(self, key: str, value: str, ttl: int = 3600) -> None:
        await self._rdb.set(key, value, ex=ttl)

    async def ping(self) -> bool:
        try:
            return await self._rdb.ping()
        except Exception:
            return False
