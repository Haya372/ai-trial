import * as React from 'react'
import {
  Controller,
  type ControllerProps,
  type FieldPath,
  type FieldValues,
  FormProvider,
  useFormContext,
} from 'react-hook-form'
import { cn } from 'cn'
import { Label } from './Label'

const Form = FormProvider

type FormFieldContextValue<
  TFieldValues extends FieldValues = FieldValues,
  TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
> = {
  name: TName
}

const FormFieldContext = React.createContext<FormFieldContextValue>(
  {} as FormFieldContextValue,
)

function FormField<
  TFieldValues extends FieldValues = FieldValues,
  TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
>({ ...props }: ControllerProps<TFieldValues, TName>) {
  return (
    <FormFieldContext.Provider value={{ name: props.name }}>
      <Controller {...props} />
    </FormFieldContext.Provider>
  )
}

function FormItem({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('flex flex-col gap-1', className)} {...props} />
}

function FormLabel(props: React.ComponentPropsWithoutRef<typeof Label>) {
  const { name } = React.useContext(FormFieldContext)
  return <Label htmlFor={name} {...props} />
}

function FormMessage({
  className,
  ...props
}: React.HTMLAttributes<HTMLSpanElement>) {
  const { name } = React.useContext(FormFieldContext)
  const { getFieldState, formState } = useFormContext()
  const { error } = getFieldState(name, formState)
  if (!error?.message) return null
  return (
    <span className={cn('text-destructive text-sm', className)} {...props}>
      {error.message}
    </span>
  )
}

export { Form, FormField, FormItem, FormLabel, FormMessage }
