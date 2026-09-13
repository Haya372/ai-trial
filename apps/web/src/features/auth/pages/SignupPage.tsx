import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { signup } from '../../../api/generated'
import { useAuthStore } from '../../../stores/auth'

const signupSchema = z.object({
  email: z.string().min(1, 'Email is required').email('Invalid email format'),
  password: z
    .string()
    .min(8, 'Password must be at least 8 characters')
    .max(128, 'Password must be at most 128 characters'),
  displayName: z
    .string()
    .max(50, 'Display name must be at most 50 characters')
    .optional(),
})

type SignupForm = z.infer<typeof signupSchema>

interface SignupPageProps {
  onSuccess?: () => void
}

export default function SignupPage({ onSuccess }: SignupPageProps) {
  const setUser = useAuthStore((s) => s.setUser)
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    setError,
  } = useForm<SignupForm>({ resolver: zodResolver(signupSchema) })

  const onSubmit = async (data: SignupForm) => {
    try {
      const res = await signup({
        email: data.email,
        password: data.password,
        displayName: data.displayName,
      })
      setUser(res.data)
      onSuccess?.()
    } catch {
      setError('root', { message: 'Signup failed. Please try again.' })
    }
  }

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      className="flex flex-col gap-4 max-w-sm mx-auto mt-16"
    >
      <h1 className="text-2xl font-bold">Sign Up</h1>
      <div className="flex flex-col gap-1">
        <label htmlFor="email">Email</label>
        <input
          id="email"
          type="email"
          {...register('email')}
          className="border rounded px-3 py-2"
        />
        {errors.email && (
          <span className="text-red-500 text-sm">{errors.email.message}</span>
        )}
      </div>
      <div className="flex flex-col gap-1">
        <label htmlFor="password">Password</label>
        <input
          id="password"
          type="password"
          {...register('password')}
          className="border rounded px-3 py-2"
        />
        {errors.password && (
          <span className="text-red-500 text-sm">
            {errors.password.message}
          </span>
        )}
      </div>
      <div className="flex flex-col gap-1">
        <label htmlFor="displayName">Display Name</label>
        <input
          id="displayName"
          type="text"
          {...register('displayName')}
          className="border rounded px-3 py-2"
        />
        {errors.displayName && (
          <span className="text-red-500 text-sm">
            {errors.displayName.message}
          </span>
        )}
      </div>
      {errors.root && (
        <span className="text-red-500 text-sm">{errors.root.message}</span>
      )}
      <button
        type="submit"
        disabled={isSubmitting}
        className="bg-blue-600 text-white rounded px-4 py-2"
      >
        {isSubmitting ? 'Signing up…' : 'Sign Up'}
      </button>
    </form>
  )
}
