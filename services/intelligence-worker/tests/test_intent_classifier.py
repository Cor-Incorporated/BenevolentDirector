"""Tests for the intent classifier."""

from __future__ import annotations

import pytest

from intelligence_worker.classification.intent_classifier import (
    VALID_INTENTS,
    IntentClassifier,
)


@pytest.fixture()
def classifier() -> IntentClassifier:
    return IntentClassifier()


class TestNewProjectClassification:
    def test_japanese_new_project(self, classifier: IntentClassifier) -> None:
        result = classifier.classify("新規でWebアプリを作りたい")
        assert result.intent == "new_project"
        assert result.confidence >= 0.4

    def test_english_new_project(self, classifier: IntentClassifier) -> None:
        result = classifier.classify("I want to build a new project from scratch")
        assert result.intent == "new_project"
        assert result.confidence >= 0.4

    def test_scratch_keyword(self, classifier: IntentClassifier) -> None:
        result = classifier.classify("ゼロからシステムを構築したい")
        assert result.intent == "new_project"


class TestFeatureAdditionClassification:
    def test_japanese_feature_addition(self, classifier: IntentClassifier) -> None:
        result = classifier.classify("検索機能の追加機能を依頼したい")
        assert result.intent == "feature_addition"
        assert result.confidence >= 0.4

    def test_english_feature_addition(self, classifier: IntentClassifier) -> None:
        result = classifier.classify("We need to add a new feature for user export")
        assert result.intent == "feature_addition"
        assert result.confidence >= 0.4


class TestBugReportClassification:
    def test_japanese_bug_report(self, classifier: IntentClassifier) -> None:
        result = classifier.classify("ログイン画面でバグが発生しています")
        assert result.intent == "bug_report"
        assert result.confidence >= 0.4

    def test_english_bug_report(self, classifier: IntentClassifier) -> None:
        result = classifier.classify("There is a bug in the payment module")
        assert result.intent == "bug_report"
        assert result.confidence >= 0.4

    def test_crash_keyword(self, classifier: IntentClassifier) -> None:
        result = classifier.classify("The app keeps crashing")
        assert result.intent == "bug_report"


class TestFixRequestClassification:
    def test_japanese_fix_request(self, classifier: IntentClassifier) -> None:
        result = classifier.classify("この画面の表示を修正してほしい")
        assert result.intent == "fix_request"
        assert result.confidence >= 0.4

    def test_english_fix_request(self, classifier: IntentClassifier) -> None:
        result = classifier.classify("Please fix the layout issue")
        assert result.intent == "fix_request"
        assert result.confidence >= 0.4


class TestConsultationClassification:
    def test_japanese_consultation(self, classifier: IntentClassifier) -> None:
        result = classifier.classify("技術的なアドバイスをいただきたい")
        assert result.intent == "consultation"
        assert result.confidence >= 0.2

    def test_english_consultation(self, classifier: IntentClassifier) -> None:
        result = classifier.classify("I have a question about architecture")
        assert result.intent == "consultation"
        assert result.confidence >= 0.2


class TestEdgeCases:
    def test_empty_string_returns_consultation(
        self, classifier: IntentClassifier
    ) -> None:
        result = classifier.classify("")
        assert result.intent == "consultation"
        assert result.confidence == 0.1
        assert result.keywords == []

    def test_whitespace_only_returns_consultation(
        self, classifier: IntentClassifier
    ) -> None:
        result = classifier.classify("   ")
        assert result.intent == "consultation"
        assert result.confidence == 0.1

    def test_no_keywords_returns_consultation(
        self, classifier: IntentClassifier
    ) -> None:
        result = classifier.classify("Hello world")
        assert result.intent == "consultation"
        assert result.confidence == 0.2
        assert result.keywords == []

    def test_multiple_matches_boost_confidence(
        self, classifier: IntentClassifier
    ) -> None:
        result = classifier.classify(
            "新規プロジェクトを新しいシステムとしてゼロから作りたい"
        )
        assert result.intent == "new_project"
        assert result.confidence > 0.4

    def test_confidence_capped_at_max(self, classifier: IntentClassifier) -> None:
        result = classifier.classify(
            "新規 新しい ゼロから スクラッチ build from scratch "
            "new project new system new app start from"
        )
        assert result.confidence <= 0.95

    def test_result_intent_is_valid(self, classifier: IntentClassifier) -> None:
        result = classifier.classify("新規プロジェクト")
        assert result.intent in VALID_INTENTS

    def test_result_is_frozen_dataclass(self, classifier: IntentClassifier) -> None:
        result = classifier.classify("新規プロジェクト")
        with pytest.raises(AttributeError):
            result.intent = "other"  # type: ignore[misc]
