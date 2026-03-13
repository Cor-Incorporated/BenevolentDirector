"""Runtime wiring for market intelligence collection."""

from __future__ import annotations

import asyncio
from dataclasses import dataclass
from typing import TYPE_CHECKING, Any

import structlog

from intelligence_worker.market.models import MarketQuery
from intelligence_worker.market.providers import (
    AsyncHTTPClient,
    BraveMarketProvider,
    GeminiMarketProvider,
    GrokMarketProvider,
    MarketProvider,
    PerplexityMarketProvider,
)
from intelligence_worker.market.repository import MarketEvidenceRepository

if TYPE_CHECKING:
    from intelligence_worker.market.orchestrator import MarketIntelligenceOrchestrator

logger = structlog.get_logger()

PostgresMarketEvidenceRepository = MarketEvidenceRepository


@dataclass(slots=True)
class MarketResearchRequestedHandler:
    """Sync Pub/Sub callback that runs async market orchestration."""

    orchestrator: MarketIntelligenceOrchestrator

    def __call__(self, payload: dict[str, Any]) -> None:
        query = MarketQuery.from_payload(_extract_market_payload(payload))
        logger.info(
            "market_research_requested",
            evidence_id=query.evidence_id,
            tenant_id=query.tenant_id,
            case_id=query.case_id,
            providers=list(query.providers),
        )
        evidence = asyncio.run(self.orchestrator.collect(query))
        logger.info(
            "market_research_completed",
            evidence_id=evidence.id,
            tenant_id=evidence.tenant_id,
            fragment_count=len(evidence.fragments),
            overall_confidence=evidence.overall_confidence,
        )


def build_market_providers(
    *,
    grok_api_key: str | None,
    brave_api_key: str | None,
    perplexity_api_key: str | None,
    gemini_api_key: str | None,
    client: AsyncHTTPClient | None = None,
) -> list[MarketProvider]:
    providers: list[MarketProvider] = []
    if grok_api_key:
        providers.append(GrokMarketProvider(api_key=grok_api_key, client=client))
    if brave_api_key:
        providers.append(BraveMarketProvider(api_key=brave_api_key, client=client))
    if perplexity_api_key:
        providers.append(
            PerplexityMarketProvider(api_key=perplexity_api_key, client=client)
        )
    if gemini_api_key:
        providers.append(GeminiMarketProvider(api_key=gemini_api_key, client=client))
    return providers


def _extract_market_payload(payload: dict[str, Any]) -> dict[str, Any]:
    nested = payload.get("payload")
    if not isinstance(nested, dict):
        return payload

    merged = dict(nested)
    for field in ("tenant_id", "case_id", "evidence_id", "job_id"):
        if field not in merged and isinstance(payload.get(field), str):
            merged[field] = payload[field]
    return merged
