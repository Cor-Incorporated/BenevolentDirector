import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { EstimateList } from '@/pages/cases/[caseId]/estimates/EstimateList'
import { useEstimates } from '@/hooks/use-estimates'

const { mockUseParams } = vi.hoisted(() => ({
  mockUseParams: vi.fn(),
}))

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>(
    'react-router-dom',
  )

  return {
    ...actual,
    useParams: mockUseParams,
  }
})

vi.mock('@/hooks/use-estimates', () => ({
  useEstimates: vi.fn(),
}))

const mockedUseEstimates = vi.mocked(useEstimates)

describe('EstimateList', () => {
  beforeEach(() => {
    mockUseParams.mockReturnValue({ caseId: 'case-1' })
  })

  it('renders empty state when no estimates are available', () => {
    mockedUseEstimates.mockReturnValue({
      estimates: [],
      total: 0,
      loading: false, isLoading: false,
      error: null,
      refresh: vi.fn(),
    })

    render(
      <MemoryRouter>
        <EstimateList />
      </MemoryRouter>,
    )

    expect(screen.getByText('No estimates yet')).toBeInTheDocument()
    expect(
      screen.getByRole('link', { name: 'Create first estimate' }),
    ).toBeInTheDocument()
  })

  it('renders estimates returned by the hook', () => {
    mockedUseEstimates.mockReturnValue({
      estimates: [
        {
          id: 'estimate-1',
          case_id: 'case-1',
          estimate_mode: 'hybrid',
          status: 'ready',
          total_your_cost: 1200000,
          your_estimated_hours: 100,
          created_at: '2026-03-13T00:00:00Z',
        },
      ],
      total: 1,
      loading: false, isLoading: false,
      error: null,
      refresh: vi.fn(),
    })

    render(
      <MemoryRouter>
        <EstimateList />
      </MemoryRouter>,
    )

    expect(screen.getByText('Estimate list')).toBeInTheDocument()
    expect(screen.getByText('Hybrid')).toBeInTheDocument()
    expect(screen.getByText('Ready')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Open detail' })).toHaveAttribute(
      'href',
      '/cases/case-1/estimates/estimate-1',
    )
  })
})
