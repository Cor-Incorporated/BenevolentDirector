import type { MissingInfoPrompt } from '@/types/conversation'
import { MissingInfoPromptItem } from './MissingInfoPromptItem'

interface MissingInfoPromptListProps {
  prompts: MissingInfoPrompt[]
  isRefreshing?: boolean
}

export function MissingInfoPromptList({
  prompts,
  isRefreshing = false,
}: MissingInfoPromptListProps) {
  if (prompts.length === 0) {
    return (
      <div className="rounded-xl border border-dashed border-slate-300 bg-white px-4 py-4">
        <p className="text-sm font-medium text-slate-900">No follow-ups queued</p>
        <p className="mt-1 text-pretty text-sm text-slate-600">
          Continue the hearing to expose missing context, or wait for the next
          assistant turn to refresh this checklist.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between gap-3">
        <p className="text-sm font-medium text-slate-700">Missing information</p>
        {isRefreshing ? (
          <span className="text-xs text-slate-500">Refreshing…</span>
        ) : null}
      </div>
      <ul className="space-y-2">
        {prompts.map((prompt) => (
          <MissingInfoPromptItem key={prompt.id} prompt={prompt} />
        ))}
      </ul>
    </div>
  )
}
