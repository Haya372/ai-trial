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
import { useTranslation } from 'react-i18next'
import { useSignupForm } from '../hooks/useSignupForm'

interface SignupPageProps {
  onSuccess?: () => void
}

export default function SignupPage({ onSuccess }: SignupPageProps) {
  const { t } = useTranslation('auth')
  const { form, onSubmit } = useSignupForm(onSuccess)

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className="flex flex-col gap-4 max-w-sm mx-auto mt-16"
      >
        <Text variant="h2">{t('signup.title')}</Text>
        <FormField
          control={form.control}
          name="email"
          render={({ field, fieldState }) => (
            <FormItem>
              <FormLabel>{t('fields.email')}</FormLabel>
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
              <FormLabel>{t('fields.password')}</FormLabel>
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
              <FormLabel>{t('fields.displayName')}</FormLabel>
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
          {form.formState.isSubmitting
            ? t('signup.submitButtonLoading')
            : t('signup.submitButton')}
        </Button>
      </form>
    </Form>
  )
}
