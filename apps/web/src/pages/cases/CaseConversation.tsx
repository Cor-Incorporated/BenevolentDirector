import { useParams } from 'react-router-dom'
import { HearingChatColumn } from '@/components/hearing/HearingChatColumn'
import { HearingLeftPanel } from '@/components/hearing/HearingLeftPanel'
import { HearingRightPanel } from '@/components/hearing/HearingRightPanel'
import { HearingShell } from '@/components/hearing/HearingShell'
import { useHearingSession } from '@/hooks/use-hearing-session'

export function CaseConversation() {
  const { caseId } = useParams<{ caseId: string }>()
  const session = useHearingSession(caseId)

  if (!caseId) {
    return (
      <div className="flex min-h-[50dvh] items-center justify-center rounded-3xl border border-slate-200 bg-white px-6 text-slate-500 shadow-sm">
        <p>No case ID provided.</p>
      </div>
    )
  }

  return (
    <HearingShell
      caseId={caseId}
      error={session.error}
      leftPanel={(
        <HearingLeftPanel
          sourceDocuments={session.sourceDocuments}
          isUploading={session.isUploading}
          uploadError={session.uploadError}
          uploadNotice={session.uploadNotice}
          onUploadFile={session.uploadFile}
          onUploadUrl={session.uploadFromUrl}
        />
      )}
      chatColumn={(
        <HearingChatColumn
          turns={session.turns}
          streamingContent={session.streamingContent}
          isStreaming={session.isStreaming}
          isLoading={session.isLoading}
          prompts={session.missingInfoPrompts}
          onSend={session.sendMessage}
        />
      )}
      rightPanel={(
        <HearingRightPanel
          completeness={session.completeness}
          checklist={session.checklist}
          prompts={session.missingInfoPrompts}
          artifact={session.requirementArtifact}
          isRefreshingObservations={session.isRefreshingObservations}
          isRefreshingArtifact={session.isRefreshingArtifact}
        />
      )}
    />
  )
}
