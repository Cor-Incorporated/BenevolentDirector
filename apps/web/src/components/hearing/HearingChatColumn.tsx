import { MessageBubble } from '@/components/conversation/MessageBubble'
import { MessageList } from '@/components/conversation/MessageList'
import { StreamingIndicator } from '@/components/conversation/StreamingIndicator'
import type { ConversationTurn, MissingInfoPrompt } from '@/types/conversation'
import { HearingMessageInput } from './HearingMessageInput'
import { IntentClassificationBadge } from './IntentClassificationBadge'

interface HearingChatColumnProps {
  turns: ConversationTurn[]
  streamingContent: string
  isStreaming: boolean
  isLoading?: boolean
  prompts: MissingInfoPrompt[]
  onSend: (content: string) => Promise<void>
}

export function HearingChatColumn({
  turns,
  streamingContent,
  isStreaming,
  isLoading = false,
  prompts,
  onSend,
}: HearingChatColumnProps) {
  return (
    <section className="flex min-h-[65dvh] flex-col overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
      <div className="border-b border-slate-200 px-5 py-4">
        <p className="text-sm font-medium text-slate-500">Conversation</p>
        <h2 className="mt-1 text-balance text-xl font-semibold text-slate-950">
          Structured hearing session
        </h2>
      </div>

      <MessageList
        turns={turns}
        streamingContent={streamingContent}
        isStreaming={isStreaming}
        className="bg-slate-50"
        emptyState={(
          <div className="flex flex-1 items-center justify-center px-6 py-10">
            <div className="max-w-md text-center">
              <h3 className="text-balance text-lg font-semibold text-slate-950">
                {isLoading ? 'Loading hearing session…' : 'Start the intake interview'}
              </h3>
              <p className="mt-2 text-pretty text-sm text-slate-600">
                {isLoading
                  ? 'Conversation turns, source documents, and the latest requirement artifact are loading.'
                  : 'Ask for scope, risks, timing, and source material so the spec and completeness score can evolve in parallel.'}
              </p>
            </div>
          </div>
        )}
        renderTurn={(turn) => (
          <div className="space-y-2">
            {turn.role === 'assistant' ? (
              <div className="pl-1">
                <IntentClassificationBadge category={turn.metadata?.category} />
              </div>
            ) : null}
            <MessageBubble turn={turn} />
          </div>
        )}
        renderStreamingIndicator={(content) => (
          <div className="space-y-2">
            <div className="pl-1">
              <IntentClassificationBadge />
            </div>
            <StreamingIndicator content={content} />
          </div>
        )}
      />

      <HearingMessageInput
        disabled={isStreaming}
        prompts={prompts}
        onSend={onSend}
      />
    </section>
  )
}
