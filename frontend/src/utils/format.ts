export const LOW_STOCK_LIMIT = 10

const rupiah = new Intl.NumberFormat('id-ID')

export const formatRupiah = (amount: number): string => `Rp ${rupiah.format(amount)}`

export const isLowStock = (stock: number): boolean => stock <= LOW_STOCK_LIMIT

export const formatRelativeTime = (iso: string | null, now: number = Date.now()): string => {
  if (!iso) return '-'
  const seconds = Math.floor((now - new Date(iso).getTime()) / 1000)
  if (seconds < 60) return 'just now'
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}d ago`
  const months = Math.floor(days / 30)
  if (months < 12) return `${months}mo ago`
  return `${Math.floor(months / 12)}y ago`
}

export const formatDateTime = (iso: string | null): string => {
  if (!iso) return '-'
  return new Date(iso).toLocaleString('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
  })
}
