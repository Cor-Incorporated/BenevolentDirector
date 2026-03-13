"""Runtime wiring for estimate proposal generation."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Protocol

import structlog

from intelligence_worker.estimates.models import EstimateQuery
from intelligence_worker.estimates.orchestrator import (
    EstimateOrchestrator,
    GatewayThreeWayProposalClient,
)
from intelligence_worker.estimates.repository import EstimateRepository
from intelligence_worker.estimates.subscriber import (
    EstimateRequestedSubscriber,
    StreamingPullFuture,
    SubscriberClient,
)

logger = structlog.get_logger()

PostgresEstimateRepository = EstimateRepository


class EstimateRuntimeConfig(Protocol):
    """Subset of worker config required for estimate subscription wiring."""

    pubsub_project_id: str
    estimate_pubsub_subscription: str
    llm_gateway_url: str
    structured_output_model: str


@dataclass(slots=True)
class EstimateRequestedHandler:
    """Sync Pub/Sub callback that runs ThreeWayProposal generation."""

    orchestrator: EstimateOrchestrator

    def __call__(self, payload: dict[str, Any]) -> None:
        query = EstimateQuery.from_payload(payload)
        logger.info(
            "estimate_requested",
            estimate_id=query.estimate_id,
            tenant_id=query.tenant_id,
            case_id=query.case_id,
        )
        proposal = self.orchestrator.generate(query)
        logger.info(
            "estimate_ready",
            estimate_id=query.estimate_id,
            tenant_id=query.tenant_id,
            proposal_hours=proposal.our_proposal.proposed_hours,
            proposal_total=proposal.our_proposal.proposed_total,
        )


@dataclass(slots=True)
class EstimateRuntime:
    """Started estimate runtime resources owned by the worker process."""

    future: StreamingPullFuture
    subscription_id: str

    def close(self) -> None:
        """Cancel the streaming pull future."""
        self.future.cancel()


def start_estimate_subscriber(
    *,
    config: EstimateRuntimeConfig,
    subscriber_client: SubscriberClient,
    conn_manager: Any,
) -> EstimateRuntime:
    estimate_handler = EstimateRequestedHandler(
        orchestrator=EstimateOrchestrator(
            repository=PostgresEstimateRepository(conn_manager),
            gateway_client=GatewayThreeWayProposalClient(
                base_url=config.llm_gateway_url,
                model=config.structured_output_model,
            ),
        )
    )
    estimate_subscriber = EstimateRequestedSubscriber(
        client=subscriber_client,
        project_id=config.pubsub_project_id,
        subscription_id=config.estimate_pubsub_subscription,
        handler=estimate_handler,
    )
    future = estimate_subscriber.start()
    logger.info(
        "estimate_subscriber_started",
        project_id=config.pubsub_project_id,
        subscription=config.estimate_pubsub_subscription,
    )
    return EstimateRuntime(
        future=future,
        subscription_id=config.estimate_pubsub_subscription,
    )
