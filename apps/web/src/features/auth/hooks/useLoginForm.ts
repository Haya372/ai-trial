import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from '@repo/ui'
import { useForm } from 'react-hook-form'
import { login } from '../../../api/generated'
import { useAuthStore } from '../../../stores/auth'
import { type LoginFormValues, loginSchema } from '../types'
import { getLoginErrorMessage } from '../utils'

export function useLoginForm(onSuccess?: () => void) {
  const setUser = useAuthStore((s) => s.setUser)
  const form = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    mode: 'onTouched',
    defaultValues: { email: '', password: '' },
  })

  const onSubmit = async (data: LoginFormValues) => {
    try {
      const res = await login({ email: data.email, password: data.password })
      if (res.status !== 200) {
        toast.error(getLoginErrorMessage(res.data))
        return
      }
      setUser(res.data)
      toast.success('ログインしました')
      onSuccess?.()
    } catch (error) {
      toast.error(getLoginErrorMessage(error))
    }
  }

  return { form, onSubmit }
}
