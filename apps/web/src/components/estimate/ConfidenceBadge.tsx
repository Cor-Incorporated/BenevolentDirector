import { cn } from '@/lib/cn'
import {
  confidenceLevelLabels,
  type ConfidenceLevel,
} from '@/types/estimate'

type ConfidenceBadgeProps = {
  level?: ConfidenceLevel | null
  className?: string
}

const toneByConfidence: Record<ConfidenceLevel, string> = {
  high: 'border-emerald-200 bg-emerald-50 text-emerald-700',
  medium: 'border-amber-200 bg-amber-50 text-amber-700',
  low: 'border-rose-200 bg-rose-50 text-rose-700',
}

export function ConfidenceBadge({
  level,
  className,
}: ConfidenceBadgeProps) {
  if (!level) {
    return (
      <span
        className={cn(
          'inline-flex items-center rounded-full border border-slate-200 bg-slate-100 px-2.5 py-1 text-xs font-medium text-slate-600',
          className,
        )}
      >
        Confidence unavailable
      </span>
    )
  }

  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full border px-2.5 py-1 text-xs font-medium',
        toneByConfidence[level],
        className,
      )}
    >
      {confidenceLevelLabels[level]}
    </span>
  )
}
