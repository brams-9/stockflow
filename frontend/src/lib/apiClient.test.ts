import { afterEach, describe, expect, it, vi } from 'vitest'
import { z } from 'zod'
import { ApiError, request } from './apiClient'

const respond = (status: number, body: string) =>
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(body, { status })))

const failure = async () => {
  try {
    await request({ method: 'GET', path: '/api/products', schema: z.unknown() })
  } catch (error) {
    return error
  }
  return undefined
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('request errors', () => {
  it('shows a friendly message for gateway failures', async () => {
    respond(502, '<h1>Bad Gateway</h1>')
    const error = await failure()
    expect(error).toBeInstanceOf(ApiError)
    expect((error as ApiError).message).toBe('Server is unavailable, try again')
    expect((error as ApiError).status).toBe(502)
  })

  it('treats a gateway status as unavailable even with a JSON body', async () => {
    respond(503, JSON.stringify({ error: { code: 'INTERNAL_ERROR', message: 'upstream' } }))
    const error = await failure()
    expect((error as ApiError).message).toBe('Server is unavailable, try again')
  })

  it('keeps the server message for envelope errors', async () => {
    respond(409, JSON.stringify({ error: { code: 'SKU_ALREADY_EXISTS', message: 'taken' } }))
    const error = await failure()
    expect((error as ApiError).code).toBe('SKU_ALREADY_EXISTS')
    expect((error as ApiError).message).toBe('taken')
  })

  it('reports network failures with status 0', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('failed')))
    const error = await failure()
    expect((error as ApiError).status).toBe(0)
    expect((error as ApiError).code).toBe('NETWORK_ERROR')
  })
})
