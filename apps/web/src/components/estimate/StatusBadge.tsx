import { cn } from '@/lib/cn'
import type { EstimateStatus } from '@/types/estimate'
import { estimateStatusLabels } from '@/types/estimate'

interface StatusBadgeProps {
  status: EstimateStatus
}

const statusStyles: Record<EstimateStatus, string> = {
  draft: 'bg-slate-100 text-slate-700',
  ready: 'bg-blue-100 text-blue-700',
  approved: 'bg-emerald-100 text-emerald-700',
  rejected: 'bg-rose-100 text-rose-700',
}

export function StatusBadge({ status }: StatusBadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
        statusStyles[status],
      )}
    >
      {estimateStatusLabels[status]}
    </span>
  )
}
