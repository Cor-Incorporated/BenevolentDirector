import { useCallback, useEffect, useState } from 'react'
import {
  createEstimate,
  getEstimate,
  getThreeWayProposal,
  listEstimates,
  getApiErrorMessage,
} from '@/lib/api-client'
import type {
  CreateEstimateInput,
  Estimate,
  EstimateWithProposal,
  ThreeWayProposal,
} from '@/types/estimate'

type UseEstimatesReturn = {
  estimates: Estimate[]
  total: number
  loading: boolean
  isLoading: boolean
  error: string | null
  refresh: () => Promise<void>
}

export function useEstimates(caseId?: string): UseEstimatesReturn {
  const [estimates, setEstimates] = useState<Estimate[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    if (!caseId) {
      setEstimates([])
      setTotal(0)
      setLoading(false)
      return
    }

    setLoading(true)
    setError(null)

    try {
      const result = await listEstimates(caseId)
      setEstimates(result.data)
      setTotal(result.total)
    } catch (err) {
      setEstimates([])
      setTotal(0)
      setError(getApiErrorMessage(err, 'Failed to load estimates.'))
    } finally {
      setLoading(false)
    }
  }, [caseId])

  useEffect(() => {
    void refresh()
  }, [refresh])

  return { estimates, total, loading, isLoading: loading, error, refresh }
}

type UseEstimateReturn = {
  estimate: EstimateWithProposal | null
  loading: boolean
  isLoading: boolean
  error: string | null
  refresh: () => Promise<void>
}

export function useEstimate(
  caseId?: string,
  estimateId?: string,
): UseEstimateReturn {
  const [estimate, setEstimate] = useState<EstimateWithProposal | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    if (!caseId || !estimateId) {
      setEstimate(null)
      setLoading(false)
      return
    }

    setLoading(true)
    setError(null)

    try {
      const result = await getEstimate(caseId, estimateId)
      setEstimate(result)
    } catch (err) {
      setEstimate(null)
      setError(getApiErrorMessage(err, 'Failed to load estimate.'))
    } finally {
      setLoading(false)
    }
  }, [caseId, estimateId])

  useEffect(() => {
    void refresh()
  }, [refresh])

  return { estimate, loading, isLoading: loading, error, refresh }
}

type UseCreateEstimateReturn = {
  create: (input: CreateEstimateInput) => Promise<Estimate | null>
  loading: boolean
  createEstimate: (input: CreateEstimateInput) => Promise<Estimate | null>
  isSubmitting: boolean
  error: string | null
}

export function useCreateEstimate(caseId?: string): UseCreateEstimateReturn {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const create = useCallback(
    async (input: CreateEstimateInput): Promise<Estimate | null> => {
      if (!caseId) {
        setError('Case ID is required.')
        return null
      }

      setLoading(true)
      setError(null)

      try {
        const result = await createEstimate(caseId, input)
        return result
      } catch (err) {
        setError(getApiErrorMessage(err, 'Failed to create estimate.'))
        return null
      } finally {
        setLoading(false)
      }
    },
    [caseId],
  )

  return {
    create,
    loading,
    createEstimate: create,
    isSubmitting: loading,
    error,
  }
}

type UseThreeWayProposalReturn = {
  proposal: ThreeWayProposal | null
  loading: boolean
  isLoading: boolean
  error: string | null
  refresh: () => Promise<void>
}

export function useThreeWayProposal(
  caseId?: string,
  estimateId?: string,
): UseThreeWayProposalReturn {
  const [proposal, setProposal] = useState<ThreeWayProposal | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    if (!caseId || !estimateId) {
      setProposal(null)
      setLoading(false)
      return
    }

    setLoading(true)
    setError(null)

    try {
      const result = await getThreeWayProposal(caseId, estimateId)
      setProposal(result)
    } catch (err) {
      setProposal(null)
      setError(getApiErrorMessage(err, 'Failed to load three-way proposal.'))
    } finally {
      setLoading(false)
    }
  }, [caseId, estimateId])

  useEffect(() => {
    void refresh()
  }, [refresh])

  return { proposal, loading, isLoading: loading, error, refresh }
}
