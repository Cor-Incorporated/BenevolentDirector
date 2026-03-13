import { cn } from '@/lib/cn'
import type { MissingInfoPrompt } from '@/types/conversation'

interface MissingInfoPromptItemProps {
  prompt: MissingInfoPrompt
}

export function MissingInfoPromptItem({
  prompt,
}: MissingInfoPromptItemProps) {
  return (
    <li className="rounded-xl border border-slate-200 bg-slate-50 px-3 py-3">
      <div className="flex items-start justify-between gap-3">
        <p className="text-sm font-medium text-slate-900">{prompt.label}</p>
        <span
          className={cn(
            'rounded-full px-2 py-0.5 text-[11px] font-medium tabular-nums',
            prompt.needsFollowup
              ? 'bg-amber-100 text-amber-700'
              : 'bg-slate-200 text-slate-600',
          )}
        >
          {Math.round(prompt.completeness * 100)}%
        </span>
      </div>
      {prompt.detail ? (
        <p className="mt-2 text-pretty text-xs text-slate-600">
          {prompt.detail}
        </p>
      ) : null}
    </li>
  )
}
