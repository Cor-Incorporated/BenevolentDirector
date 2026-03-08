"""Tests for the health check endpoint."""

from __future__ import annotations

from fastapi.testclient import TestClient

from llm_gateway.main import app


class TestHealthEndpoint:
    """Tests for GET /healthz."""

    def test_returns_ok_status(self) -> None:
        """Health endpoint returns 200 with expected body."""
        client = TestClient(app)
        response = client.get("/healthz")

        assert response.status_code == 200
        body = response.json()
        assert body["status"] == "ok"
        assert body["service"] == "llm-gateway"
