"""幂等初始化 Qdrant collection 与 payload index。

由 docker-compose 的 qdrant-init 服务在启动时执行。
读取环境变量 QDRANT_URL 与 EMBEDDING_DIMENSION。
"""
import os
import json
import urllib.request
import urllib.error

QDRANT_URL = os.getenv("QDRANT_URL", "http://localhost:6333").rstrip("/")
DIM = int(os.getenv("EMBEDDING_DIMENSION", "1536"))

COLLECTIONS = {
    "zhipath_kb_v1": {
        "indexes": ["tenant_id", "domain", "category", "version", "risk_level", "doc_id", "chunk_id"],
    },
    "zhipath_user_memory_v1": {
        "indexes": ["tenant_id", "user_id", "domain", "memory_type", "mbti_type", "importance_score", "created_at"],
    },
    "zhipath_conversation_summary_v1": {
        "indexes": ["tenant_id", "user_id", "conversation_id", "domain", "summary_type", "created_at"],
    },
}


def req(method: str, path: str, body: dict | None = None):
    url = f"{QDRANT_URL}{path}"
    data = json.dumps(body).encode() if body else None
    r = urllib.request.Request(url, data=data, method=method)
    r.add_header("Content-Type", "application/json")
    try:
        with urllib.request.urlopen(r) as resp:
            return resp.status, json.loads(resp.read() or "null")
    except urllib.error.HTTPError as e:
        return e.code, json.loads(e.read() or "null")


def main() -> None:
    print(f"Qdrant: {QDRANT_URL}, embedding dim: {DIM}")
    for name, cfg in COLLECTIONS.items():
        status, _ = req("GET", f"/collections/{name}")
        if status == 200:
            print(f"  ✓ {name} 已存在，跳过创建")
        else:
            status, body = req(
                "PUT",
                f"/collections/{name}",
                {"vectors": {"size": DIM, "distance": "Cosine"}},
            )
            if status == 200:
                print(f"  ✓ 创建 {name}")
            else:
                print(f"  ✗ 创建 {name} 失败: {status} {body}")
                continue
        for field in cfg["indexes"]:
            req("PUT", f"/collections/{name}/index", {"field_name": field, "field_schema": "keyword"})
        print(f"    payload index: {', '.join(cfg['indexes'])}")
    print("✓ Qdrant collection 初始化完成")


if __name__ == "__main__":
    main()
