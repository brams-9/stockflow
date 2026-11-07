type BadgeProps = {
  tone: 'green' | 'gray' | 'red' | 'amber'
  label: string
}

export default function Badge({ tone, label }: BadgeProps) {
  return <span className={`badge badge-${tone}`}>{label}</span>
}
