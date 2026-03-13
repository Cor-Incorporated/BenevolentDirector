import { act, renderHook } from '@testing-library/react'
import { useHearingSession } from '@/hooks/use-hearing-session'
import { useNDJSONStream } from '@/hooks/use-ndjson-stream'
import {
  getRequirementArtifact,
  listConversationTurns,
  listObservationQAPairs,
  listSourceDocuments,
  uploadSourceDocument,
} from '@/lib/api-client'

vi.mock('@/hooks/use-ndjson-stream', () => ({
  useNDJSONStream: vi.fn(),
}))

vi.mock('@/lib/api-client', () => ({
  getApiErrorMessage: (error: unknown, fallback: string) =>
    error instanceof Error ? error.message : fallback,
  getRequirementArtifact: vi.fn(),
  listConversationTurns: vi.fn(),
  listObservationQAPairs: vi.fn(),
  listSourceDocuments: vi.fn(),
  uploadSourceDocument: vi.fn(),
}))

const mockedUseNDJSONStream = vi.mocked(useNDJSONStream)
const mockedListConversationTurns = vi.mocked(listConversationTurns)
const mockedListSourceDocuments = vi.mocked(listSourceDocuments)
const mockedGetRequirementArtifact = vi.mocked(getRequirementArtifact)
const mockedListObservationQAPairs = vi.mocked(listObservationQAPairs)
const mockedUploadSourceDocument = vi.mocked(uploadSourceDocument)

async function flushAsyncWork() {
  await act(async () => {
    await Promise.resolve()
  })
}

describe('useHearingSession', () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
    vi.clearAllMocks()
    mockedListObservationQAPairs.mockResolvedValue([])
    mockedUseNDJSONStream.mockReturnValue({
      streamingContent: '',
      isStreaming: false,
      error: null,
      sendStreamMessage: vi.fn(),
      cancelStream: vi.fn(),
    })
  })

  afterEach(() => {
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
  })

  it('loads the hearing session and hydrates the initial layout state', async () => {
    mockedListConversationTurns.mockResolvedValue([
      {
        id: 'turn-1',
        case_id: 'case-1',
        role: 'assistant',
        content: 'What problem are we solving?',
        created_at: '2026-03-13T00:00:00Z',
      },
    ])
    mockedListSourceDocuments.mockResolvedValue([
      {
        id: 'doc-1',
        case_id: 'case-1',
        file_name: 'brief.pdf',
        source_kind: 'file_upload',
        status: 'completed',
        created_at: '2026-03-13T00:00:00Z',
      },
    ])
    mockedGetRequirementArtifact.mockResolvedValue({
      id: 'artifact-1',
      case_id: 'case-1',
      version: 2,
      markdown: '# Draft spec',
      status: 'draft',
    })

    const { result } = renderHook(() => useHearingSession('case-1'))
    await flushAsyncWork()

    expect(result.current.turns).toHaveLength(1)
    expect(result.current.sourceDocuments).toHaveLength(1)
    expect(result.current.requirementArtifact?.version).toBe(2)
    expect(result.current.completeness).toBe(0)
  })

  it('refreshes conversations, QA pairs, and artifact data after a streamed turn completes', async () => {
    const sendStreamMessage = vi.fn().mockResolvedValue({
      id: 'turn-2',
      case_id: 'case-1',
      role: 'assistant',
      content: 'Tell me about your deployment constraints.',
      created_at: '2026-03-13T00:00:02Z',
    })
    mockedUseNDJSONStream.mockReturnValue({
      streamingContent: '',
      isStreaming: false,
      error: null,
      sendStreamMessage,
      cancelStream: vi.fn(),
    })

    mockedListConversationTurns
      .mockResolvedValueOnce([])
      .mockResolvedValueOnce([
        {
          id: 'turn-1',
          case_id: 'case-1',
          role: 'user',
          content: 'We need a new admin dashboard.',
          created_at: '2026-03-13T00:00:01Z',
        },
        {
          id: 'turn-2',
          case_id: 'case-1',
          role: 'assistant',
          content: 'Tell me about your deployment constraints.',
          metadata: { category: 'constraint' },
          created_at: '2026-03-13T00:00:02Z',
        },
      ])
    mockedListSourceDocuments.mockResolvedValue([])
    mockedGetRequirementArtifact
      .mockResolvedValueOnce(null)
      .mockResolvedValueOnce({
        id: 'artifact-1',
        case_id: 'case-1',
        version: 1,
        markdown: '# Draft spec',
        status: 'draft',
      })
    mockedListObservationQAPairs.mockResolvedValue([
      {
        id: 'qa-1',
        case_id: 'case-1',
        session_id: 'session-1',
        question_text: 'What deployment constraints matter?',
        answer_text: 'Must run in the customer VPC.',
        quality: {
          confidence: 0.92,
          completeness: 0.64,
          coherence: 0.88,
          needs_followup: true,
          is_complete: false,
          rationale: 'Hosting constraints need more detail.',
        },
        created_at: '2026-03-13T00:00:03Z',
      },
    ])

    const { result } = renderHook(() => useHearingSession('case-1'))
    await flushAsyncWork()

    await act(async () => {
      await result.current.sendMessage('We need a new admin dashboard.')
    })

    expect(sendStreamMessage).toHaveBeenCalledWith(
      'case-1',
      'We need a new admin dashboard.',
    )
    expect(result.current.turns).toHaveLength(2)

    await act(async () => {
      vi.advanceTimersByTime(500)
      await Promise.resolve()
    })

    expect(result.current.turns[1]?.metadata?.category).toBe('constraint')
    expect(result.current.completeness).toBeCloseTo(0.64)
    expect(result.current.missingInfoPrompts).toHaveLength(1)

    await act(async () => {
      vi.advanceTimersByTime(2_000)
      await Promise.resolve()
    })

    await flushAsyncWork()
    expect(result.current.requirementArtifact?.version).toBe(1)
  })

  it('uploads a source document and merges the queued item into state', async () => {
    mockedListConversationTurns.mockResolvedValue([])
    mockedListSourceDocuments
      .mockResolvedValueOnce([])
      .mockResolvedValueOnce([
        {
          id: 'doc-1',
          case_id: 'case-1',
          file_name: 'brief.pdf',
          source_kind: 'file_upload',
          status: 'pending',
          created_at: '2026-03-13T00:00:00Z',
        },
      ])
    mockedGetRequirementArtifact.mockResolvedValue(null)
    mockedUploadSourceDocument.mockResolvedValue({
      id: 'doc-1',
      case_id: 'case-1',
      file_name: 'brief.pdf',
      source_kind: 'file_upload',
      status: 'pending',
      created_at: '2026-03-13T00:00:00Z',
    })

    const { result } = renderHook(() => useHearingSession('case-1'))
    await flushAsyncWork()

    await act(async () => {
      await result.current.uploadFile(
        new File(['test'], 'brief.pdf', { type: 'application/pdf' }),
      )
    })

    expect(result.current.sourceDocuments).toHaveLength(1)
    expect(result.current.sourceDocuments[0]?.file_name).toBe('brief.pdf')

    await act(async () => {
      vi.advanceTimersByTime(1_500)
      await Promise.resolve()
    })

    expect(mockedListSourceDocuments).toHaveBeenCalledTimes(2)
  })
})
