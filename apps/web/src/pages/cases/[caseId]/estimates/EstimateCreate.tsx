import { type ChangeEvent, type FormEvent, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useCreateEstimate } from '@/hooks/use-estimates'

type FormValues = {
  yourHourlyRate: string
  region: string
  includeMarketEvidence: boolean
}

type FormErrors = Partial<Record<keyof FormValues, string>>

const initialFormValues: FormValues = {
  yourHourlyRate: '',
  region: 'japan',
  includeMarketEvidence: true,
}

function validateForm(values: FormValues): FormErrors {
  const errors: FormErrors = {}
  const parsedRate = Number(values.yourHourlyRate)

  if (!values.yourHourlyRate.trim()) {
    errors.yourHourlyRate = 'Hourly rate is required.'
  } else if (!Number.isFinite(parsedRate) || parsedRate <= 0) {
    errors.yourHourlyRate = 'Enter a valid hourly rate.'
  }

  if (!values.region.trim()) {
    errors.region = 'Region is required.'
  }

  return errors
}

export function EstimateCreate() {
  const navigate = useNavigate()
  const { caseId } = useParams<{ caseId: string }>()
  const { createEstimate: create, isSubmitting, error } = useCreateEstimate(caseId)
  const [values, setValues] = useState<FormValues>(initialFormValues)
  const [errors, setErrors] = useState<FormErrors>({})

  function handleFieldChange(
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) {
    const { name, value, type } = event.target
    const fieldName = name as keyof FormValues
    const nextValue =
      type === 'checkbox'
        ? (event.target as HTMLInputElement).checked
        : value

    setValues((currentValues) => ({
      ...currentValues,
      [fieldName]: nextValue,
    }))

    setErrors((currentErrors) => {
      const nextErrors = { ...currentErrors }
      delete nextErrors[fieldName]
      return nextErrors
    })
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const nextErrors = validateForm(values)
    setErrors(nextErrors)

    if (Object.keys(nextErrors).length > 0) {
      return
    }

    const estimate = await create({
      your_hourly_rate: Number(values.yourHourlyRate),
      region: values.region.trim(),
      include_market_evidence: values.includeMarketEvidence,
    })

    if (estimate?.id && caseId) {
      navigate(`/cases/${caseId}/estimates/${estimate.id}`)
    }
  }

  if (!caseId) {
    return (
      <main className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <p className="text-sm text-slate-600">Case id is missing from the route.</p>
      </main>
    )
  }

  return (
    <main className="space-y-6">
      <header className="space-y-2">
        <p className="text-sm font-medium text-slate-500">Estimates</p>
        <h1 className="text-balance text-3xl font-semibold text-slate-950">
          Create an estimate
        </h1>
        <p className="max-w-2xl text-pretty text-sm text-slate-600">
          Start with your hourly rate and whether the estimate should include
          market evidence. The backend determines the estimation mode from these
          inputs.
        </p>
      </header>

      <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <form className="space-y-6" onSubmit={handleSubmit} noValidate>
          <div className="grid gap-6 md:grid-cols-2">
            <label className="space-y-2 text-sm font-medium text-slate-700">
              <span>Your hourly rate</span>
              <input
                name="yourHourlyRate"
                type="number"
                min="0"
                step="1000"
                value={values.yourHourlyRate}
                onChange={handleFieldChange}
                aria-invalid={Boolean(errors.yourHourlyRate)}
                className="block w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 shadow-sm focus:border-slate-500 focus:outline-none"
                placeholder="12000"
              />
              {errors.yourHourlyRate ? (
                <span className="text-sm text-rose-700">
                  {errors.yourHourlyRate}
                </span>
              ) : null}
            </label>

            <label className="space-y-2 text-sm font-medium text-slate-700">
              <span>Region</span>
              <input
                name="region"
                value={values.region}
                onChange={handleFieldChange}
                aria-invalid={Boolean(errors.region)}
                className="block w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 shadow-sm focus:border-slate-500 focus:outline-none"
                placeholder="japan"
              />
              {errors.region ? (
                <span className="text-sm text-rose-700">{errors.region}</span>
              ) : null}
            </label>
          </div>

          <label className="flex items-start gap-3 rounded-xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-700">
            <input
              name="includeMarketEvidence"
              type="checkbox"
              checked={values.includeMarketEvidence}
              onChange={handleFieldChange}
              className="mt-0.5 size-4 rounded border border-slate-300"
            />
            <span className="text-pretty">
              Include market evidence so the estimate can compare your proposal
              against external benchmarks and generate the three-way proposal.
            </span>
          </label>

          {error ? (
            <p className="rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
              {error}
            </p>
          ) : null}

          <div className="flex flex-col gap-3 sm:flex-row">
            <button
              type="submit"
              disabled={isSubmitting}
              className="inline-flex items-center justify-center rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-slate-700 disabled:cursor-not-allowed disabled:bg-slate-300"
            >
              {isSubmitting ? 'Creating...' : 'Generate estimate'}
            </button>

            <Link
              to={`/cases/${caseId}/estimates`}
              className="inline-flex items-center justify-center rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 transition-colors hover:border-slate-400 hover:text-slate-950"
            >
              Cancel
            </Link>
          </div>
        </form>
      </section>
    </main>
  )
}
