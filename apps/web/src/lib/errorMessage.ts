export function hasErrorCode(value: unknown): value is { code: string } {
  return typeof value === 'object' && value !== null && 'code' in value
}

export function mapErrorToMessage<K extends string>(
  error: unknown,
  t: (key: K) => string,
  codeToKey: Record<string, K>,
  fallbackKey: K,
): string {
  if (hasErrorCode(error) && error.code in codeToKey) {
    return t(codeToKey[error.code])
  }
  return t(fallbackKey)
}
