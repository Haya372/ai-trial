export function SemgrepCanary({ html }: { html: string }) {
  return <div dangerouslySetInnerHTML={{ __html: html }} />;
}
