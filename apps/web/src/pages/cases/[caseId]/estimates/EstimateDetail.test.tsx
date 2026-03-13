import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { EstimateDetail } from '@/pages/cases/[caseId]/estimates/EstimateDetail'
import { useEstimate, useThreeWayProposal } from '@/hooks/use-estimates'

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
  useEstimate: vi.fn(),
  useThreeWayProposal: vi.fn(),
}))

vi.mock('@/components/estimate/ThreeWayProposalView', () => ({
  ThreeWayProposalView: () => <div>Three-way proposal rendered</div>,
}))

vi.mock('@/components/estimate/ApprovalActions', () => ({
  ApprovalActions: () => <div>Approval actions rendered</div>,
}))

const mockedUseEstimate = vi.mocked(useEstimate)
const mockedUseThreeWayProposal = vi.mocked(useThreeWayProposal)

describe('EstimateDetail', () => {
  beforeEach(() => {
    mockUseParams.mockReturnValue({ caseId: 'case-1', estimateId: 'estimate-1' })
  })

  it('renders loading state', () => {
    mockedUseEstimate.mockReturnValue({
      estimate: null,
      loading: true, isLoading: true,
      error: null,
      refresh: vi.fn(),
    })
    mockedUseThreeWayProposal.mockReturnValue({
      proposal: null,
      loading: true, isLoading: true,
      error: null,
      refresh: vi.fn(),
    })

    render(
      <MemoryRouter>
        <EstimateDetail />
      </MemoryRouter>,
    )

    expect(screen.getByText('Loading estimate...')).toBeInTheDocument()
  })

  it('renders estimate metrics and child sections', () => {
    mockedUseEstimate.mockReturnValue({
      estimate: {
        id: 'estimate-1',
        case_id: 'case-1',
        estimate_mode: 'hybrid',
        status: 'ready',
        total_your_cost: 1200000,
        your_estimated_hours: 100,
        total_market_cost: 1500000,
        your_hourly_rate: 12000,
        risk_flags: ['Evidence variance detected'],
      },
      loading: false, isLoading: false,
      error: null,
      refresh: vi.fn(),
    })
    mockedUseThreeWayProposal.mockReturnValue({
      proposal: {
        market_benchmark: {
          confidence: 'high',
        },
      },
      loading: false, isLoading: false,
      error: null,
      refresh: vi.fn(),
    })

    render(
      <MemoryRouter>
        <EstimateDetail />
      </MemoryRouter>,
    )

    expect(screen.getByText('Hybrid')).toBeInTheDocument()
    expect(screen.getByText('Ready')).toBeInTheDocument()
    expect(screen.getByText('Risk flags')).toBeInTheDocument()
    expect(screen.getByText('Three-way proposal rendered')).toBeInTheDocument()
    expect(screen.getByText('Approval actions rendered')).toBeInTheDocument()
  })
})
