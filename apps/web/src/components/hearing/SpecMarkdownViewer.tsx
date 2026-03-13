interface SpecMarkdownViewerProps {
  markdown: string
}

function renderLine(line: string, index: number) {
  if (line.startsWith('### ')) {
    return (
      <h4 key={index} className="text-balance text-base font-semibold text-slate-900">
        {line.slice(4)}
      </h4>
    )
  }

  if (line.startsWith('## ')) {
    return (
      <h3 key={index} className="text-balance text-lg font-semibold text-slate-950">
        {line.slice(3)}
      </h3>
    )
  }

  if (line.startsWith('# ')) {
    return (
      <h2 key={index} className="text-balance text-xl font-semibold text-slate-950">
        {line.slice(2)}
      </h2>
    )
  }

  if (/^[-*] /.test(line)) {
    return (
      <p key={index} className="pl-4 text-pretty text-sm text-slate-700">
        • {line.slice(2)}
      </p>
    )
  }

  if (/^\d+\. /.test(line)) {
    return (
      <p key={index} className="pl-4 text-pretty text-sm text-slate-700">
        {line}
      </p>
    )
  }

  if (line.trim() === '') {
    return <div key={index} className="h-1" />
  }

  return (
    <p key={index} className="text-pretty text-sm leading-6 text-slate-700">
      {line}
    </p>
  )
}

export function SpecMarkdownViewer({ markdown }: SpecMarkdownViewerProps) {
  return (
    <div className="space-y-3">
      {markdown.split('\n').map((line, index) => renderLine(line, index))}
    </div>
  )
}
