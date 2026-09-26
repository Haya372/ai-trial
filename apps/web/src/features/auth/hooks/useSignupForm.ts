import { useEffect, useMemo } from 'react'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from '@repo/ui'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { signup } from '../../../api/generated'
import { useAuthStore } from '../../../stores/auth'
import { createSignupSchema, type SignupFormValues } from '../types'
import { getSignupErrorMessage } from '../utils'

export function useSignupForm(onSuccess?: () => void) {
  const { t } = useTranslation(['auth', 'common'])
  const setUser = useAuthStore((s) => s.setUser)
  const schema = useMemo(() => createSignupSchema(t), [t])
  const form = useForm<SignupFormValues>({
    resolver: zodResolver(schema),
    mode: 'onTouched',
    defaultValues: { email: '', password: '', displayName: '' },
  })

  useEffect(() => {
    const erroredFields = Object.keys(form.formState.errors) as Array<
      keyof SignupFormValues
    >
    if (erroredFields.length > 0) {
      form.trigger(erroredFields)
    }
    // Deliberately keyed on `t` alone: this re-validates currently-errored
    // fields only when the language changes, so a displayed error message
    // is re-translated instead of staying stuck until the field is next
    // touched. Adding form.formState.errors here would make trigger()
    // re-fire this same effect in a loop.
    // oxlint-disable-next-line react-hooks/exhaustive-deps
  }, [t])

  const onSubmit = async (data: SignupFormValues) => {
    try {
      const res = await signup({
        email: data.email,
        password: data.password,
        displayName: data.displayName || undefined,
      })
      if (res.status !== 201) {
        toast.error(getSignupErrorMessage(res.data, t))
        return
      }
      setUser(res.data)
      toast.success(t('signup.toastSuccess'))
      onSuccess?.()
    } catch (error) {
      toast.error(getSignupErrorMessage(error, t))
    }
  }

  return { form, onSubmit }
}
