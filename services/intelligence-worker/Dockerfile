FROM python:3.12-slim AS base

WORKDIR /app

COPY pyproject.toml .
RUN pip install --no-cache-dir .

COPY src/ src/
RUN pip install --no-cache-dir .

CMD ["python", "-m", "intelligence_worker.main"]
