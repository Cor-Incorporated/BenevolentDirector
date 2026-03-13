"""Tests for estimate subscriber and runtime wiring."""

from __future__ import annotations

import json
from dataclasses import dataclass, field
from typing import Any
from unittest.mock import MagicMock

from intelligence_worker.estimates.models import EstimateQuery, OurProposal
from intelligence_worker.estimates.runtime import (
    EstimateRequestedHandler,
    start_estimate_subscriber,
)
from intelligence_worker.estimates.subscriber import EstimateRequestedSubscriber


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
        self.queries: list[EstimateQuery] = []

    def generate(self, query: EstimateQuery) -> Any:
        self.queries.append(query)
        return type(
            "_Proposal",
            (),
            {
                "our_proposal": OurProposal(
                    proposed_hours=180.0,
                    proposed_rate=12000.0,
                    proposed_total=2160000.0,
                    savings_vs_market_percent=31.4,
                    competitive_advantages=["実績ベース"],
                    calibration_note="実績で調整",
                )
            },
        )()


def test_estimate_subscriber_accepts_canonical_and_legacy_events() -> None:
    client = _FakeClient()
    handled: list[dict[str, object]] = []
    subscriber = EstimateRequestedSubscriber(
        client=client,
        project_id="proj",
        subscription_id="estimate-sub",
        handler=lambda payload: handled.append(payload),
    )
    subscriber.start()

    canonical = _FakeMessage(
        payload={"event_type": "estimate.requested", "payload": {"estimate_id": "e-1"}}
    )
    legacy = _FakeMessage(
        payload={"event_name": "EstimateRequested", "payload": {"estimate_id": "e-2"}}
    )

    assert client.callback is not None
    client.callback(canonical)
    client.callback(legacy)

    assert len(handled) == 2
    assert canonical.acked is True and canonical.nacked is False
    assert legacy.acked is True and legacy.nacked is False


def test_estimate_handler_extracts_nested_payload() -> None:
    orchestrator = _FakeOrchestrator()
    handler = EstimateRequestedHandler(orchestrator=orchestrator)  # type: ignore[arg-type]

    handler(
        {
            "event_type": "estimate.requested",
            "tenant_id": "tenant-top-level",
            "payload": {
                "estimate_id": "estimate-1",
                "case_id": "case-1",
            },
        }
    )

    assert len(orchestrator.queries) == 1
    assert orchestrator.queries[0].tenant_id == "tenant-top-level"
    assert orchestrator.queries[0].estimate_id == "estimate-1"


def test_start_estimate_subscriber_wires_runtime() -> None:
    config = type(
        "_Config",
        (),
        {
            "pubsub_project_id": "proj",
            "estimate_pubsub_subscription": "estimate-sub",
            "llm_gateway_url": "http://gateway:8081",
            "structured_output_model": "qwen3.5-7b",
        },
    )()
    subscriber_client = _FakeClient()

    runtime = start_estimate_subscriber(
        config=config,
        subscriber_client=subscriber_client,
        conn_manager=MagicMock(),
    )

    assert runtime.future is subscriber_client.future
    assert runtime.subscription_id == "estimate-sub"
    assert subscriber_client.subscription == "projects/proj/subscriptions/estimate-sub"
