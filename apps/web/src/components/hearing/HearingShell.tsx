import { type ReactNode, useState } from 'react'
import { Link } from 'react-router-dom'
import { cn } from '@/lib/cn'

interface HearingShellProps {
  caseId: string
  error?: string | null
  leftPanel: ReactNode
  chatColumn: ReactNode
  rightPanel: ReactNode
}

export function HearingShell({
  caseId,
  error,
  leftPanel,
  chatColumn,
  rightPanel,
}: HearingShellProps) {
  const [showLeftPanel, setShowLeftPanel] = useState(false)
  const [showRightPanel, setShowRightPanel] = useState(false)

  return (
    <div className="space-y-4">
      <header className="rounded-3xl border border-slate-200 bg-white px-5 py-4 shadow-sm">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div className="space-y-2">
            <Link
              to="/cases"
              className="inline-flex text-sm text-slate-500 hover:text-slate-700"
            >
              ← Back to cases
            </Link>
            <div>
              <p className="text-sm font-medium text-slate-500">Hearing</p>
              <h1 className="text-balance text-2xl font-semibold text-slate-950">
                Three-panel intake workspace
              </h1>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <span className="rounded-full bg-slate-100 px-3 py-1 text-sm text-slate-600">
              {caseId}
            </span>
            <button
              type="button"
              onClick={() => setShowLeftPanel((value) => !value)}
              className="rounded-xl border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 lg:hidden"
            >
              {showLeftPanel ? 'Hide sources' : 'Show sources'}
            </button>
            <button
              type="button"
              onClick={() => setShowRightPanel((value) => !value)}
              className="rounded-xl border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 lg:hidden"
            >
              {showRightPanel ? 'Hide spec' : 'Show spec'}
            </button>
          </div>
        </div>

        {error ? (
          <p className="mt-4 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
            {error}
          </p>
        ) : null}
      </header>

      <div className="space-y-4 lg:hidden">
        {showLeftPanel ? <div>{leftPanel}</div> : null}
        {showRightPanel ? <div>{rightPanel}</div> : null}
      </div>

      <div className="grid gap-4 lg:grid-cols-[240px,minmax(0,1fr),320px]">
        <div className="hidden lg:block">{leftPanel}</div>
        <div className={cn(showLeftPanel || showRightPanel ? 'order-last lg:order-none' : '')}>
          {chatColumn}
        </div>
        <div className="hidden lg:block">{rightPanel}</div>
      </div>
    </div>
  )
}
