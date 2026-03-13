import { render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { CaseConversation } from '../CaseConversation'

function renderWithRouter(caseId: string) {
  const router = createMemoryRouter(
    [{ path: '/cases/:caseId/conversation', element: <CaseConversation /> }],
    { initialEntries: [`/cases/${caseId}/conversation`] },
  )

  return render(<RouterProvider router={router} />)
}

function buildJsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

describe('CaseConversation', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    Element.prototype.scrollIntoView = vi.fn()
  })

  it('renders the hearing workspace with source context and artifact preview', async () => {
    vi.spyOn(globalThis, 'fetch').mockImplementation((input) => {
      const url = String(input)

      if (url.endsWith('/conversations')) {
        return Promise.resolve(buildJsonResponse({
          data: [
            {
              id: 'turn-1',
              case_id: 'test-case-123',
              role: 'assistant',
              content: 'Please describe the current workflow.',
              metadata: { category: 'scope' },
              created_at: '2026-03-13T00:00:00Z',
            },
          ],
          total: 1,
        }))
      }

      if (url.endsWith('/source-documents')) {
        return Promise.resolve(buildJsonResponse({
          data: [
            {
              id: 'doc-1',
              case_id: 'test-case-123',
              file_name: 'requirements.pdf',
              status: 'completed',
              source_kind: 'file_upload',
              created_at: '2026-03-13T00:00:00Z',
            },
          ],
          total: 1,
        }))
      }

      if (url.endsWith('/requirement-artifact')) {
        return Promise.resolve(buildJsonResponse({
          data: {
            id: 'artifact-1',
            case_id: 'test-case-123',
            version: 3,
            markdown: '# Scope\n- Intake interview\n- Spec refresh',
            status: 'draft',
          },
        }))
      }

      if (url.endsWith('/observation/qa-pairs')) {
        return Promise.resolve(buildJsonResponse({
          data: [],
          total: 0,
        }))
      }

      return Promise.reject(new Error(`Unexpected fetch: ${url}`))
    })

    renderWithRouter('test-case-123')

    expect(await screen.findByText('Three-panel intake workspace')).toBeInTheDocument()
    expect(screen.getByText('test-case-123')).toBeInTheDocument()
    expect(screen.getByText('Please describe the current workflow.')).toBeInTheDocument()
    expect(screen.getByText('requirements.pdf')).toBeInTheDocument()
    expect(screen.getByText('Version 3 • draft')).toBeInTheDocument()
    expect(screen.getAllByText('Scope')).toHaveLength(2)
    expect(screen.getByText('Completeness scoring will appear after the assistant finishes a turn.')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Choose file' })).toBeInTheDocument()
  })
})
