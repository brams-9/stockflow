import type { ReactNode } from 'react'

type CardProps = {
  title: string
  children: ReactNode
}

export default function Card({ title, children }: CardProps) {
  return (
    <section>
      <h2>{title}</h2>
      <div>{children}</div>
    </section>
  )
}
