"""Publish Observation Pipeline events to Pub/Sub."""

from __future__ import annotations

import json
import time
import uuid
from typing import TYPE_CHECKING, Protocol

if TYPE_CHECKING:
    from intelligence_worker.completeness_tracker import CompletenessTrackingSnapshot


class PublishFuture(Protocol):
    """Minimal publish future contract."""

    def result(self, timeout: float | None = None) -> str: ...


class PublisherClient(Protocol):
    """Minimal Pub/Sub publisher client contract."""

    def topic_path(self, project_id: str, topic_id: str) -> str: ...

    def publish(
        self,
        topic: str,
        data: bytes,
        ordering_key: str = "",
    ) -> PublishFuture: ...


class CompletenessUpdatedPublisher:
    """Publish `observation.completeness.updated` envelopes."""

    EVENT_TYPE = "observation.completeness.updated"
    AGGREGATE_TYPE = "observation"
    PRODUCER = "intelligence-worker"
    SCHEMA_VERSION = "1.0.0"

    def __init__(
        self,
        *,
        client: PublisherClient,
        project_id: str,
        topic_id: str,
    ) -> None:
        self._client = client
        self._topic = client.topic_path(project_id, topic_id)

    def publish_snapshot(
        self,
        *,
        tenant_id: str,
        session_id: str,
        source_domain: str,
        aggregate_version: int,
        snapshot: CompletenessTrackingSnapshot,
        causation_id: str | None = None,
        correlation_id: str | None = None,
    ) -> str:
        event_id = str(uuid.uuid4())
        payload = {
            "event_id": event_id,
            "event_type": self.EVENT_TYPE,
            "schema_version": self.SCHEMA_VERSION,
            "aggregate_type": self.AGGREGATE_TYPE,
            "aggregate_id": session_id,
            "aggregate_version": aggregate_version,
            "idempotency_key": f"{session_id}:{aggregate_version}:completeness",
            "occurred_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
            "producer": self.PRODUCER,
            "tenant_id": tenant_id,
            "source_domain": source_domain,
            "payload": {
                "session_id": session_id,
                "checklist": {
                    key: {
                        "status": value.status,
                        "confidence": value.confidence,
                    }
                    for key, value in snapshot.checklist.items()
                },
                "overall_completeness": snapshot.overall_completeness,
                "suggested_next_topics": list(snapshot.suggested_next_topics),
            },
        }
        if causation_id:
            payload["causation_id"] = causation_id
        if correlation_id:
            payload["correlation_id"] = correlation_id

        future = self._client.publish(
            self._topic,
            json.dumps(payload, ensure_ascii=False).encode("utf-8"),
            ordering_key=session_id,
        )
        return future.result(timeout=5)
