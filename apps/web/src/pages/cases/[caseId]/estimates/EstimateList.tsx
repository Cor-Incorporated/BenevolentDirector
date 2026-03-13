import { Link, useParams } from 'react-router-dom'
import { useEstimates } from '@/hooks/use-estimates'
import { formatDateTime } from '@/lib/api-client'
import {
  estimateModeLabels,
  estimateStatusLabels,
  formatEstimateCurrency,
  formatEstimateHours,
} from '@/types/estimate'

export function EstimateList() {
  const { caseId } = useParams<{ caseId: string }>()
  const { estimates, total, loading, error } = useEstimates(caseId)
  const readyCount = estimates.filter((estimate) => estimate.status === 'ready').length
  const approvedCount = estimates.filter(
    (estimate) => estimate.status === 'approved',
  ).length

  if (!caseId) {
    return (
      <main className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <p className="text-sm text-slate-600">Case id is missing from the route.</p>
      </main>
    )
  }

  return (
    <main className="space-y-6">
      <header className="flex flex-col gap-4 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm lg:flex-row lg:items-end lg:justify-between">
        <div className="space-y-2">
          <p className="text-sm font-medium text-slate-500">Estimates</p>
          <h1 className="text-balance text-3xl font-semibold text-slate-950">
            Estimate workspace
          </h1>
          <p className="max-w-2xl text-pretty text-sm text-slate-600">
            Review generated estimates, compare readiness, and open the detailed
            proposal for each option.
          </p>
        </div>

        <div className="flex flex-col gap-3 sm:flex-row">
          <Link
            to={`/cases/${caseId}`}
            className="inline-flex items-center justify-center rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 transition-colors hover:border-slate-400 hover:text-slate-950"
          >
            Back to case
          </Link>
          <Link
            to={`/cases/${caseId}/estimates/new`}
            className="inline-flex items-center justify-center rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-slate-700"
          >
            Create estimate
          </Link>
        </div>
      </header>

      <section className="grid gap-4 md:grid-cols-3">
        <article className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
          <p className="text-sm font-medium text-slate-500">Total estimates</p>
          <p className="mt-2 text-3xl font-semibold text-slate-950 tabular-nums">
            {total}
          </p>
        </article>

        <article className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
          <p className="text-sm font-medium text-slate-500">Ready to review</p>
          <p className="mt-2 text-3xl font-semibold text-slate-950 tabular-nums">
            {readyCount}
          </p>
        </article>

        <article className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
          <p className="text-sm font-medium text-slate-500">Approved</p>
          <p className="mt-2 text-3xl font-semibold text-slate-950 tabular-nums">
            {approvedCount}
          </p>
        </article>
      </section>

      <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
        <div className="flex items-center justify-between border-b border-slate-200 px-6 py-4">
          <div>
            <h2 className="text-lg font-semibold text-slate-950">
              Estimate list
            </h2>
            <p className="mt-1 text-sm text-slate-500">
              <span className="tabular-nums">{total}</span> estimates captured
            </p>
          </div>
        </div>

        {error ? (
          <div className="border-b border-rose-200 bg-rose-50 px-6 py-4 text-sm text-rose-700">
            {error}
          </div>
        ) : null}

        {loading ? (
          <div className="px-6 py-12 text-sm text-slate-500">
            Loading estimates...
          </div>
        ) : estimates.length === 0 ? (
          <div className="px-6 py-12">
            <p className="text-lg font-semibold text-slate-950">
              No estimates yet
            </p>
            <p className="mt-2 max-w-xl text-sm text-pretty text-slate-600">
              Generate the first estimate to compare market evidence, internal
              effort, and approval readiness.
            </p>
            <Link
              to={`/cases/${caseId}/estimates/new`}
              className="mt-4 inline-flex items-center rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-slate-700"
            >
              Create first estimate
            </Link>
          </div>
        ) : (
          <div className="divide-y divide-slate-200">
            {estimates.map((estimate) => (
              <article
                key={estimate.id}
                className="flex flex-col gap-4 px-6 py-5 lg:flex-row lg:items-center lg:justify-between"
              >
                <div className="space-y-3">
                  <div className="flex flex-wrap gap-2">
                    <span className="rounded-full bg-slate-100 px-3 py-1 text-sm text-slate-700">
                      {estimateModeLabels[estimate.estimate_mode]}
                    </span>
                    <span className="rounded-full bg-slate-100 px-3 py-1 text-sm text-slate-700">
                      {estimateStatusLabels[estimate.status]}
                    </span>
                    {(estimate.risk_flags ?? []).slice(0, 2).map((flag) => (
                      <span
                        key={flag}
                        className="rounded-full bg-amber-50 px-3 py-1 text-sm text-amber-700"
                      >
                        {flag}
                      </span>
                    ))}
                  </div>

                  <div className="grid gap-3 sm:grid-cols-3">
                    <div>
                      <p className="text-sm font-medium text-slate-500">Our total</p>
                      <p className="mt-1 text-sm font-semibold text-slate-950 tabular-nums">
                        {formatEstimateCurrency(estimate.total_your_cost)}
                      </p>
                    </div>
                    <div>
                      <p className="text-sm font-medium text-slate-500">Hours</p>
                      <p className="mt-1 text-sm font-semibold text-slate-950 tabular-nums">
                        {formatEstimateHours(estimate.your_estimated_hours)}
                      </p>
                    </div>
                    <div>
                      <p className="text-sm font-medium text-slate-500">
                        Created
                      </p>
                      <p className="mt-1 text-sm font-semibold text-slate-950 tabular-nums">
                        {formatDateTime(estimate.created_at)}
                      </p>
                    </div>
                  </div>
                </div>

                <Link
                  to={`/cases/${caseId}/estimates/${estimate.id}`}
                  className="inline-flex items-center justify-center rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 transition-colors hover:border-slate-400 hover:text-slate-950"
                >
                  Open detail
                </Link>
              </article>
            ))}
          </div>
        )}
      </section>
    </main>
  )
}
