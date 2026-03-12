"""Tests for the missing info extractor."""

from __future__ import annotations

from intelligence_worker.classification.missing_info import (
    MissingField,
    MissingInfoExtractor,
)


def _field_names(fields: list[MissingField]) -> list[str]:
    """Helper to extract field names from results."""
    return [f.field_name for f in fields]


class TestMissingInfoExtraction:
    def test_all_fields_missing_for_vague_request(self) -> None:
        extractor = MissingInfoExtractor()
        result = extractor.extract("何かアプリを作りたい")
        names = _field_names(result)
        assert "project_scope" in names
        assert "budget" in names
        assert "timeline" in names
        assert "tech_stack" in names
        assert "team_size" in names

    def test_scope_present_reduces_missing(self) -> None:
        extractor = MissingInfoExtractor()
        result = extractor.extract("ECサイトの要件を整理したい")
        names = _field_names(result)
        assert "project_scope" not in names
        assert "budget" in names

    def test_budget_present_reduces_missing(self) -> None:
        extractor = MissingInfoExtractor()
        result = extractor.extract("予算は500万円です")
        names = _field_names(result)
        assert "budget" not in names

    def test_timeline_present_reduces_missing(self) -> None:
        extractor = MissingInfoExtractor()
        result = extractor.extract("納期は来月末です")
        names = _field_names(result)
        assert "timeline" not in names

    def test_tech_stack_present_reduces_missing(self) -> None:
        extractor = MissingInfoExtractor()
        result = extractor.extract("Reactで開発したい")
        names = _field_names(result)
        assert "tech_stack" not in names

    def test_team_size_present_reduces_missing(self) -> None:
        extractor = MissingInfoExtractor()
        result = extractor.extract("チーム3人で進めたい")
        names = _field_names(result)
        assert "team_size" not in names

    def test_complete_request_returns_empty(self) -> None:
        extractor = MissingInfoExtractor()
        result = extractor.extract(
            "ECサイトの要件: 予算500万円、納期は来月末、"
            "Reactフレームワークで、チーム5人体制"
        )
        assert result == []

    def test_empty_text_returns_all_fields(self) -> None:
        extractor = MissingInfoExtractor()
        result = extractor.extract("")
        assert len(result) == 5

    def test_whitespace_only_returns_all_fields(self) -> None:
        extractor = MissingInfoExtractor()
        result = extractor.extract("   ")
        assert len(result) == 5


class TestMissingFieldPriority:
    def test_results_sorted_by_priority(self) -> None:
        extractor = MissingInfoExtractor()
        result = extractor.extract("何かアプリを作りたい")
        priorities = [f.priority for f in result]
        expected_order = ["high", "high", "medium", "medium", "low"]
        assert priorities == expected_order

    def test_high_priority_fields(self) -> None:
        extractor = MissingInfoExtractor()
        result = extractor.extract("何かアプリを作りたい")
        high_fields = [f.field_name for f in result if f.priority == "high"]
        assert "project_scope" in high_fields
        assert "budget" in high_fields


class TestMissingFieldDataclass:
    def test_field_has_question(self) -> None:
        extractor = MissingInfoExtractor()
        result = extractor.extract("何かアプリを作りたい")
        for f in result:
            assert f.question
            assert len(f.question) > 0

    def test_field_is_frozen(self) -> None:
        field = MissingField(
            field_name="test",
            question="test?",
            priority="high",
        )
        try:
            field.field_name = "other"  # type: ignore[misc]
            assert False, "Should have raised"  # noqa: B011
        except AttributeError:
            pass
