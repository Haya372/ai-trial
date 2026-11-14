import { Spinner } from './Spinner'

type LoadingProps = {
  label?: string
}

function Loading({ label }: LoadingProps) {
  return (
    <div
      data-slot="loading"
      role="status"
      aria-live="polite"
      className="flex items-center justify-center gap-2 text-sm text-muted-foreground"
    >
      <Spinner size="md" />
      {label && <span data-slot="loading-label">{label}</span>}
    </div>
  )
}

export { Loading }
