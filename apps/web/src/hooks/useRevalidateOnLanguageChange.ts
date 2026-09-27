import { useEffect } from 'react'
import type { FieldValues, UseFormReturn } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

export function useRevalidateOnLanguageChange<T extends FieldValues>(
  form: UseFormReturn<T>,
): void {
  const { i18n } = useTranslation()

  useEffect(() => {
    const erroredFields = Object.keys(form.formState.errors) as Array<keyof T>
    if (erroredFields.length > 0) {
      form.trigger(erroredFields)
    }
    // Deliberately keyed on `i18n.language` alone: this re-validates
    // currently-errored fields only when the language changes, so a
    // displayed error message is re-translated instead of staying stuck
    // until the field is next touched. Adding form.formState.errors here
    // would make trigger() re-fire this same effect in a loop.
    // oxlint-disable-next-line react-hooks/exhaustive-deps
  }, [i18n.language])
}
