import type { Summary } from '../domains/summary/summary'
import { formatRupiah } from '../utils/format'

type SummaryCardsProps = {
  summary?: Summary
}

export default function SummaryCards({ summary }: SummaryCardsProps) {
  const items = [
    { label: 'Total products', value: summary?.totalProducts.toString() },
    { label: 'Active products', value: summary?.activeProducts.toString() },
    { label: 'Stock value', value: summary ? formatRupiah(summary.stockValue) : undefined },
    { label: 'Low stock', value: summary?.lowStock.toString() },
  ]

  return (
    <div className="summary">
      {items.map((item) => (
        <div key={item.label} className="summary-card">
          <span className="summary-label">{item.label}</span>
          <strong className="summary-value">{item.value ?? '-'}</strong>
        </div>
      ))}
    </div>
  )
}
