import { cn } from '@/lib/cn'

interface IntentClassificationBadgeProps {
  category?: string | undefined
}

const categoryStyles: Array<{
  match: RegExp
  className: string
  label: string
}> = [
  {
    match: /scope|requirement|feature/i,
    className: 'border-blue-200 bg-blue-50 text-blue-700',
    label: 'Scope',
  },
  {
    match: /constraint|risk|issue|blocker/i,
    className: 'border-amber-200 bg-amber-50 text-amber-700',
    label: 'Constraint',
  },
  {
    match: /timeline|schedule|deadline/i,
    className: 'border-cyan-200 bg-cyan-50 text-cyan-700',
    label: 'Timeline',
  },
  {
    match: /budget|cost|estimate|pricing/i,
    className: 'border-emerald-200 bg-emerald-50 text-emerald-700',
    label: 'Budget',
  },
]

function getCategoryStyle(category?: string) {
  const trimmed = category?.trim()
  if (!trimmed) {
    return {
      className: 'border-slate-200 bg-slate-100 text-slate-600',
      label: 'Assistant',
    }
  }

  return (
    categoryStyles.find((style) => style.match.test(trimmed)) ?? {
      className: 'border-slate-200 bg-slate-100 text-slate-700',
      label: trimmed.replace(/[_-]+/g, ' '),
    }
  )
}

export function IntentClassificationBadge({
  category,
}: IntentClassificationBadgeProps) {
  const style = getCategoryStyle(category)

  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full border px-2.5 py-1 text-xs font-medium capitalize',
        style.className,
      )}
    >
      {style.label}
    </span>
  )
}
