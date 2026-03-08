"""Health check endpoint."""

from __future__ import annotations

from fastapi import APIRouter

router = APIRouter()


@router.get("/healthz")
async def healthz() -> dict[str, str]:
    """Return service health status.

    Returns:
        A dict with status and service name.
    """
    return {"status": "ok", "service": "llm-gateway"}
