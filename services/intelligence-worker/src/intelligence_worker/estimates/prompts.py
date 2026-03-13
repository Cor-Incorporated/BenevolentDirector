"""Prompt templates for ThreeWayProposal generation."""

from __future__ import annotations

import json
from typing import Any

THREE_WAY_PROPOSAL_SYSTEM_PROMPT = """\
You generate a Japanese ThreeWayProposal for software delivery estimates.

Rules:
- Return JSON only.
- Keep all numeric values grounded in the provided inputs.
- Do not invent citations, providers, hours, or rates that are not present.
- Echo `our_track_record` and `market_benchmark` from the inputs when available.
- `our_proposal.competitive_advantages` must contain 2 to 4 concise strings.
- `our_proposal.calibration_note` must explain how the proposal was calibrated.

Return this shape exactly:
{
  "our_track_record": {
    "similar_projects": [{"name": "string", "actual_hours": 0, "similarity_score": 0}],
    "median_hours": 0,
    "velocity_score": 0
  },
  "market_benchmark": {
    "consensus_hours": {"min": 0, "max": 0},
    "consensus_rate": {"min": 0, "max": 0},
    "confidence": "high|medium|low",
    "provider_count": 0,
    "citations": [
      {
        "url": "https://example.com",
        "title": "string",
        "source_authority": "official|industry|community|unknown",
        "snippet": "string"
      }
    ]
  },
  "our_proposal": {
    "proposed_hours": 0,
    "proposed_rate": 0,
    "proposed_total": 0,
    "savings_vs_market_percent": 0,
    "competitive_advantages": ["string"],
    "calibration_note": "string"
  }
}
"""


def build_three_way_proposal_prompt(*, payload: dict[str, Any]) -> str:
    """Render a single user prompt with all grounded estimate inputs."""
    return "\n".join(
        [
            "Generate a ThreeWayProposal for this estimate context.",
            (
                "Use the provided numbers as source of truth and explain the "
                "calibration in Japanese."
            ),
            "",
            json.dumps(payload, ensure_ascii=False, indent=2),
        ]
    )
