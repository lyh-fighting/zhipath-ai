"""ZhiPath Worker 入口：轮询 outbox_events 并分派到各 handler。

用法：python -m app.worker
依赖环境变量：
  DATABASE_URL   go-sql-driver DSN，如 consult:consult@tcp(mysql:3306)/zhipath?parseTime=true&loc=Local
  QDRANT_URL    http://qdrant:6333（可选，缺省则 qdrant handler 降级跳过）
  WORKER_POLL_INTERVAL  轮询间隔秒，默认 5
"""
from __future__ import annotations

import json
import logging
import os
import re
import signal
import sys
import time
from typing import Any

import pymysql

try:
    from qdrant_client import QdrantClient
except ImportError:  # 未安装时降级
    QdrantClient = None  # type: ignore

from app.outbox import OutboxEvent, backoff_seconds, dispatch, is_dead

logging.basicConfig(
    level=getattr(logging, os.getenv("LOG_LEVEL", "info").upper(), logging.INFO),
    format="%(asctime)s [worker] %(levelname)s %(message)s",
)
log = logging.getLogger("worker")

DSN_RE = re.compile(
    r"(?P<user>[^:]+):(?P<pass>[^@]*)@tcp\((?P<host>[^:]+):(?P<port>\d+)\)/(?P<db>[^?]+)"
)

_BATCH = int(os.getenv("WORKER_BATCH_SIZE", "50"))
_INTERVAL = int(os.getenv("WORKER_POLL_INTERVAL", "5"))
_stop = False


def _parse_dsn(dsn: str) -> dict:
    m = DSN_RE.search(dsn or "")
    if not m:
        raise RuntimeError(f"无法解析 DATABASE_URL: {dsn!r}")
    return m.groupdict()


def _build_qdrant() -> Any:
    url = os.getenv("QDRANT_URL")
    if not url:
        log.warning("QDRANT_URL 未设置，qdrant 类 handler 将降级跳过")
        return None
    if QdrantClient is None:
        log.warning("qdrant_client 未安装，qdrant 类 handler 将降级跳过")
        return None
    return QdrantClient(url=url, timeout=10)


def _connect(dsn: str) -> pymysql.connections.Connection:
    cfg = _parse_dsn(dsn)
    return pymysql.connect(
        host=cfg["host"],
        port=int(cfg["port"]),
        user=cfg["user"],
        password=cfg["pass"],
        database=cfg["db"],
        charset="utf8mb4",
        autocommit=False,
    )


def _to_event(row: dict) -> OutboxEvent:
    return OutboxEvent(
        event_id=row["event_id"],
        event_type=row["event_type"],
        aggregate_type=row["aggregate_type"],
        aggregate_id=row["aggregate_id"],
        payload=json.loads(row["payload"]) if isinstance(row["payload"], str) else row["payload"],
        status=row["status"],
        retry_count=int(row["retry_count"] or 0),
        max_retries=int(row["max_retries"] or 5),
    )


def _poll_once(conn: pymysql.connections.Connection, deps: dict) -> int:
    """拉取一批 pending 事件并处理，返回处理数量。"""
    with conn.cursor(pymysql.cursors.DictCursor) as cur:
        cur.execute(
            """
            SELECT event_id, event_type, aggregate_type, aggregate_id,
                   payload, status, retry_count, max_retries
            FROM outbox_events
            WHERE status = 'pending'
              AND (next_attempt_at IS NULL OR next_attempt_at <= NOW())
            ORDER BY created_at ASC
            LIMIT %s FOR UPDATE SKIP LOCKED
            """,
            (_BATCH,),
        )
        rows = cur.fetchall()
        for row in rows:
            ev = _to_event(row)
            try:
                dispatch(ev, deps)
                cur.execute(
                    "UPDATE outbox_events SET status='processed', processed_at=NOW() WHERE event_id=%s",
                    (ev.event_id,),
                )
                conn.commit()
            except Exception as exc:  # 单事件失败隔离，不拖垮循环
                conn.rollback()
                rc = ev.retry_count + 1
                err = str(exc)[:2000]
                if is_dead(rc, ev.max_retries):
                    log.error("event %s 达重试上限，置 dead: %s", ev.event_id, exc)
                    cur.execute(
                        "UPDATE outbox_events SET status='dead', retry_count=%s, last_error=%s WHERE event_id=%s",
                        (rc, err, ev.event_id),
                    )
                else:
                    nxt = backoff_seconds(rc)
                    log.warning("event %s 处理失败(retry %d, %ds 后重试): %s", ev.event_id, rc, nxt, exc)
                    cur.execute(
                        "UPDATE outbox_events SET retry_count=%s, next_attempt_at=DATE_ADD(NOW(), INTERVAL %s SECOND), last_error=%s WHERE event_id=%s",
                        (rc, nxt, err, ev.event_id),
                    )
                conn.commit()
        return len(rows)


def main() -> int:
    dsn = os.getenv("DATABASE_URL")
    if not dsn:
        log.error("DATABASE_URL 未设置，worker 无法启动")
        return 1

    # 优雅退出
    def _on_signal(signum, frame):
        global _stop
        log.info("收到信号 %s，准备退出", signum)
        _stop = True

    signal.signal(signal.SIGINT, _on_signal)
    signal.signal(signal.SIGTERM, _on_signal)

    conn = _connect(dsn)
    deps = {
        "qdrant": _build_qdrant(),
        "object_store": None,  # 待接入 minio client（pyproject 暂未包含）
    }
    log.info("ZhiPath worker 已启动（轮询间隔 %ds，批次 %d）", _INTERVAL, _BATCH)

    idle = 0
    try:
        while not _stop:
            try:
                n = _poll_once(conn, deps)
            except pymysql.MySQLError as exc:
                log.error("轮询数据库异常，5s 后重试: %s", exc)
                conn.close()
                time.sleep(5)
                conn = _connect(dsn)
                continue
            if n == 0:
                idle += 1
                if idle % 12 == 0:  # 每分钟打一次心跳
                    log.debug("worker 空闲心跳（已连续 %ds 无事件）", idle * _INTERVAL)
                time.sleep(_INTERVAL)
            else:
                idle = 0
                log.info("本轮处理 %d 个 outbox 事件", n)
    finally:
        try:
            conn.close()
        except Exception:
            pass
    log.info("worker 已退出")
    return 0


if __name__ == "__main__":
    sys.exit(main())
