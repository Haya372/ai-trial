import { useMemo } from 'react'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from '@repo/ui'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { signup } from '../../../api/generated'
import { useAuthStore } from '../../../stores/auth'
import { createSignupSchema, type SignupFormValues } from '../types'
import { getSignupErrorMessage } from '../utils'

export function useSignupForm(onSuccess?: () => void) {
  const { t } = useTranslation('auth')
  const setUser = useAuthStore((s) => s.setUser)
  const schema = useMemo(() => createSignupSchema(t), [t])
  const form = useForm<SignupFormValues>({
    resolver: zodResolver(schema),
    mode: 'onTouched',
    defaultValues: { email: '', password: '', displayName: '' },
  })

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
