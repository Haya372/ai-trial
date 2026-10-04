import { useTranslation } from 'react-i18next'
import type { ShareResponse } from '../../../api/generated'
import { formatShareDateTime } from '../utils'

interface ShareEventDetailProps {
  event: ShareResponse
}

interface LabeledFieldProps {
  label: string
  children: React.ReactNode
}

function LabeledField({ label, children }: LabeledFieldProps) {
  return (
    <div>
      <span className="text-muted-foreground">{label}: </span>
      <span>{children}</span>
    </div>
  )
}

export default function ShareEventDetail({ event }: ShareEventDetailProps) {
  const { t } = useTranslation('share')

  return (
    <div className="flex flex-col gap-3">
      <h2 className="text-xl font-semibold">{event.title}</h2>
      <div className="flex flex-col gap-2 text-sm">
        <LabeledField label={t('fields.startLabel')}>
          {formatShareDateTime(event.startAt, t)}
        </LabeledField>
        <LabeledField label={t('fields.endLabel')}>
          {formatShareDateTime(event.endAt, t)}
        </LabeledField>
        {event.location != null && (
          <LabeledField label={t('fields.locationLabel')}>
            {event.location}
          </LabeledField>
        )}
        {event.url != null && (
          <LabeledField label={t('fields.urlLabel')}>
            <a
              href={event.url}
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary underline"
            >
              {event.url}
            </a>
          </LabeledField>
        )}
        {event.description != null && (
          <LabeledField label={t('fields.noteLabel')}>
            {event.description}
          </LabeledField>
        )}
      </div>
    </div>
  )
}
