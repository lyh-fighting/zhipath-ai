"""ZhiPath OCR Service 入口。

OCR 只能用短期签名 URL 读取文件，不直接持有用户凭证。
"""
from fastapi import Depends, FastAPI, Header, HTTPException
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    internal_service_token: str = ""
    object_storage_endpoint: str = "localhost:9000"
    object_storage_access_key: str = "minioadmin"
    object_storage_secret_key: str = "minioadmin"

    model_config = {"env_file": ".env", "extra": "ignore"}


settings = Settings()
app = FastAPI(title="ZhiPath OCR Service", version="0.1.0")


async def verify_internal_token(x_internal_token: str = Header(None)) -> None:
    if not settings.internal_service_token or x_internal_token != settings.internal_service_token:
        raise HTTPException(status_code=401, detail="internal auth failed")


@app.get("/healthz")
async def healthz() -> dict:
    return {"status": "ok"}


@app.post("/internal/v1/ocr/extract", dependencies=[Depends(verify_internal_token)])
async def ocr_extract(payload: dict) -> dict:
    """OCR 识别。只能用短期签名 URL 读取文件。"""
    from .engine import MockOCREngine

    file_url = payload.get("file_url", "")
    file_type = payload.get("file_type", "image")
    engine = MockOCREngine()
    result = engine.extract(file_url, file_type)
    return {
        "code": "SUCCESS",
        "message": "ok",
        "data": {
            "clean_text": result.clean_text,
            "avg_confidence": result.avg_confidence,
            "blocks": result.blocks,
            "need_manual_review": result.need_manual_review,
            "ocr_status": result.ocr_status,
        },
        "trace_id": payload.get("trace_id", ""),
    }
