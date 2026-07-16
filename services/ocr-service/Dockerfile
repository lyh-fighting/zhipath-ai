# ZhiPath OCR Service (Python FastAPI + PaddleOCR)
# Task 14 提供 pyproject.toml 与 app 代码后可正常 build
FROM python:3.12-slim AS base
ENV PYTHONUNBUFFERED=1 PIP_NO_CACHE_DIR=1
WORKDIR /app
COPY pyproject.toml uv.lock* ./
RUN pip install --no-cache-dir uv \
    && (uv sync --frozen 2>/dev/null || uv sync || pip install --no-cache-dir -e .)
COPY app ./app
EXPOSE 8002
HEALTHCHECK --interval=10s --timeout=5s --retries=10 \
  CMD python -c "import urllib.request;urllib.request.urlopen('http://localhost:8002/healthz')" || exit 1
CMD ["uv", "run", "uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8002"]
