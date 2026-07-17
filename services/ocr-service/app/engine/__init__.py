"""OCR 引擎接口与实现。"""
from __future__ import annotations

from dataclasses import dataclass, field
from typing import Protocol


@dataclass
class OCRResult:
    clean_text: str = ""
    avg_confidence: float = 0.0
    blocks: list[dict] = field(default_factory=list)
    need_manual_review: bool = False
    ocr_status: str = "completed"


class OCREngine(Protocol):
    name: str

    def extract(self, file_url: str, file_type: str = "image") -> OCRResult: ...


class MockOCREngine:
    """Mock OCR，返回确定性结果。低置信度结果必须要求用户确认。"""

    name = "mock"

    def extract(self, file_url: str, file_type: str = "image") -> OCRResult:
        text = f"mock OCR result for {file_url[-20:]}"
        return OCRResult(
            clean_text=text,
            avg_confidence=0.95,
            blocks=[{"type": "text", "text": text, "confidence": 0.95}],
        )


__all__ = ["OCREngine", "OCRResult", "MockOCREngine"]
