import { afterEach, describe, expect, it, vi } from 'vitest'
import { createOrder } from './api'

const order = {
  id: 1,
  customer: 'Budi',
  status: 'CREATED',
  total: 2000,
  created_at: '2026-01-01T00:00:00Z',
  cancelled_at: null,
  items: [{ id: 1, sku: 'A-1', qty: 2, unit_price: 1000 }],
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('createOrder', () => {
  it('sends the idempotency key with the json headers', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(order), { status: 201 }))
    vi.stubGlobal('fetch', fetchMock)

    const result = await createOrder(
      { customer: 'Budi', items: [{ sku: 'A-1', qty: 2 }] },
      'abc-123',
    )

    expect(result.id).toBe(1)
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toBe('/api/orders')
    expect(init.method).toBe('POST')
    expect(init.headers).toEqual({
      'Content-Type': 'application/json',
      'Idempotency-Key': 'abc-123',
    })
    expect(JSON.parse(init.body as string)).toEqual({
      customer: 'Budi',
      items: [{ sku: 'A-1', qty: 2 }],
    })
  })

  it('reuses the same key when the caller retries', async () => {
    const fetchMock = vi
      .fn()
      .mockRejectedValueOnce(new TypeError('network'))
      .mockResolvedValueOnce(new Response(JSON.stringify(order), { status: 201 }))
    vi.stubGlobal('fetch', fetchMock)
    const input = { customer: 'Budi', items: [{ sku: 'A-1', qty: 2 }] }

    await expect(createOrder(input, 'same-key')).rejects.toThrow()
    await createOrder(input, 'same-key')

    const keys = fetchMock.mock.calls.map(
      ([, init]) => (init as RequestInit & { headers: Record<string, string> }).headers['Idempotency-Key'],
    )
    expect(keys).toEqual(['same-key', 'same-key'])
  })
})
