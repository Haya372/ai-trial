import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    environment: 'node',
    // Vitest blanks CSS by default; the tag color tests read the raw token values.
    css: { include: [/tokens\.css/] },
  },
})
