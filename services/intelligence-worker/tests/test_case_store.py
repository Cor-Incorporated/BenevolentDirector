"""Tests for the in-memory case store."""

from __future__ import annotations

import pytest

from intelligence_worker.classification.intent_classifier import (
    ClassificationResult,
)
from intelligence_worker.classification.missing_info import (
    MissingField,
)
from intelligence_worker.store.case_store import CaseStore


@pytest.fixture()
def store() -> CaseStore:
    return CaseStore()


class TestCaseCreate:
    def test_create_returns_case(self, store: CaseStore) -> None:
        case = store.create("c1", "tenant-a", "新規開発の依頼")
        assert case.id == "c1"
        assert case.tenant_id == "tenant-a"
        assert case.raw_text == "新規開発の依頼"
        assert case.status == "pending"
        assert case.classification is None
        assert case.missing_fields == []

    def test_create_duplicate_raises(self, store: CaseStore) -> None:
        store.create("c1", "tenant-a", "text")
        with pytest.raises(ValueError, match="already exists"):
            store.create("c1", "tenant-a", "other text")

    def test_create_sets_timestamp(self, store: CaseStore) -> None:
        case = store.create("c1", "tenant-a", "text")
        assert case.created_at is not None


class TestCaseGet:
    def test_get_existing_case(self, store: CaseStore) -> None:
        store.create("c1", "tenant-a", "text")
        case = store.get("c1")
        assert case is not None
        assert case.id == "c1"

    def test_get_nonexistent_returns_none(self, store: CaseStore) -> None:
        assert store.get("nonexistent") is None


class TestUpdateClassification:
    def test_update_classification_success(self, store: CaseStore) -> None:
        store.create("c1", "tenant-a", "新規プロジェクト")
        result = ClassificationResult(
            intent="new_project",
            confidence=0.8,
            keywords=["新規"],
        )
        updated = store.update_classification("c1", result)
        assert updated.classification == result
        assert updated.status == "classified"

    def test_update_classification_nonexistent_raises(self, store: CaseStore) -> None:
        result = ClassificationResult(
            intent="new_project",
            confidence=0.8,
            keywords=[],
        )
        with pytest.raises(KeyError, match="Case not found"):
            store.update_classification("nonexistent", result)

    def test_update_preserves_other_fields(self, store: CaseStore) -> None:
        original = store.create("c1", "tenant-a", "raw text")
        result = ClassificationResult(
            intent="bug_report",
            confidence=0.5,
            keywords=["bug"],
        )
        updated = store.update_classification("c1", result)
        assert updated.tenant_id == original.tenant_id
        assert updated.raw_text == original.raw_text
        assert updated.created_at == original.created_at


class TestUpdateMissingFields:
    def test_update_missing_fields_success(self, store: CaseStore) -> None:
        store.create("c1", "tenant-a", "text")
        fields = [
            MissingField(
                field_name="budget",
                question="予算は？",
                priority="high",
            )
        ]
        updated = store.update_missing_fields("c1", fields)
        assert updated.missing_fields == fields
        assert updated.status == "needs_info"

    def test_empty_fields_keeps_status(self, store: CaseStore) -> None:
        store.create("c1", "tenant-a", "text")
        updated = store.update_missing_fields("c1", [])
        assert updated.missing_fields == []
        assert updated.status == "pending"

    def test_update_missing_fields_nonexistent_raises(self, store: CaseStore) -> None:
        with pytest.raises(KeyError, match="Case not found"):
            store.update_missing_fields("nonexistent", [])


class TestCaseImmutability:
    def test_case_is_frozen(self, store: CaseStore) -> None:
        case = store.create("c1", "tenant-a", "text")
        with pytest.raises(AttributeError):
            case.status = "other"  # type: ignore[misc]

    def test_get_returns_same_reference_after_update(self, store: CaseStore) -> None:
        store.create("c1", "tenant-a", "text")
        result = ClassificationResult(
            intent="new_project",
            confidence=0.8,
            keywords=[],
        )
        store.update_classification("c1", result)
        fetched = store.get("c1")
        assert fetched is not None
        assert fetched.classification == result
        assert fetched.status == "classified"
