"""Chat completions endpoint (OpenAI-compatible stub)."""

from __future__ import annotations

import time
from typing import Any

from fastapi import APIRouter
from pydantic import BaseModel

router = APIRouter()


class ChatMessage(BaseModel):
    """A single message in the chat conversation."""

    role: str
    content: str


class ChatCompletionRequest(BaseModel):
    """OpenAI-compatible chat completion request."""

    model: str = "stub"
    messages: list[ChatMessage]
    temperature: float = 0.7
    max_tokens: int | None = None


@router.post("/v1/chat/completions")
async def chat_completions(request: ChatCompletionRequest) -> dict[str, Any]:
    """Return a mock OpenAI-compatible chat completion response.

    Args:
        request: The chat completion request body.

    Returns:
        A dict matching the OpenAI chat completion response format.
    """
    return {
        "id": "chatcmpl-stub",
        "object": "chat.completion",
        "created": int(time.time()),
        "model": request.model,
        "choices": [
            {
                "index": 0,
                "message": {
                    "role": "assistant",
                    "content": "This is a stub response from llm-gateway.",
                },
                "finish_reason": "stop",
            }
        ],
        "usage": {
            "prompt_tokens": 0,
            "completion_tokens": 0,
            "total_tokens": 0,
        },
    }
