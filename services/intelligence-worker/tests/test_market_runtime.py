"""Tests for market runtime wiring."""

from __future__ import annotations

import json
from dataclasses import dataclass, field
from typing import Any
from unittest.mock import MagicMock

from intelligence_worker.market.models import (
    AggregatedEvidence,
    EvidenceFragment,
    Range,
)
from intelligence_worker.market.runtime import (
    MarketResearchRequestedHandler,
    PostgresMarketEvidenceRepository,
    build_market_providers,
)
from intelligence_worker.market.subscriber import MarketResearchRequestedSubscriber


@dataclass
class _FakeMessage:
    payload: dict[str, object]
    acked: bool = False
    nacked: bool = False

    @property
    def data(self) -> bytes:
        return json.dumps(self.payload).encode("utf-8")

    def ack(self) -> None:
        self.acked = True

    def nack(self) -> None:
        self.nacked = True


@dataclass
class _FakeFuture:
    canceled: bool = False

    def cancel(self) -> None:
        self.canceled = True

    def result(self) -> None:
        return None


@dataclass
class _FakeClient:
    subscription: str | None = None
    callback: Any = None
    future: _FakeFuture = field(default_factory=_FakeFuture)

    def subscribe(self, subscription: str, callback: Any) -> _FakeFuture:
        self.subscription = subscription
        self.callback = callback
        return self.future


class _FakeOrchestrator:
    def __init__(self) -> None:
        self.queries: list[Any] = []

    async def collect(self, query: Any) -> AggregatedEvidence:
        self.queries.append(query)
        return AggregatedEvidence(
            evidence_id=query.evidence_id,
            tenant_id=query.tenant_id,
            case_id=query.case_id,
            fragments=[],
        )


def test_market_subscriber_accepts_canonical_and_legacy_alias() -> None:
    client = _FakeClient()
    handled: list[dict[str, object]] = []
    subscriber = MarketResearchRequestedSubscriber(
        client=client,
        project_id="proj",
        subscription_id="market-sub",
        handler=lambda payload: handled.append(payload),
    )
    subscriber.start()

    canonical = _FakeMessage(
        payload={"event_type": "market.research.requested", "payload": {}}
    )
    legacy = _FakeMessage(
        payload={
            "event_type": "MarketResearchRequested",
            "payload": {},
        }
    )

    client.callback(canonical)
    client.callback(legacy)

    assert len(handled) == 2
    assert canonical.acked is True and canonical.nacked is False
    assert legacy.acked is True and legacy.nacked is False


def test_market_handler_extracts_nested_payload() -> None:
    orchestrator = _FakeOrchestrator()
    handler = MarketResearchRequestedHandler(orchestrator=orchestrator)  # type: ignore[arg-type]

    handler(
        {
            "event_type": "market.research.requested",
            "tenant_id": "tenant-top-level",
            "payload": {
                "evidence_id": "e-1",
                "case_id": "c-1",
                "case_type": "new_project",
                "context": "Build analytics product",
                "region": "japan",
                "providers": ["grok", "brave"],
            },
        }
    )

    assert len(orchestrator.queries) == 1
    assert orchestrator.queries[0].tenant_id == "tenant-top-level"
    assert orchestrator.queries[0].providers == ("grok", "brave")


def test_postgres_market_repository_persists_fragments_and_aggregate() -> None:
    mock_cursor = MagicMock()
    mock_cursor.fetchone.side_effect = [None, ("fragment-1",)]
    mock_conn = MagicMock()
    mock_conn.cursor.return_value.__enter__.return_value = mock_cursor
    mock_conn.cursor.return_value.__exit__.return_value = False
    mock_conn.__enter__.return_value = mock_conn
    mock_conn.__exit__.return_value = False

    mock_manager = MagicMock()
    mock_manager.get_connection.return_value.__enter__.return_value = mock_conn
    mock_manager.get_connection.return_value.__exit__.return_value = False

    repository = PostgresMarketEvidenceRepository(mock_manager)
    fragment = EvidenceFragment(
        provider="grok",
        hourly_rate_range=Range(100, 160),
        total_hours_range=Range(200, 320),
        provider_confidence=0.8,
        raw_response="raw",
    )
    aggregate = AggregatedEvidence(
        evidence_id="e-1",
        tenant_id="t-1",
        case_id="c-1",
        fragments=[fragment],
        consensus_hours_range=Range(200, 320),
        consensus_rate_range=Range(100, 160),
        overall_confidence="medium",
    )

    repository.save(
        query=orchestrator_query(),
        aggregate=aggregate,
    )

    assert mock_cursor.execute.call_count == 3


def test_build_market_providers_skips_missing_api_keys() -> None:
    providers = build_market_providers(
        grok_api_key="grok",
        brave_api_key=None,
        perplexity_api_key="perplexity",
        gemini_api_key=None,
    )

    assert [provider.provider_name() for provider in providers] == [
        "grok",
        "perplexity",
    ]


def orchestrator_query() -> Any:
    return type(
        "_Query",
        (),
        {
            "tenant_id": "t-1",
            "evidence_id": "e-1",
            "case_id": "c-1",
            "case_type": "new_project",
            "context": "Build analytics product",
        },
    )()
