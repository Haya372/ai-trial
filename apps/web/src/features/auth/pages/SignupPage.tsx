import {
  Button,
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
  Text,
} from '@repo/ui'
import { useSignupForm } from '../hooks/useSignupForm'

interface SignupPageProps {
  onSuccess?: () => void
}

export default function SignupPage({ onSuccess }: SignupPageProps) {
  const { form, onSubmit } = useSignupForm(onSuccess)

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className="flex flex-col gap-4 max-w-sm mx-auto mt-16"
      >
        <Text variant="h2">新規登録</Text>
        <FormField
          control={form.control}
          name="email"
          render={({ field, fieldState }) => (
            <FormItem>
              <FormLabel>メールアドレス</FormLabel>
              <Input
                id={field.name}
                type="email"
                state={fieldState.error ? 'error' : 'default'}
                aria-invalid={!!fieldState.error}
                {...field}
              />
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="password"
          render={({ field, fieldState }) => (
            <FormItem>
              <FormLabel>パスワード</FormLabel>
              <Input
                id={field.name}
                type="password"
                state={fieldState.error ? 'error' : 'default'}
                aria-invalid={!!fieldState.error}
                {...field}
              />
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="displayName"
          render={({ field }) => (
            <FormItem>
              <FormLabel>表示名</FormLabel>
              <Input id={field.name} type="text" {...field} />
              <FormMessage />
            </FormItem>
          )}
        />
        <Button
          type="submit"
          variant="primary"
          disabled={form.formState.isSubmitting}
        >
          {form.formState.isSubmitting ? '登録中…' : '登録'}
        </Button>
      </form>
    </Form>
  )
}
