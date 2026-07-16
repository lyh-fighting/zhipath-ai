# ZhiPath Worker Service (Python, Outbox poller + 异步任务)
# Task 12/15 提供 pyproject.toml 与 app 代码后可正常 build
FROM python:3.12-slim AS base
ENV PYTHONUNBUFFERED=1 PIP_NO_CACHE_DIR=1
WORKDIR /app
COPY pyproject.toml uv.lock* ./
RUN pip install --no-cache-dir uv \
    && (uv sync --frozen 2>/dev/null || uv sync || pip install --no-cache-dir -e .)
COPY app ./app
CMD ["uv", "run", "python", "-m", "app.worker"]
