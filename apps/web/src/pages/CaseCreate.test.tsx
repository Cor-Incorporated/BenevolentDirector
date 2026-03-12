import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { CaseCreate } from './CaseCreate'

const mockNavigate = vi.fn()
const mockPost = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>(
    'react-router-dom',
  )

  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

vi.mock('@/lib/api-client', () => ({
  apiClient: {
    POST: mockPost,
  },
  caseTypeLabels: {
    new_project: 'New project',
    bug_report: 'Bug report',
    fix_request: 'Fix request',
    feature_addition: 'Feature addition',
    undetermined: 'Undetermined',
  },
  caseTypeOptions: [
    'new_project',
    'bug_report',
    'fix_request',
    'feature_addition',
    'undetermined',
  ],
  getApiErrorMessage: () => 'Unable to create case.',
}))

describe('CaseCreate', () => {
  afterEach(() => {
    mockNavigate.mockReset()
    mockPost.mockReset()
  })

  it('shows validation errors when required fields are missing', async () => {
    render(
      <MemoryRouter>
        <CaseCreate />
      </MemoryRouter>,
    )

    fireEvent.click(screen.getByRole('button', { name: 'Create case' }))

    expect(await screen.findByText('Title is required.')).toBeInTheDocument()
    expect(screen.getByText('Type is required.')).toBeInTheDocument()
    expect(mockPost).not.toHaveBeenCalled()
  })

  it('submits the form and navigates to the new case', async () => {
    mockPost.mockResolvedValue({
      data: {
        data: {
          id: 'c9d1f5bc-4bb6-4df3-97a1-7281df4f5a11',
        },
      },
    })

    render(
      <MemoryRouter>
        <CaseCreate />
      </MemoryRouter>,
    )

    fireEvent.change(screen.getByLabelText('Title'), {
      target: { value: 'Payment portal upgrade' },
    })
    fireEvent.change(screen.getByLabelText('Type'), {
      target: { value: 'feature_addition' },
    })
    fireEvent.change(screen.getByLabelText('Company name'), {
      target: { value: 'Acme Corp.' },
    })
    fireEvent.change(screen.getByLabelText('Contact name'), {
      target: { value: 'Keiko Tanaka' },
    })
    fireEvent.change(screen.getByLabelText('Contact email'), {
      target: { value: 'keiko.tanaka@example.com' },
    })
    fireEvent.change(screen.getByLabelText('Existing system URL'), {
      target: { value: 'https://example.com/current' },
    })

    fireEvent.click(screen.getByRole('button', { name: 'Create case' }))

    await waitFor(() => {
      expect(mockPost).toHaveBeenCalledWith('/v1/cases', {
        body: {
          title: 'Payment portal upgrade',
          type: 'feature_addition',
          company_name: 'Acme Corp.',
          contact_name: 'Keiko Tanaka',
          contact_email: 'keiko.tanaka@example.com',
          existing_system_url: 'https://example.com/current',
        },
      })
    })

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith(
        '/cases/c9d1f5bc-4bb6-4df3-97a1-7281df4f5a11',
      )
    })
  })
})
