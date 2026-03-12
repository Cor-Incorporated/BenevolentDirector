"""Rule-based intent classifier for incoming requests.

This module provides a keyword-matching classifier with confidence
scoring as a lightweight stub.  It can be replaced with an LLM-based
classifier later without changing the public interface.
"""

from __future__ import annotations

import logging
import re
from dataclasses import dataclass, field

logger = logging.getLogger(__name__)

# Valid intent categories
VALID_INTENTS: tuple[str, ...] = (
    "new_project",
    "feature_addition",
    "bug_report",
    "fix_request",
    "consultation",
)

# Keyword patterns mapped to intents.  Each entry is a compiled regex
# that matches one or more keywords associated with the intent.
_INTENT_PATTERNS: dict[str, re.Pattern[str]] = {
    "new_project": re.compile(
        r"(新規|新しい|ゼロから|スクラッチ|build from scratch"
        r"|new project|new system|new app|start from)",
        re.IGNORECASE,
    ),
    "feature_addition": re.compile(
        r"(機能追加|追加機能|新機能|add feature|new feature"
        r"|enhance|拡張|improvement)",
        re.IGNORECASE,
    ),
    "bug_report": re.compile(
        r"(バグ|不具合|エラー|bug|defect|broken|crash"
        r"|動かない|壊れ|おかしい)",
        re.IGNORECASE,
    ),
    "fix_request": re.compile(
        r"(修正|直し|fix|repair|patch|hotfix|改修"
        r"|対応して|修復)",
        re.IGNORECASE,
    ),
    "consultation": re.compile(
        r"(相談|コンサル|アドバイス|consult|advice"
        r"|question|質問|教えて|検討)",
        re.IGNORECASE,
    ),
}

# Base confidence for a single keyword match
_BASE_CONFIDENCE = 0.4
# Bonus per additional keyword match (capped at 0.95 total)
_MATCH_BONUS = 0.15
_MAX_CONFIDENCE = 0.95


@dataclass(frozen=True)
class ClassificationResult:
    """Result of intent classification.

    Attributes:
        intent: One of the valid intent categories.
        confidence: Confidence score between 0.0 and 1.0.
        keywords: Keywords that triggered the classification.
    """

    intent: str
    confidence: float
    keywords: list[str] = field(default_factory=list)


class IntentClassifier:
    """Classify raw request text into intent categories.

    Uses keyword pattern matching with confidence scoring.  Designed
    as a drop-in stub that can be swapped with an LLM classifier.
    """

    def classify(self, raw_text: str) -> ClassificationResult:
        """Classify the given text into an intent category.

        Args:
            raw_text: Unstructured request text from the user.

        Returns:
            A ClassificationResult with intent, confidence, and
            matched keywords.
        """
        if not raw_text or not raw_text.strip():
            logger.warning("Empty text received for classification")
            return ClassificationResult(
                intent="consultation",
                confidence=0.1,
                keywords=[],
            )

        scores = self._score_intents(raw_text)
        best_intent, best_score, best_keywords = self._pick_best(scores)

        logger.info(
            "Classified intent",
            extra={
                "intent": best_intent,
                "confidence": best_score,
                "keywords": best_keywords,
            },
        )
        return ClassificationResult(
            intent=best_intent,
            confidence=best_score,
            keywords=best_keywords,
        )

    def _score_intents(self, text: str) -> dict[str, tuple[float, list[str]]]:
        """Score each intent based on keyword matches.

        Args:
            text: Raw request text.

        Returns:
            Mapping of intent to (score, matched_keywords).
        """
        results: dict[str, tuple[float, list[str]]] = {}
        for intent, pattern in _INTENT_PATTERNS.items():
            matches = pattern.findall(text)
            if matches:
                score = min(
                    _BASE_CONFIDENCE + _MATCH_BONUS * (len(matches) - 1),
                    _MAX_CONFIDENCE,
                )
                results[intent] = (score, list(dict.fromkeys(matches)))
            else:
                results[intent] = (0.0, [])
        return results

    @staticmethod
    def _pick_best(
        scores: dict[str, tuple[float, list[str]]],
    ) -> tuple[str, float, list[str]]:
        """Select the highest-scoring intent.

        Falls back to ``consultation`` with low confidence when no
        keywords match.

        Args:
            scores: Mapping of intent to (score, keywords).

        Returns:
            Tuple of (intent, confidence, keywords).
        """
        best_intent = "consultation"
        best_score = 0.0
        best_keywords: list[str] = []

        for intent, (score, keywords) in scores.items():
            if score > best_score:
                best_intent = intent
                best_score = score
                best_keywords = keywords

        if best_score == 0.0:
            return "consultation", 0.2, []

        return best_intent, best_score, best_keywords
