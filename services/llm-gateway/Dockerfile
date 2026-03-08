FROM python:3.12-slim AS base

WORKDIR /app

COPY pyproject.toml .
RUN pip install --no-cache-dir .

COPY src/ src/
RUN pip install --no-cache-dir .

EXPOSE 8081

CMD ["uvicorn", "llm_gateway.main:app", "--host", "0.0.0.0", "--port", "8081"]
