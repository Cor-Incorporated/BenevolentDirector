import { beforeEach, describe, expect, it } from 'vitest'
import {
  approveProposal,
  caseStatusLabels,
  caseStatusOptions,
  caseTypeLabels,
  caseTypeOptions,
  createEstimate,
  createProposal,
  formatDateTime,
  getEstimate,
  getApiErrorMessage,
  getRequirementArtifact,
  INTERNAL_DATA_CLASSIFICATION,
  listEstimates,
  listObservationQAPairs,
  listProposals,
  rejectProposal,
  getThreeWayProposal,
  listSourceDocuments,
  uploadSourceDocument,
} from './api-client'

beforeEach(() => {
  vi.restoreAllMocks()
})

describe('formatDateTime', () => {
  it('formats a valid ISO date string', () => {
    const result = formatDateTime('2026-03-12T09:00:00Z')
    expect(typeof result).toBe('string')
    expect(result).not.toBe('Not available')
    expect(result.length).toBeGreaterThan(0)
  })

  it('returns "Not available" for undefined', () => {
    expect(formatDateTime(undefined)).toBe('Not available')
  })

  it('returns "Not available" for empty string', () => {
    expect(formatDateTime('')).toBe('Not available')
  })
})

describe('getApiErrorMessage', () => {
  it('extracts message from API error payload', () => {
    const error = { error: { message: 'Case not found' } }
    expect(getApiErrorMessage(error)).toBe('Case not found')
  })

  it('extracts message from Error instance', () => {
    const error = new Error('Network failure')
    expect(getApiErrorMessage(error)).toBe('Network failure')
  })

  it('returns fallback for non-object errors', () => {
    expect(getApiErrorMessage(null)).toBe(
      'Something went wrong. Please try again.',
    )
  })

  it('returns custom fallback when provided', () => {
    expect(getApiErrorMessage(42, 'Custom fallback')).toBe('Custom fallback')
  })

  it('returns fallback for API payload without message', () => {
    expect(getApiErrorMessage({ error: {} })).toBe(
      'Something went wrong. Please try again.',
    )
  })
})

describe('caseStatusLabels', () => {
  it('has labels for all expected statuses', () => {
    expect(caseStatusLabels.draft).toBe('Draft')
    expect(caseStatusLabels.interviewing).toBe('Interviewing')
    expect(caseStatusLabels.approved).toBe('Approved')
    expect(caseStatusLabels.rejected).toBe('Rejected')
    expect(caseStatusLabels.on_hold).toBe('On hold')
  })
})

describe('caseTypeLabels', () => {
  it('has labels for all expected types', () => {
    expect(caseTypeLabels.new_project).toBe('New project')
    expect(caseTypeLabels.bug_report).toBe('Bug report')
    expect(caseTypeLabels.undetermined).toBe('Undetermined')
  })
})

describe('caseStatusOptions', () => {
  it('contains all status values', () => {
    expect(caseStatusOptions).toContain('draft')
    expect(caseStatusOptions).toContain('approved')
    expect(caseStatusOptions).toContain('on_hold')
    expect(caseStatusOptions).toHaveLength(8)
  })
})

describe('caseTypeOptions', () => {
  it('contains all type values', () => {
    expect(caseTypeOptions).toContain('new_project')
    expect(caseTypeOptions).toContain('bug_report')
    expect(caseTypeOptions).toContain('undetermined')
    expect(caseTypeOptions).toHaveLength(5)
  })
})

describe('hearing API helpers', () => {
  it('lists source documents with the internal classification header', async () => {
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(
        JSON.stringify({
          data: [
            {
              id: 'doc-1',
              case_id: 'case-1',
              file_name: 'brief.pdf',
              status: 'completed',
              created_at: '2026-03-13T09:00:00Z',
            },
          ],
          total: 1,
        }),
        {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        },
      ),
    )

    const documents = await listSourceDocuments('case-1')

    expect(documents).toHaveLength(1)
    expect(fetchSpy).toHaveBeenCalledWith(
      expect.stringContaining('/v1/cases/case-1/source-documents'),
      expect.objectContaining({
        headers: expect.objectContaining({
          'X-Data-Classification': INTERNAL_DATA_CLASSIFICATION,
        }),
      }),
    )
  })

  it('returns null when the requirement artifact is not ready yet', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(null, { status: 404 }),
    )

    await expect(getRequirementArtifact('case-1')).resolves.toBeNull()
  })

  it('uploads a source document as multipart form data', async () => {
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            id: 'doc-1',
            case_id: 'case-1',
            file_name: 'brief.pdf',
            status: 'pending',
            created_at: '2026-03-13T09:00:00Z',
          },
          job_id: 'job-1',
        }),
        {
          status: 202,
          headers: { 'Content-Type': 'application/json' },
        },
      ),
    )

    const file = new File(['hello'], 'brief.pdf', { type: 'application/pdf' })
    await uploadSourceDocument('case-1', { file })

    expect(fetchSpy).toHaveBeenCalledWith(
      expect.stringContaining('/v1/cases/case-1/source-documents'),
      expect.objectContaining({
        method: 'POST',
        body: expect.any(FormData),
        headers: expect.objectContaining({
          'X-Data-Classification': INTERNAL_DATA_CLASSIFICATION,
        }),
      }),
    )
  })

  it('lists observation QA pairs for completeness refresh', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(
        JSON.stringify({
          data: [
            {
              id: 'qa-1',
              case_id: 'case-1',
              session_id: 'session-1',
              question_text: 'What is the launch window?',
              answer_text: 'Q3 2026',
              quality: {
                completeness: 0.75,
                coherence: 0.9,
                rationale: 'Timing exists but needs more detail.',
                needs_followup: true,
                is_complete: false,
              },
              created_at: '2026-03-13T09:00:00Z',
            },
          ],
          total: 1,
        }),
        {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        },
      ),
    )

    await expect(listObservationQAPairs('case-1')).resolves.toEqual([
      expect.objectContaining({
        id: 'qa-1',
        question_text: 'What is the launch window?',
      }),
    ])
  })
})

describe('estimate API helpers', () => {
  it('lists estimates for a case', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(
        JSON.stringify({
          data: [
            {
              id: 'estimate-1',
              case_id: 'case-1',
              estimate_mode: 'hybrid',
              status: 'ready',
            },
          ],
          total: 1,
        }),
        {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        },
      ),
    )

    await expect(listEstimates('case-1')).resolves.toEqual({
      data: [
        expect.objectContaining({
          id: 'estimate-1',
          estimate_mode: 'hybrid',
        }),
      ],
      total: 1,
    })
  })

  it('creates and fetches estimate detail', async () => {
    const fetchSpy = vi
      .spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              id: 'estimate-1',
              case_id: 'case-1',
              estimate_mode: 'hybrid',
              status: 'draft',
            },
          }),
          {
            status: 201,
            headers: { 'Content-Type': 'application/json' },
          },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              id: 'estimate-1',
              case_id: 'case-1',
              estimate_mode: 'hybrid',
              status: 'ready',
              three_way_proposal: {
                market_benchmark: { confidence: 'high' },
              },
            },
          }),
          {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          },
        ),
      )

    await expect(
      createEstimate('case-1', {
        your_hourly_rate: 12000,
        region: 'japan',
        include_market_evidence: true,
      }),
    ).resolves.toEqual(
      expect.objectContaining({
        id: 'estimate-1',
      }),
    )

    await expect(getEstimate('case-1', 'estimate-1')).resolves.toEqual(
      expect.objectContaining({
        status: 'ready',
      }),
    )

    expect(fetchSpy).toHaveBeenNthCalledWith(
      1,
      expect.stringContaining('/v1/cases/case-1/estimates'),
      expect.objectContaining({
        method: 'POST',
      }),
    )
  })

  it('handles proposal and approval endpoints', async () => {
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: [
              {
                id: 'proposal-1',
                case_id: 'case-1',
                estimate_id: 'estimate-1',
                status: 'draft',
              },
            ],
            total: 1,
          }),
          {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              id: 'proposal-1',
              case_id: 'case-1',
              estimate_id: 'estimate-1',
              status: 'draft',
            },
          }),
          {
            status: 201,
            headers: { 'Content-Type': 'application/json' },
          },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              market_benchmark: { confidence: 'high' },
            },
          }),
          {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              decision: 'approved',
            },
          }),
          {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              decision: 'rejected',
            },
          }),
          {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          },
        ),
      )

    await expect(listProposals('case-1')).resolves.toHaveLength(1)
    await expect(createProposal('case-1', 'estimate-1')).resolves.toEqual(
      expect.objectContaining({
        id: 'proposal-1',
      }),
    )
    await expect(
      getThreeWayProposal('case-1', 'estimate-1'),
    ).resolves.toEqual(
      expect.objectContaining({
        market_benchmark: expect.objectContaining({ confidence: 'high' }),
      }),
    )
    await expect(
      approveProposal('case-1', 'proposal-1', 'Looks good'),
    ).resolves.toEqual(expect.objectContaining({ decision: 'approved' }))
    await expect(
      rejectProposal('case-1', 'proposal-1', 'Need more evidence'),
    ).resolves.toEqual(expect.objectContaining({ decision: 'rejected' }))
  })
})
