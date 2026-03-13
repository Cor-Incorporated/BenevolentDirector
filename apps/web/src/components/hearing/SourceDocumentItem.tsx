import { formatDateTime } from '@/lib/api-client'
import { cn } from '@/lib/cn'
import type { SourceDocument } from '@/types/conversation'

interface SourceDocumentItemProps {
  document: SourceDocument
}

function formatFileSize(size?: number) {
  if (!size) {
    return null
  }

  if (size >= 1024 * 1024) {
    return `${(size / (1024 * 1024)).toFixed(1)} MB`
  }

  if (size >= 1024) {
    return `${Math.round(size / 1024)} KB`
  }

  return `${size} B`
}

const statusStyles: Record<SourceDocument['status'], string> = {
  pending: 'bg-amber-100 text-amber-700',
  processing: 'bg-blue-100 text-blue-700',
  completed: 'bg-emerald-100 text-emerald-700',
  failed: 'bg-rose-100 text-rose-700',
}

export function SourceDocumentItem({ document }: SourceDocumentItemProps) {
  const meta = [document.file_type, formatFileSize(document.file_size)]
    .filter(Boolean)
    .join(' • ')

  return (
    <li className="rounded-xl border border-slate-200 bg-white px-4 py-3">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="truncate text-sm font-medium text-slate-900">
            {document.file_name}
          </p>
          {meta ? (
            <p className="mt-1 text-xs text-slate-500">{meta}</p>
          ) : null}
          {document.source_url ? (
            /^https?:\/\//.test(document.source_url) ? (
              <a
                href={document.source_url}
                target="_blank"
                rel="noreferrer"
                className="mt-2 inline-flex text-xs text-blue-700 hover:text-blue-800"
              >
                {document.source_url}
              </a>
            ) : (
              <p className="mt-2 text-xs text-slate-500">
                {document.source_url}
              </p>
            )
          ) : null}
        </div>
        <span
          className={cn(
            'rounded-full px-2 py-0.5 text-[11px] font-medium capitalize',
            statusStyles[document.status],
          )}
        >
          {document.status}
        </span>
      </div>
      <p className="mt-3 text-xs text-slate-500">
        Added {formatDateTime(document.created_at)}
      </p>
    </li>
  )
}
