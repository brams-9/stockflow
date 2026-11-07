import { describe, expect, it } from 'vitest'
import { formatRelativeTime, formatRupiah, isLowStock } from './format'

describe('formatRupiah', () => {
  it('uses dot separators', () => {
    expect(formatRupiah(1234567)).toBe('Rp 1.234.567')
  })

  it('formats zero and small values', () => {
    expect(formatRupiah(0)).toBe('Rp 0')
    expect(formatRupiah(500)).toBe('Rp 500')
  })
})

describe('formatRelativeTime', () => {
  const now = new Date('2026-03-10T12:00:00Z').getTime()
  const ago = (ms: number) => new Date(now - ms).toISOString()

  it('returns a dash when there is no timestamp', () => {
    expect(formatRelativeTime(null, now)).toBe('-')
  })

  it('handles seconds, minutes, hours and days', () => {
    expect(formatRelativeTime(ago(20_000), now)).toBe('just now')
    expect(formatRelativeTime(ago(5 * 60_000), now)).toBe('5m ago')
    expect(formatRelativeTime(ago(2 * 3_600_000), now)).toBe('2h ago')
    expect(formatRelativeTime(ago(3 * 86_400_000), now)).toBe('3d ago')
  })

  it('handles months and years', () => {
    expect(formatRelativeTime(ago(65 * 86_400_000), now)).toBe('2mo ago')
    expect(formatRelativeTime(ago(400 * 86_400_000), now)).toBe('1y ago')
  })

  it('treats future timestamps as just now', () => {
    expect(formatRelativeTime(ago(-60_000), now)).toBe('just now')
  })
})

describe('isLowStock', () => {
  it('is low at the boundary of 10 and below', () => {
    expect(isLowStock(0)).toBe(true)
    expect(isLowStock(10)).toBe(true)
  })

  it('is not low above 10', () => {
    expect(isLowStock(11)).toBe(false)
  })
})
