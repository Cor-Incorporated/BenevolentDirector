import { formatDateTime } from '@/lib/api-client'
import type { RequirementArtifact } from '@/types/conversation'
import { SpecMarkdownViewer } from './SpecMarkdownViewer'

interface SpecPreviewPanelProps {
  artifact: RequirementArtifact | null
  isRefreshing?: boolean
}

export function SpecPreviewPanel({
  artifact,
  isRefreshing = false,
}: SpecPreviewPanelProps) {
  if (!artifact) {
    return (
      <section className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <div className="flex items-center justify-between gap-3">
          <h3 className="text-base font-semibold text-slate-950">
            Spec preview
          </h3>
          {isRefreshing ? (
            <span className="text-xs text-slate-500">Refreshing…</span>
          ) : null}
        </div>
        <p className="mt-3 text-pretty text-sm text-slate-600">
          No requirement artifact is available yet. Continue the hearing to
          generate the first draft specification.
        </p>
      </section>
    )
  }

  return (
    <section className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h3 className="text-base font-semibold text-slate-950">
            Spec preview
          </h3>
          <p className="mt-1 text-sm text-slate-500">
            Version {artifact.version} • {artifact.status}
          </p>
        </div>
        {isRefreshing ? (
          <span className="text-xs text-slate-500">Refreshing…</span>
        ) : null}
      </div>
      <p className="mt-3 text-xs text-slate-500">
        Updated {formatDateTime(artifact.updated_at ?? artifact.created_at)}
      </p>
      <div className="mt-4 max-h-[24rem] overflow-y-auto rounded-2xl bg-slate-50 p-4">
        <SpecMarkdownViewer markdown={artifact.markdown} />
      </div>
    </section>
  )
}
