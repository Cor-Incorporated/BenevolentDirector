import type { SourceDocument } from '@/types/conversation'
import { SourceDocumentItem } from './SourceDocumentItem'
import { SourceDocumentUploadArea } from './SourceDocumentUploadArea'

interface HearingLeftPanelProps {
  sourceDocuments: SourceDocument[]
  isUploading: boolean
  uploadError?: string | null
  uploadNotice?: string | null
  onUploadFile: (file: File) => Promise<void>
  onUploadUrl: (sourceUrl: string) => Promise<void>
}

export function HearingLeftPanel({
  sourceDocuments,
  isUploading,
  uploadError,
  uploadNotice,
  onUploadFile,
  onUploadUrl,
}: HearingLeftPanelProps) {
  return (
    <aside className="space-y-4">
      <section className="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
        <p className="text-sm font-medium text-slate-500">Context</p>
        <h2 className="mt-1 text-balance text-lg font-semibold text-slate-950">
          Source documents
        </h2>
        <p className="mt-2 text-pretty text-sm text-slate-600">
          Keep background material close to the hearing so the assistant can
          ground follow-up questions and later spec drafts.
        </p>
      </section>

      <section className="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
        <SourceDocumentUploadArea
          isUploading={isUploading}
          error={uploadError}
          notice={uploadNotice}
          onUploadFile={onUploadFile}
          onUploadUrl={onUploadUrl}
        />
      </section>

      <section className="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
        <div className="flex items-center justify-between gap-3">
          <h3 className="text-sm font-medium text-slate-700">Uploaded</h3>
          <span className="text-sm font-semibold tabular-nums text-slate-950">
            {sourceDocuments.length}
          </span>
        </div>
        {sourceDocuments.length === 0 ? (
          <p className="mt-3 text-pretty text-sm text-slate-600">
            Upload the first document or queue a URL to give the hearing more
            reliable background context.
          </p>
        ) : (
          <ul className="mt-3 space-y-3">
            {sourceDocuments.map((document) => (
              <SourceDocumentItem key={document.id} document={document} />
            ))}
          </ul>
        )}
      </section>
    </aside>
  )
}
