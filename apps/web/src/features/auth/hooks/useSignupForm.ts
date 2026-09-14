import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from '@repo/ui'
import { useForm } from 'react-hook-form'
import { signup } from '../../../api/generated'
import { useAuthStore } from '../../../stores/auth'
import { type SignupFormValues, signupSchema } from '../types'
import { getSignupErrorMessage } from '../utils'

export function useSignupForm(onSuccess?: () => void) {
  const setUser = useAuthStore((s) => s.setUser)
  const form = useForm<SignupFormValues>({
    resolver: zodResolver(signupSchema),
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
      setUser(res.data)
      toast.success('アカウントを作成しました')
      onSuccess?.()
    } catch (error) {
      toast.error(getSignupErrorMessage(error))
    }
  }

  return { form, onSubmit }
}
