import createClient from 'openapi-fetch'
import type { components, paths } from '@/types/api'
import type {
  ConversationTurn,
  DataClassification,
  GetRequirementArtifactResponse,
  ListObservationQAPairsResponse,
  ListSourceDocumentsResponse,
  NDJSONChunk,
  ObservationQAPair,
  RequirementArtifact,
  SourceDocument,
  UploadSourceDocumentResponse,
} from '@/types/conversation'

export type CaseRecord = components['schemas']['Case']
export type CaseDetailRecord = components['schemas']['CaseWithDetails']
export type CaseStatus = components['schemas']['CaseStatus']
export type CaseType = components['schemas']['CaseType']

// WARNING: dev-only stub — must be replaced before production (ADR-0003)
// In production, tenant ID must come from Firebase Auth token claims.
const DEV_TENANT_ID = '11111111-1111-1111-1111-111111111111'
export const DEFAULT_TENANT_ID =
  import.meta.env.VITE_TENANT_ID ?? DEV_TENANT_ID
export const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
export const INTERNAL_DATA_CLASSIFICATION: DataClassification = 'internal'

export const caseTypeOptions: CaseType[] = [
  'new_project',
  'bug_report',
  'fix_request',
  'feature_addition',
  'undetermined',
]

export const caseStatusOptions: CaseStatus[] = [
  'draft',
  'interviewing',
  'analyzing',
  'estimating',
  'proposed',
  'approved',
  'rejected',
  'on_hold',
]

export const caseTypeLabels: Record<CaseType, string> = {
  new_project: 'New project',
  bug_report: 'Bug report',
  fix_request: 'Fix request',
  feature_addition: 'Feature addition',
  undetermined: 'Undetermined',
}

export const caseStatusLabels: Record<CaseStatus, string> = {
  draft: 'Draft',
  interviewing: 'Interviewing',
  analyzing: 'Analyzing',
  estimating: 'Estimating',
  proposed: 'Proposed',
  approved: 'Approved',
  rejected: 'Rejected',
  on_hold: 'On hold',
}

export const apiClient = createClient<paths>({
  baseUrl: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
    'X-Tenant-ID': DEFAULT_TENANT_ID,
  },
})

type ApiErrorPayload = {
  error?: {
    message?: string
  }
}

function isApiErrorPayload(value: unknown): value is ApiErrorPayload {
  return typeof value === 'object' && value !== null && 'error' in value
}

// Builds common headers for all API requests. X-Data-Classification: internal
// is intentionally attached to every request (including read-only endpoints
// like listConversationTurns) so the control-api can enforce data-handling
// policies uniformly. The control-api ignores unrecognised headers, so this
// addition is backwards-compatible.
function buildHeaders(
  overrides?: Record<string, string>,
): Record<string, string> {
  return {
    'X-Tenant-ID': DEFAULT_TENANT_ID,
    'X-Data-Classification': INTERNAL_DATA_CLASSIFICATION,
    ...overrides,
  }
}

async function parseJsonResponse<T>(response: Response): Promise<T | null> {
  try {
    return (await response.json()) as T
  } catch {
    return null
  }
}

export function getApiErrorMessage(
  error: unknown,
  fallback = 'Something went wrong. Please try again.',
) {
  if (isApiErrorPayload(error) && typeof error.error?.message === 'string') {
    return error.error.message
  }

  if (error instanceof Error && error.message) {
    return error.message
  }

  return fallback
}

// ─── Conversation API helpers ───────────────────────────────
// TODO: Auth headers will be added when Firebase Auth is integrated (ADR-0003).
// Currently listConversationTurns and streamMessage send X-Tenant-ID only.

export async function listConversationTurns(
  caseId: string,
): Promise<ConversationTurn[]> {
  const res = await fetch(
    `${API_BASE_URL}/v1/cases/${encodeURIComponent(caseId)}/conversations`,
    {
      headers: buildHeaders({ 'Content-Type': 'application/json' }),
    },
  )

  if (!res.ok) {
    throw new Error(`API error ${res.status}`)
  }

  const json: unknown = await res.json()
  if (
    typeof json !== 'object' ||
    json === null ||
    !('data' in json) ||
    !Array.isArray((json as Record<string, unknown>).data)
  ) {
    throw new Error('Unexpected API response shape')
  }
  return (json as { data: ConversationTurn[] }).data
}

export async function listSourceDocuments(
  caseId: string,
): Promise<SourceDocument[]> {
  const res = await fetch(
    `${API_BASE_URL}/v1/cases/${encodeURIComponent(caseId)}/source-documents`,
    {
      headers: buildHeaders({ 'Content-Type': 'application/json' }),
    },
  )

  if (!res.ok) {
    const body = await parseJsonResponse<unknown>(res)
    throw new Error(getApiErrorMessage(body, `API error ${res.status}`))
  }

  const json = await parseJsonResponse<ListSourceDocumentsResponse>(res)
  return json?.data ?? []
}

export async function getRequirementArtifact(
  caseId: string,
): Promise<RequirementArtifact | null> {
  const res = await fetch(
    `${API_BASE_URL}/v1/cases/${encodeURIComponent(caseId)}/requirement-artifact`,
    {
      headers: buildHeaders({ 'Content-Type': 'application/json' }),
    },
  )

  if (res.status === 404) {
    return null
  }

  if (!res.ok) {
    const body = await parseJsonResponse<unknown>(res)
    throw new Error(getApiErrorMessage(body, `API error ${res.status}`))
  }

  const json = await parseJsonResponse<GetRequirementArtifactResponse>(res)
  return json?.data ?? null
}

export async function uploadSourceDocument(
  caseId: string,
  input: { file?: File; sourceUrl?: string },
): Promise<SourceDocument | null> {
  if (!input.file && !input.sourceUrl) {
    throw new Error('Provide a file or source URL to upload.')
  }

  const formData = new FormData()
  if (input.file) {
    formData.append('file', input.file)
  }
  if (input.sourceUrl) {
    formData.append('source_url', input.sourceUrl)
  }

  const res = await fetch(
    `${API_BASE_URL}/v1/cases/${encodeURIComponent(caseId)}/source-documents`,
    {
      method: 'POST',
      headers: buildHeaders(),
      body: formData,
    },
  )

  if (!res.ok) {
    const body = await parseJsonResponse<unknown>(res)
    throw new Error(getApiErrorMessage(body, `API error ${res.status}`))
  }

  const json = await parseJsonResponse<UploadSourceDocumentResponse>(res)
  return json?.data ?? null
}

export async function listObservationQAPairs(
  caseId: string,
): Promise<ObservationQAPair[]> {
  const res = await fetch(
    `${API_BASE_URL}/v1/cases/${encodeURIComponent(caseId)}/observation/qa-pairs`,
    {
      headers: buildHeaders({ 'Content-Type': 'application/json' }),
    },
  )

  if (!res.ok) {
    const body = await parseJsonResponse<unknown>(res)
    throw new Error(getApiErrorMessage(body, `API error ${res.status}`))
  }

  const json = await parseJsonResponse<ListObservationQAPairsResponse>(res)
  return json?.data ?? []
}

export async function* streamMessage(
  caseId: string,
  content: string,
  signal?: AbortSignal,
): AsyncGenerator<NDJSONChunk> {
  const res = await fetch(
    `${API_BASE_URL}/v1/cases/${encodeURIComponent(caseId)}/conversations/stream`,
    {
      method: 'POST',
      headers: buildHeaders({
        Accept: 'application/x-ndjson',
        'Content-Type': 'application/json',
      }),
      body: JSON.stringify({ content }),
      ...(signal ? { signal } : {}),
    },
  )

  if (!res.ok) {
    throw new Error(`API error ${res.status}`)
  }

  const reader = res.body?.getReader()
  if (!reader) {
    throw new Error('No response body')
  }

  const decoder = new TextDecoder()
  let buffer = ''

  try {
    for (;;) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })

      const lines = buffer.split('\n')
      buffer = lines.pop() ?? ''

      for (const line of lines) {
        const trimmed = line.trim()
        if (trimmed === '') continue
        try {
          yield JSON.parse(trimmed) as NDJSONChunk
        } catch {
          yield { type: 'error', error: `Malformed NDJSON: ${trimmed.slice(0, 100)}` } as NDJSONChunk
        }
      }
    }

    // Flush any remaining multi-byte UTF-8 sequences from the decoder
    buffer += decoder.decode()

    if (buffer.trim() !== '') {
      try {
        yield JSON.parse(buffer.trim()) as NDJSONChunk
      } catch {
        yield { type: 'error', error: `Malformed NDJSON: ${buffer.trim().slice(0, 100)}` } as NDJSONChunk
      }
    }
  } finally {
    reader.releaseLock()
  }
}

export function formatDateTime(value?: string) {
  if (!value) {
    return 'Not available'
  }

  return new Intl.DateTimeFormat('ja-JP', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}
