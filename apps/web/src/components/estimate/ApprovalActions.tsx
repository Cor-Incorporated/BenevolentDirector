import { useCallback, useEffect, useState } from 'react'
import { cn } from '@/lib/cn'
import {
  approveProposal,
  createProposal,
  getApiErrorMessage,
  listProposals,
  rejectProposal,
} from '@/lib/api-client'
import {
  proposalStatusLabels,
  type ApprovalDecision,
  type ProposalSession,
} from '@/types/estimate'

type ApprovalActionsProps = {
  caseId: string
  estimateId: string
  className?: string
  onDecision?: (decision: ApprovalDecision) => void
}

function findEstimateProposal(
  proposals: ProposalSession[],
  estimateId: string,
) {
  return proposals.find((proposal) => proposal.estimate_id === estimateId) ?? null
}

export function ApprovalActions({
  caseId,
  estimateId,
  className,
  onDecision,
}: ApprovalActionsProps) {
  const [proposal, setProposal] = useState<ProposalSession | null>(null)
  const [note, setNote] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const loadProposal = useCallback(async () => {
    setIsLoading(true)

    try {
      const proposals = await listProposals(caseId)
      setProposal(findEstimateProposal(proposals, estimateId))
      setError(null)
    } catch (nextError) {
      setError(getApiErrorMessage(nextError, 'Unable to load approval status.'))
    } finally {
      setIsLoading(false)
    }
  }, [caseId, estimateId])

  useEffect(() => {
    void loadProposal()
  }, [loadProposal])

  const ensureProposal = useCallback(async () => {
    if (proposal) {
      return proposal
    }

    const createdProposal = await createProposal(caseId, estimateId)

    if (!createdProposal?.id) {
      throw new Error('The API response did not include a proposal id.')
    }

    setProposal(createdProposal)
    setNotice('Approval session prepared.')
    return createdProposal
  }, [caseId, estimateId, proposal])

  async function handlePrepareProposal() {
    setIsSubmitting(true)
    setError(null)
    setNotice(null)

    try {
      await ensureProposal()
    } catch (nextError) {
      setError(getApiErrorMessage(nextError, 'Unable to prepare proposal.'))
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleApprove() {
    setIsSubmitting(true)
    setError(null)
    setNotice(null)

    try {
      const activeProposal = await ensureProposal()
      const decision = await approveProposal(caseId, activeProposal.id, note.trim())

      setProposal((currentProposal) =>
        currentProposal
          ? {
              ...currentProposal,
              status: 'approved',
            }
          : currentProposal,
      )
      setNotice('Proposal approved.')

      if (decision) {
        onDecision?.(decision)
      }
    } catch (nextError) {
      setError(getApiErrorMessage(nextError, 'Unable to approve proposal.'))
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleReject() {
    const reason = note.trim()

    if (!reason) {
      setError('Enter a rejection reason before rejecting this proposal.')
      return
    }

    setIsSubmitting(true)
    setError(null)
    setNotice(null)

    try {
      const activeProposal = await ensureProposal()
      const decision = await rejectProposal(caseId, activeProposal.id, reason)

      setProposal((currentProposal) =>
        currentProposal
          ? {
              ...currentProposal,
              status: 'rejected',
            }
          : currentProposal,
      )
      setNotice('Proposal rejected.')

      if (decision) {
        onDecision?.(decision)
      }
    } catch (nextError) {
      setError(getApiErrorMessage(nextError, 'Unable to reject proposal.'))
    } finally {
      setIsSubmitting(false)
    }
  }

  const isDecisionLocked =
    proposal?.status === 'approved' ||
    proposal?.status === 'rejected' ||
    proposal?.status === 'expired'

  return (
    <section
      className={cn(
        'rounded-2xl border border-slate-200 bg-white p-6 shadow-sm',
        className,
      )}
    >
      <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
        <div className="space-y-2">
          <p className="text-sm font-medium text-slate-500">Approval</p>
          <h2 className="text-balance text-2xl font-semibold text-slate-950">
            Approval actions
          </h2>
          <p className="max-w-2xl text-pretty text-sm text-slate-600">
            Prepare a proposal session, then approve or reject this estimate
            with a decision note.
          </p>
        </div>

        {proposal ? (
          <span className="inline-flex items-center rounded-full border border-slate-200 bg-slate-100 px-3 py-1 text-sm font-medium text-slate-700">
            {proposalStatusLabels[proposal.status]}
          </span>
        ) : null}
      </div>

      {error ? (
        <p className="mt-4 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
          {error}
        </p>
      ) : null}

      {notice ? (
        <p className="mt-4 rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
          {notice}
        </p>
      ) : null}

      {isLoading ? (
        <p className="mt-6 text-sm text-slate-500">Loading approval status...</p>
      ) : (
        <div className="mt-6 space-y-4">
          {!proposal ? (
            <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
              <p className="text-sm text-slate-600">
                No approval session exists yet for this estimate.
              </p>
              <button
                type="button"
                onClick={handlePrepareProposal}
                disabled={isSubmitting}
                className="mt-4 inline-flex items-center rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-slate-700 disabled:cursor-not-allowed disabled:bg-slate-300"
              >
                {isSubmitting ? 'Preparing...' : 'Prepare approval session'}
              </button>
            </div>
          ) : null}

          <label className="block space-y-2 text-sm font-medium text-slate-700">
            <span>Decision note</span>
            <textarea
              value={note}
              onChange={(event) => setNote(event.target.value)}
              disabled={isSubmitting || isDecisionLocked}
              rows={4}
              className="block w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 shadow-sm focus:border-slate-500 focus:outline-none disabled:cursor-not-allowed disabled:bg-slate-100"
              placeholder="Summarize the rationale for approval or rejection."
            />
          </label>

          <div className="flex flex-col gap-3 sm:flex-row">
            <button
              type="button"
              onClick={handleApprove}
              disabled={isSubmitting || isDecisionLocked}
              className="inline-flex items-center justify-center rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-slate-700 disabled:cursor-not-allowed disabled:bg-slate-300"
            >
              {isSubmitting ? 'Saving...' : 'Approve proposal'}
            </button>

            <button
              type="button"
              onClick={handleReject}
              disabled={isSubmitting || isDecisionLocked}
              className="inline-flex items-center justify-center rounded-lg border border-rose-300 px-4 py-2 text-sm font-medium text-rose-700 transition-colors hover:border-rose-400 hover:bg-rose-50 disabled:cursor-not-allowed disabled:border-slate-200 disabled:text-slate-400"
            >
              {isSubmitting ? 'Saving...' : 'Reject proposal'}
            </button>
          </div>
        </div>
      )}
    </section>
  )
}
