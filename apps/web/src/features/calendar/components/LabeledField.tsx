import type { ReactNode } from 'react'

interface LabeledFieldProps {
  label: string
  children: ReactNode
}

export default function LabeledField({ label, children }: LabeledFieldProps) {
  return (
    <div>
      <span className="text-muted-foreground">{label}: </span>
      <span>{children}</span>
    </div>
  )
}
