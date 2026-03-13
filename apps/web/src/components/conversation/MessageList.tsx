import { type ReactNode, useEffect, useRef } from 'react'
import { cn } from '@/lib/cn'
import type { ConversationTurn } from '@/types/conversation'
import { MessageBubble } from './MessageBubble'
import { StreamingIndicator } from './StreamingIndicator'

interface MessageListProps {
  turns: ConversationTurn[]
  streamingContent: string
  isStreaming: boolean
  className?: string
  emptyState?: ReactNode
  renderTurn?: (turn: ConversationTurn) => ReactNode
  renderStreamingIndicator?: (content: string) => ReactNode
}

export function MessageList({
  turns,
  streamingContent,
  isStreaming,
  className,
  emptyState,
  renderTurn,
  renderStreamingIndicator,
}: MessageListProps) {
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (typeof bottomRef.current?.scrollIntoView === 'function') {
      bottomRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [turns.length, streamingContent])

  if (turns.length === 0 && !isStreaming) {
    return emptyState ?? (
      <div className="flex-1 flex items-center justify-center text-gray-400">
        <p className="text-sm">Send a message to start the conversation.</p>
      </div>
    )
  }

  return (
    <div className={cn('flex-1 overflow-y-auto px-4 py-6', className)}>
      {turns.map((turn) => (
        <div key={turn.id}>
          {renderTurn ? renderTurn(turn) : <MessageBubble turn={turn} />}
        </div>
      ))}
      {isStreaming && (
        renderStreamingIndicator
          ? renderStreamingIndicator(streamingContent)
          : <StreamingIndicator content={streamingContent} />
      )}
      <div ref={bottomRef} />
    </div>
  )
}
