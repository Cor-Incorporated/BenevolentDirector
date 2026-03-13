import { sourceAuthorityLabels, type Citation } from '@/types/estimate'

type CitationListProps = {
  citations?: Citation[]
}

function getCitationTitle(citation: Citation, index: number) {
  return citation.title?.trim() || `Evidence source ${index + 1}`
}

export function CitationList({ citations = [] }: CitationListProps) {
  if (citations.length === 0) {
    return (
      <p className="text-sm text-slate-500">
        No evidence citations were returned for this estimate.
      </p>
    )
  }

  return (
    <div className="space-y-3">
      {citations.map((citation, index) => {
        const title = getCitationTitle(citation, index)
        const authority =
          citation.source_authority &&
          sourceAuthorityLabels[citation.source_authority]

        return (
          <details
            key={`${citation.url ?? title}-${index}`}
            className="rounded-xl border border-slate-200 bg-slate-50 p-4"
          >
            <summary className="cursor-pointer list-none">
              <div className="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
                <div className="space-y-1">
                  <p className="text-sm font-medium text-slate-950">{title}</p>
                  {citation.url &&
                  /^https?:\/\//i.test(citation.url) ? (
                    <a
                      href={citation.url}
                      target="_blank"
                      rel="noreferrer"
                      className="text-sm text-slate-600 underline underline-offset-2"
                    >
                      Open source
                    </a>
                  ) : (
                    <p className="text-sm text-slate-500">
                      Source URL unavailable
                    </p>
                  )}
                </div>

                <span className="inline-flex items-center self-start rounded-full border border-slate-200 bg-white px-2.5 py-1 text-xs font-medium text-slate-600">
                  {authority ?? 'Authority unavailable'}
                </span>
              </div>
            </summary>

            {citation.snippet ? (
              <p className="mt-3 text-sm text-pretty text-slate-600">
                {citation.snippet}
              </p>
            ) : (
              <p className="mt-3 text-sm text-slate-500">
                No supporting snippet was included.
              </p>
            )}
          </details>
        )
      })}
    </div>
  )
}
