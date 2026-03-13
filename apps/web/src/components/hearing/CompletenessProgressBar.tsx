import { cn } from '@/lib/cn'

interface CompletenessProgressBarProps {
  value: number
  label?: string
}

function clamp(value: number) {
  return Math.min(1, Math.max(0, value))
}

export function CompletenessProgressBar({
  value,
  label = 'Hearing completeness',
}: CompletenessProgressBarProps) {
  const normalized = clamp(value)
  const percentage = Math.round(normalized * 100)

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between gap-4">
        <p className="text-sm font-medium text-slate-700">{label}</p>
        <span className="text-sm font-semibold tabular-nums text-slate-950">
          {percentage}%
        </span>
      </div>
      <div
        aria-label={label}
        aria-valuemax={100}
        aria-valuemin={0}
        aria-valuenow={percentage}
        role="progressbar"
        className="h-2.5 overflow-hidden rounded-full bg-slate-200"
      >
        <div
          className={cn(
            'h-full rounded-full bg-blue-600',
            percentage >= 80 && 'bg-emerald-600',
          )}
          style={{ width: `${percentage}%` }}
        />
      </div>
    </div>
  )
}
