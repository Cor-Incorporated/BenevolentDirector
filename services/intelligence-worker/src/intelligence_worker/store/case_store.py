"""In-memory case store stub.

Provides CRUD operations for intake cases.  This is a lightweight
in-memory implementation that will be replaced with a PostgreSQL-backed
store in a later phase.
"""

from __future__ import annotations

import logging
from dataclasses import dataclass, field, replace
from datetime import UTC, datetime
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from intelligence_worker.classification.intent_classifier import (
        ClassificationResult,
    )
    from intelligence_worker.classification.missing_info import (
        MissingField,
    )

logger = logging.getLogger(__name__)


@dataclass(frozen=True)
class Case:
    """An intake case representing a single user request.

    Attributes:
        id: Unique case identifier.
        tenant_id: Tenant that owns this case.
        raw_text: Original unstructured request text.
        classification: Intent classification result, if available.
        missing_fields: Fields still needed from the user.
        status: Current case status.
        created_at: Timestamp when the case was created.
    """

    id: str
    tenant_id: str
    raw_text: str
    classification: ClassificationResult | None = None
    missing_fields: list[MissingField] = field(default_factory=list)
    status: str = "pending"
    created_at: datetime = field(default_factory=lambda: datetime.now(tz=UTC))


class CaseStore:
    """In-memory case store.

    Thread-safety is not guaranteed; this stub is intended for
    single-threaded test and development use.
    """

    def __init__(self) -> None:
        self._cases: dict[str, Case] = {}

    def create(self, case_id: str, tenant_id: str, raw_text: str) -> Case:
        """Create a new case.

        Args:
            case_id: Unique identifier for the case.
            tenant_id: Tenant that owns this case.
            raw_text: Original request text.

        Returns:
            The newly created Case.

        Raises:
            ValueError: If a case with the same ID already exists.
        """
        if case_id in self._cases:
            raise ValueError(f"Case already exists: {case_id}")

        case = Case(
            id=case_id,
            tenant_id=tenant_id,
            raw_text=raw_text,
        )
        self._cases[case_id] = case
        logger.info(
            "Case created",
            extra={"case_id": case_id, "tenant_id": tenant_id},
        )
        return case

    def get(self, case_id: str) -> Case | None:
        """Retrieve a case by ID.

        Args:
            case_id: Unique identifier of the case.

        Returns:
            The Case if found, otherwise None.
        """
        return self._cases.get(case_id)

    def update_classification(self, case_id: str, result: ClassificationResult) -> Case:
        """Attach a classification result to an existing case.

        Args:
            case_id: Unique identifier of the case.
            result: Classification result to attach.

        Returns:
            The updated Case.

        Raises:
            KeyError: If the case does not exist.
        """
        existing = self._cases.get(case_id)
        if existing is None:
            raise KeyError(f"Case not found: {case_id}")

        updated = replace(
            existing,
            classification=result,
            status="classified",
        )
        self._cases[case_id] = updated
        logger.info(
            "Case classification updated",
            extra={
                "case_id": case_id,
                "intent": result.intent,
            },
        )
        return updated

    def update_missing_fields(self, case_id: str, fields: list[MissingField]) -> Case:
        """Attach missing-field information to an existing case.

        Args:
            case_id: Unique identifier of the case.
            fields: List of missing fields to attach.

        Returns:
            The updated Case.

        Raises:
            KeyError: If the case does not exist.
        """
        existing = self._cases.get(case_id)
        if existing is None:
            raise KeyError(f"Case not found: {case_id}")

        status = "needs_info" if fields else existing.status
        updated = replace(
            existing,
            missing_fields=fields,
            status=status,
        )
        self._cases[case_id] = updated
        logger.info(
            "Case missing fields updated",
            extra={
                "case_id": case_id,
                "missing_count": len(fields),
            },
        )
        return updated
