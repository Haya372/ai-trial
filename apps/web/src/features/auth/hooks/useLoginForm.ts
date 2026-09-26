import { useMemo } from 'react'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from '@repo/ui'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { login } from '../../../api/generated'
import { useAuthStore } from '../../../stores/auth'
import { createLoginSchema, type LoginFormValues } from '../types'
import { getLoginErrorMessage } from '../utils'

export function useLoginForm(onSuccess?: () => void) {
  const { t } = useTranslation('auth')
  const setUser = useAuthStore((s) => s.setUser)
  const schema = useMemo(() => createLoginSchema(t), [t])
  const form = useForm<LoginFormValues>({
    resolver: zodResolver(schema),
    mode: 'onTouched',
    defaultValues: { email: '', password: '' },
  })

  const onSubmit = async (data: LoginFormValues) => {
    try {
      const res = await login({ email: data.email, password: data.password })
      if (res.status !== 200) {
        toast.error(getLoginErrorMessage(res.data, t))
        return
      }
      setUser(res.data)
      toast.success(t('login.toastSuccess'))
      onSuccess?.()
    } catch (error) {
      toast.error(getLoginErrorMessage(error, t))
    }
  }

  return { form, onSubmit }
}
