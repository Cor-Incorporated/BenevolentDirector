import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { EstimateCreate } from '@/pages/cases/[caseId]/estimates/EstimateCreate'
import { useCreateEstimate } from '@/hooks/use-estimates'

const { mockNavigate, mockUseParams } = vi.hoisted(() => ({
  mockNavigate: vi.fn(),
  mockUseParams: vi.fn(),
}))

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>(
    'react-router-dom',
  )

  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: mockUseParams,
  }
})

vi.mock('@/hooks/use-estimates', () => ({
  useCreateEstimate: vi.fn(),
}))

const mockedUseCreateEstimate = vi.mocked(useCreateEstimate)

describe('EstimateCreate', () => {
  beforeEach(() => {
    mockUseParams.mockReturnValue({ caseId: 'case-1' })
    mockNavigate.mockReset()
  })

  it('validates required input before submitting', async () => {
    mockedUseCreateEstimate.mockReturnValue({
      createEstimate: vi.fn(),
      isSubmitting: false,
      error: null,
    })

    render(
      <MemoryRouter>
        <EstimateCreate />
      </MemoryRouter>,
    )

    fireEvent.click(screen.getByRole('button', { name: 'Generate estimate' }))

    expect(
      await screen.findByText('Hourly rate is required.'),
    ).toBeInTheDocument()
  })

  it('submits the form and navigates to the new estimate', async () => {
    const create = vi.fn().mockResolvedValue({
      id: 'estimate-1',
    })

    mockedUseCreateEstimate.mockReturnValue({
      createEstimate: create,
      isSubmitting: false,
      error: null,
    })

    render(
      <MemoryRouter>
        <EstimateCreate />
      </MemoryRouter>,
    )

    fireEvent.change(screen.getByLabelText('Your hourly rate'), {
      target: { value: '12000' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'Generate estimate' }))

    await waitFor(() => {
      expect(create).toHaveBeenCalledWith({
        your_hourly_rate: 12000,
        region: 'japan',
        include_market_evidence: true,
      })
    })

    expect(mockNavigate).toHaveBeenCalledWith('/cases/case-1/estimates/estimate-1')
  })
})
