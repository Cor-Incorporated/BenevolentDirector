export type EstimateMode = 'market_comparison' | 'hours_only' | 'hybrid'

export type EstimateStatus = 'draft' | 'ready' | 'approved' | 'rejected'

export type ConfidenceLevel = 'high' | 'medium' | 'low'

export type SourceAuthority = 'official' | 'industry' | 'community' | 'unknown'

export type ProposalSessionStatus =
  | 'draft'
  | 'presented'
  | 'approved'
  | 'rejected'
  | 'expired'

export type ApprovalDecisionType = 'approved' | 'rejected'

export type GoNoGoDecision = 'go' | 'go_with_conditions' | 'no_go'

export interface Range {
  min?: number
  max?: number
}

export interface Citation {
  url?: string
  title?: string
  source_authority?: SourceAuthority
  snippet?: string
}

export interface ProposalContradiction {
  provider_a?: string
  provider_b?: string
  field?: string
  description?: string
}

export interface SimilarProject {
  name?: string
  actual_hours?: number
  similarity_score?: number
}

export interface TrackRecordAxis {
  similar_projects?: SimilarProject[]
  median_hours?: number
  velocity_score?: number
}

export interface MarketBenchmarkAxis {
  consensus_hours?: Range
  consensus_rate?: Range
  confidence?: ConfidenceLevel
  provider_count?: number
  citations?: Citation[]
  contradictions?: ProposalContradiction[]
}

export interface OurProposalAxis {
  proposed_hours?: number
  proposed_rate?: number
  proposed_total?: number
  savings_vs_market_percent?: number
  competitive_advantages?: string[]
  calibration_note?: string
}

export interface ThreeWayProposal {
  our_track_record?: TrackRecordAxis
  market_benchmark?: MarketBenchmarkAxis
  our_proposal?: OurProposalAxis
}

export interface GoNoGoScores {
  profitability?: number
  strategic_alignment?: number
  capacity?: number
  technical_risk?: number
}

export interface GoNoGoResult {
  decision?: GoNoGoDecision
  scores?: GoNoGoScores
  weights?: GoNoGoScores
  reasoning?: string
}

export interface Estimate {
  id: string
  case_id: string
  estimate_mode: EstimateMode
  status: EstimateStatus
  your_hourly_rate?: number
  your_estimated_hours?: number
  total_your_cost?: number
  hours_investigation?: number
  hours_implementation?: number
  hours_testing?: number
  hours_buffer?: number
  market_hourly_rate?: number
  market_estimated_hours?: number
  total_market_cost?: number
  aggregated_evidence_id?: string
  calibration_ratio?: number
  risk_flags?: string[]
  created_at?: string
}

export interface EstimateWithProposal extends Estimate {
  three_way_proposal?: ThreeWayProposal
  go_no_go?: GoNoGoResult
}

export interface CreateEstimateInput {
  your_hourly_rate: number
  region?: string
  include_market_evidence?: boolean
}

export interface ProposalSession {
  id: string
  case_id: string
  estimate_id: string
  status: ProposalSessionStatus
  presented_at?: string
  decided_at?: string
  created_at?: string
}

export interface ApprovalDecision {
  id?: string
  proposal_id?: string
  decision?: ApprovalDecisionType
  decided_by_uid?: string
  comment?: string
  decided_at?: string
}

export interface EstimateListResponse {
  data?: Estimate[]
  total?: number
}

export interface EstimateDetailResponse {
  data?: EstimateWithProposal | null
}

export interface EstimateCreateResponse {
  data?: Estimate | null
}

export interface ThreeWayProposalResponse {
  data?: ThreeWayProposal | null
}

export interface ProposalListResponse {
  data?: ProposalSession[]
  total?: number
}

export interface ProposalSessionResponse {
  data?: ProposalSession | null
}

export interface ApprovalDecisionResponse {
  data?: ApprovalDecision | null
}

export const estimateModeLabels: Record<EstimateMode, string> = {
  market_comparison: 'Market comparison',
  hours_only: 'Hours only',
  hybrid: 'Hybrid',
}

export const estimateStatusLabels: Record<EstimateStatus, string> = {
  draft: 'Draft',
  ready: 'Ready',
  approved: 'Approved',
  rejected: 'Rejected',
}

export const confidenceLevelLabels: Record<ConfidenceLevel, string> = {
  high: 'High confidence',
  medium: 'Medium confidence',
  low: 'Low confidence',
}

export const sourceAuthorityLabels: Record<SourceAuthority, string> = {
  official: 'Official',
  industry: 'Industry',
  community: 'Community',
  unknown: 'Unknown',
}

export const proposalStatusLabels: Record<ProposalSessionStatus, string> = {
  draft: 'Draft',
  presented: 'Presented',
  approved: 'Approved',
  rejected: 'Rejected',
  expired: 'Expired',
}

export const goNoGoDecisionLabels: Record<GoNoGoDecision, string> = {
  go: 'Go',
  go_with_conditions: 'Go with conditions',
  no_go: 'No-go',
}

function isFiniteNumber(value?: number | null): value is number {
  return typeof value === 'number' && Number.isFinite(value)
}

export function formatEstimateCurrency(value?: number | null) {
  if (!isFiniteNumber(value)) {
    return 'Not available'
  }

  return new Intl.NumberFormat('ja-JP', {
    style: 'currency',
    currency: 'JPY',
    maximumFractionDigits: 0,
  }).format(value)
}

export function formatEstimateHours(value?: number | null) {
  if (!isFiniteNumber(value)) {
    return 'Not available'
  }

  return new Intl.NumberFormat('ja-JP', {
    maximumFractionDigits: value % 1 === 0 ? 0 : 1,
  }).format(value).concat('h')
}

export function formatEstimateNumber(value?: number | null) {
  if (!isFiniteNumber(value)) {
    return 'Not available'
  }

  return new Intl.NumberFormat('ja-JP', {
    maximumFractionDigits: value % 1 === 0 ? 0 : 2,
  }).format(value)
}

export function formatEstimatePercent(value?: number | null) {
  if (!isFiniteNumber(value)) {
    return 'Not available'
  }

  return `${formatEstimateNumber(value)}%`
}

export function formatEstimateRatio(value?: number | null) {
  if (!isFiniteNumber(value)) {
    return 'Not available'
  }

  return `${formatEstimateNumber(value)}x`
}

export function formatEstimateRange(
  range?: Range | null,
  formatter: (value?: number | null) => string = formatEstimateNumber,
) {
  if (!range || (!isFiniteNumber(range.min) && !isFiniteNumber(range.max))) {
    return 'Not available'
  }

  if (isFiniteNumber(range.min) && isFiniteNumber(range.max)) {
    return `${formatter(range.min)} - ${formatter(range.max)}`
  }

  if (isFiniteNumber(range.min)) {
    return `${formatter(range.min)}+`
  }

  return `Up to ${formatter(range.max)}`
}
