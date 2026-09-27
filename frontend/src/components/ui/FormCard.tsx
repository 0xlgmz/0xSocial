import type { ReactNode } from 'react'

type FormCardProps = {
  eyebrow: string
  title: string
  description: ReactNode
  children: ReactNode
  footer?: ReactNode
}

export function FormCard({ eyebrow, title, description, children, footer }: FormCardProps) {
  return (
    <div className="card border border-base-300 bg-base-100 shadow-sm">
      <div className="card-body gap-5 p-6 sm:p-8">
        <header>
          <p className="font-mono text-[.65rem] font-medium uppercase tracking-[.2em] text-primary">{eyebrow}</p>
          <h1 className="mt-2 text-3xl font-semibold tracking-[-.035em] text-base-content">{title}</h1>
          <p className="mt-2 text-sm leading-6 text-base-content/55">{description}</p>
        </header>
        {children}
      </div>
      {footer && <div className="border-t border-base-300 px-6 py-4 text-center text-xs text-base-content/55 sm:px-8">{footer}</div>}
    </div>
  )
}
