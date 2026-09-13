import { Button, Input, Text } from '@repo/ui'

function App() {
  return (
    <div className="flex flex-col gap-8 p-8">
      <section className="flex flex-col gap-3">
        <Text variant="h2">Typography</Text>
        <Text variant="h1">Heading 1</Text>
        <Text variant="h2">Heading 2</Text>
        <Text variant="h3">Heading 3</Text>
        <Text variant="body">The quick brown fox jumps over the lazy dog.</Text>
        <Text variant="caption">This is a caption or helper text.</Text>
        <Text variant="code">const greeting = "Hello, world!"</Text>
      </section>

      <section className="flex flex-col gap-3">
        <Text variant="h2">Input</Text>
        <div className="flex flex-col gap-2 w-72">
          <Input placeholder="Default input" />
          <Input placeholder="Small input" size="sm" />
          <Input placeholder="Large input" size="lg" />
          <Input placeholder="Error state" state="error" aria-invalid />
          <Input placeholder="Disabled" disabled />
        </div>
      </section>

      <section className="flex flex-col gap-3">
        <Text variant="h2">Button</Text>
        <div className="flex gap-2 flex-wrap">
          <Button variant="primary">Primary</Button>
          <Button variant="secondary">Secondary</Button>
          <Button variant="destructive">Destructive</Button>
          <Button variant="outline">Outline</Button>
          <Button variant="ghost">Ghost</Button>
          <Button variant="link">Link</Button>
        </div>
        <div className="flex gap-2 items-center">
          <Button size="sm">Small</Button>
          <Button size="md">Medium</Button>
          <Button size="lg">Large</Button>
        </div>
      </section>
    </div>
  )
}

export default App
