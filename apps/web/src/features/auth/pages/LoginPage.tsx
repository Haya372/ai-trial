import { zodResolver } from '@hookform/resolvers/zod'
import { Button, Input, Label, Text } from '@repo/ui'
import { useForm } from 'react-hook-form'
import { login } from '../../../api/generated'
import { useAuthStore } from '../../../stores/auth'
import { type LoginFormValues, loginSchema } from '../types'

interface LoginPageProps {
  onSuccess?: () => void
}

export default function LoginPage({ onSuccess }: LoginPageProps) {
  const setUser = useAuthStore((s) => s.setUser)
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    setError,
  } = useForm<LoginFormValues>({ resolver: zodResolver(loginSchema) })

  const onSubmit = async (data: LoginFormValues) => {
    try {
      const res = await login({ email: data.email, password: data.password })
      setUser(res.data)
      onSuccess?.()
    } catch {
      setError('root', { message: 'Invalid email or password' })
    }
  }

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      className="flex flex-col gap-4 max-w-sm mx-auto mt-16"
    >
      <Text variant="h2">Login</Text>
      <div className="flex flex-col gap-1">
        <Label htmlFor="email">Email</Label>
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
        <Label htmlFor="password">Password</Label>
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
      {errors.root && (
        <span className="text-destructive text-sm">{errors.root.message}</span>
      )}
      <Button type="submit" variant="primary" disabled={isSubmitting}>
        {isSubmitting ? 'Logging in…' : 'Login'}
      </Button>
    </form>
  )
}
