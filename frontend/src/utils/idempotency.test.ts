import { describe, expect, it } from 'vitest'
import { createKeyHolder } from './idempotency'

const sequence = () => {
  let count = 0
  return () => `key-${++count}`
}

describe('createKeyHolder', () => {
  it('generates lazily and keeps the same key across reads', () => {
    const holder = createKeyHolder(sequence())
    expect(holder.get()).toBe('key-1')
    expect(holder.get()).toBe('key-1')
  })

  it('does not generate until it is asked', () => {
    let calls = 0
    const holder = createKeyHolder(() => `key-${++calls}`)
    expect(calls).toBe(0)
    holder.get()
    expect(calls).toBe(1)
  })

  it('issues a new key after a reset', () => {
    const holder = createKeyHolder(sequence())
    const first = holder.get()
    holder.reset()
    const second = holder.get()
    expect(second).not.toBe(first)
    expect(holder.get()).toBe(second)
  })

  it('treats repeated resets as a single reset', () => {
    const holder = createKeyHolder(sequence())
    holder.get()
    holder.reset()
    holder.reset()
    expect(holder.get()).toBe('key-2')
  })

  it('produces uuid keys by default', () => {
    const key = createKeyHolder().get()
    expect(key).toMatch(/^[0-9a-f-]{36}$/)
  })
})
