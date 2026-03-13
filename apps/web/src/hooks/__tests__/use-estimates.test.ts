import { act, renderHook } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  useCreateEstimate,
  useEstimate,
  useEstimates,
  useThreeWayProposal,
} from '@/hooks/use-estimates'
import {
  createEstimate,
  getEstimate,
  getThreeWayProposal,
  listEstimates,
} from '@/lib/api-client'

vi.mock('@/lib/api-client', () => ({
  createEstimate: vi.fn(),
  getApiErrorMessage: (error: unknown, fallback: string) =>
    error instanceof Error ? error.message : fallback,
  getEstimate: vi.fn(),
  getThreeWayProposal: vi.fn(),
  listEstimates: vi.fn(),
}))

const mockedListEstimates = vi.mocked(listEstimates)
const mockedGetEstimate = vi.mocked(getEstimate)
const mockedCreateEstimate = vi.mocked(createEstimate)
const mockedGetThreeWayProposal = vi.mocked(getThreeWayProposal)

async function flushAsyncWork() {
  await act(async () => {
    await Promise.resolve()
  })
}

describe('use-estimates hooks', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads estimates for the provided case', async () => {
    mockedListEstimates.mockResolvedValue({
      data: [
        {
          id: 'estimate-1',
          case_id: 'case-1',
          estimate_mode: 'hybrid',
          status: 'ready',
        },
      ],
      total: 1,
    })

    const { result } = renderHook(() => useEstimates('case-1'))
    await flushAsyncWork()

    expect(result.current.estimates).toHaveLength(1)
    expect(result.current.total).toBe(1)
    expect(result.current.error).toBeNull()
  })

  it('loads estimate detail and three-way proposal independently', async () => {
    mockedGetEstimate.mockResolvedValue({
      id: 'estimate-1',
      case_id: 'case-1',
      estimate_mode: 'hybrid',
      status: 'ready',
    })
    mockedGetThreeWayProposal.mockResolvedValue({
      market_benchmark: {
        confidence: 'high',
      },
    })

    const { result: estimateResult } = renderHook(() =>
      useEstimate('case-1', 'estimate-1'),
    )
    const { result: proposalResult } = renderHook(() =>
      useThreeWayProposal('case-1', 'estimate-1'),
    )

    await flushAsyncWork()

    expect(estimateResult.current.estimate?.id).toBe('estimate-1')
    expect(
      proposalResult.current.proposal?.market_benchmark?.confidence,
    ).toBe('high')
  })

  it('creates an estimate and exposes submission errors', async () => {
    mockedCreateEstimate.mockResolvedValue({
      id: 'estimate-2',
      case_id: 'case-1',
      estimate_mode: 'market_comparison',
      status: 'draft',
    })

    const { result } = renderHook(() => useCreateEstimate('case-1'))

    await act(async () => {
      const estimate = await result.current.createEstimate({
        your_hourly_rate: 12000,
        region: 'japan',
        include_market_evidence: true,
      })

      expect(estimate?.id).toBe('estimate-2')
    })

    mockedCreateEstimate.mockRejectedValueOnce(new Error('Failed request'))

    await act(async () => {
      const estimate = await result.current.createEstimate({
        your_hourly_rate: 12000,
      })

      expect(estimate).toBeNull()
    })

    expect(result.current.error).toBe('Failed request')
  })
})
