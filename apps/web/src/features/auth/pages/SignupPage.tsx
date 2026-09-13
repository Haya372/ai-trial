import { zodResolver } from '@hookform/resolvers/zod'
import { Button, Input, Label, Text, toast } from '@repo/ui'
import { useForm } from 'react-hook-form'
import { signup } from '../../../api/generated'
import { useAuthStore } from '../../../stores/auth'
import { getSignupErrorMessage } from '../utils'
import { type SignupFormValues, signupSchema } from '../types'

interface SignupPageProps {
  onSuccess?: () => void
}

export default function SignupPage({ onSuccess }: SignupPageProps) {
  const setUser = useAuthStore((s) => s.setUser)
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<SignupFormValues>({
    resolver: zodResolver(signupSchema),
    mode: 'onTouched',
  })

  const onSubmit = async (data: SignupFormValues) => {
    try {
      const res = await signup({
        email: data.email,
        password: data.password,
        displayName: data.displayName,
      })
      setUser(res.data)
      toast.success('アカウントを作成しました')
      onSuccess?.()
    } catch (error) {
      toast.error(getSignupErrorMessage(error))
    }
  }

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      className="flex flex-col gap-4 max-w-sm mx-auto mt-16"
    >
      <Text variant="h2">新規登録</Text>
      <div className="flex flex-col gap-1">
        <Label htmlFor="email">メールアドレス</Label>
        <Input
          id="email"
          type="email"
          state={errors.email ? 'error' : 'default'}
          aria-invalid={!!errors.email}
          {...register('email')}
        />
        {errors.email && (
          <span className="text-destructive text-sm">
            {errors.email.message}
          </span>
        )}
      </div>
      <div className="flex flex-col gap-1">
        <Label htmlFor="password">パスワード</Label>
        <Input
          id="password"
          type="password"
          state={errors.password ? 'error' : 'default'}
          aria-invalid={!!errors.password}
          {...register('password')}
        />
        {errors.password && (
          <span className="text-destructive text-sm">
            {errors.password.message}
          </span>
        )}
      </div>
      <div className="flex flex-col gap-1">
        <Label htmlFor="displayName">表示名</Label>
        <Input id="displayName" type="text" {...register('displayName')} />
        {errors.displayName && (
          <span className="text-destructive text-sm">
            {errors.displayName.message}
          </span>
        )}
      </div>
      <Button type="submit" variant="primary" disabled={isSubmitting}>
        {isSubmitting ? '登録中…' : '登録'}
      </Button>
    </form>
  )
}
