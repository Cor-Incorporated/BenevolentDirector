"""Missing information extractor for incoming requests.

Identifies which fields are absent from a request so the system can
prompt the user for additional details.
"""

from __future__ import annotations

import logging
import re
from dataclasses import dataclass

logger = logging.getLogger(__name__)


@dataclass(frozen=True)
class MissingField:
    """A field that is missing from the request.

    Attributes:
        field_name: Canonical name of the missing field.
        question: Suggested follow-up question for the user.
        priority: ``high``, ``medium``, or ``low``.
    """

    field_name: str
    question: str
    priority: str


# Each detector returns ``True`` when the field is *present* in the
# text so the extractor can skip it.
_FIELD_DETECTORS: list[tuple[str, re.Pattern[str], str, str]] = [
    (
        "project_scope",
        re.compile(
            r"(スコープ|scope|要件|requirement|仕様|spec"
            r"|やりたいこと|目的|目標|ゴール|goal)",
            re.IGNORECASE,
        ),
        "プロジェクトのスコープや要件を教えてください。",
        "high",
    ),
    (
        "budget",
        re.compile(
            r"(予算|budget|費用|コスト|cost|金額|万円|百万|億)",
            re.IGNORECASE,
        ),
        "予算感を教えてください。",
        "high",
    ),
    (
        "timeline",
        re.compile(
            r"(期限|deadline|納期|スケジュール|schedule"
            r"|いつまで|timeline|月末|年末|quarter)",
            re.IGNORECASE,
        ),
        "希望する納期やスケジュールを教えてください。",
        "medium",
    ),
    (
        "tech_stack",
        re.compile(
            r"(技術|tech|stack|言語|language|framework"
            r"|フレームワーク|react|python|go|java|ruby|node)",
            re.IGNORECASE,
        ),
        "使用する技術スタックや制約はありますか？",
        "medium",
    ),
    (
        "team_size",
        re.compile(
            r"(チーム|team|人数|人員|メンバー|member|体制"
            r"|エンジニア|developer|人月)",
            re.IGNORECASE,
        ),
        "チーム体制やリソースの想定はありますか？",
        "low",
    ),
]


class MissingInfoExtractor:
    """Detect missing information fields in a request.

    Scans the raw text for indicators of each required field and
    returns a list of fields that appear to be absent.
    """

    def extract(self, raw_text: str) -> list[MissingField]:
        """Identify missing fields in the given text.

        Args:
            raw_text: Unstructured request text from the user.

        Returns:
            List of MissingField objects for fields not detected
            in the text, ordered by priority (high first).
        """
        if not raw_text or not raw_text.strip():
            logger.warning("Empty text received for missing info extraction")
            return [
                MissingField(
                    field_name=name,
                    question=question,
                    priority=priority,
                )
                for name, _, question, priority in _FIELD_DETECTORS
            ]

        missing: list[MissingField] = []
        for name, pattern, question, priority in _FIELD_DETECTORS:
            if not pattern.search(raw_text):
                missing.append(
                    MissingField(
                        field_name=name,
                        question=question,
                        priority=priority,
                    )
                )

        priority_order = {"high": 0, "medium": 1, "low": 2}
        missing.sort(key=lambda f: priority_order.get(f.priority, 9))

        logger.info(
            "Missing info extraction complete",
            extra={
                "missing_count": len(missing),
                "fields": [f.field_name for f in missing],
            },
        )
        return missing
