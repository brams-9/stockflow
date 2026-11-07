import { z } from 'zod'
import { apiErrorSchema, errorCodes } from '../types/api'

const baseUrl: string = import.meta.env.VITE_API_BASE_URL ?? ''

type Method = 'GET' | 'POST' | 'PUT' | 'DELETE'

type QueryValue = string | number | boolean | undefined | null

type RequestOptions<T extends z.ZodType> = {
  method: Method
  path: string
  schema: T
  query?: Record<string, QueryValue>
  body?: unknown
}

type ApiErrorInit = {
  status: number
  code: string
  message: string
  fields?: Record<string, string>
  sku?: string
  available?: number
}

export class ApiError extends Error {
  status: number
  code: string
  fields?: Record<string, string>
  sku?: string
  available?: number

  constructor({ status, code, message, fields, sku, available }: ApiErrorInit) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.fields = fields
    this.sku = sku
    this.available = available
  }
}

const buildUrl = (path: string, query?: Record<string, QueryValue>) => {
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(query ?? {})) {
    if (value === undefined || value === null || value === '') continue
    params.set(key, String(value))
  }
  const search = params.toString()
  return `${baseUrl}${path}${search ? `?${search}` : ''}`
}

const readJson = async (res: Response): Promise<unknown> => {
  const text = await res.text()
  if (!text) return undefined
  return JSON.parse(text)
}

const parseErrorBody = async (res: Response) => {
  try {
    return apiErrorSchema.safeParse(await readJson(res))
  } catch {
    return undefined
  }
}

const toApiError = async (res: Response) => {
  const parsed = await parseErrorBody(res)
  if (!parsed?.success) {
    return new ApiError({
      status: res.status,
      code: errorCodes.unknown,
      message: res.statusText || 'Unexpected error',
    })
  }
  const { code, message, fields, sku, available } = parsed.data.error
  return new ApiError({
    status: res.status,
    code,
    message: message ?? res.statusText,
    fields,
    sku,
    available,
  })
}

export async function request<T extends z.ZodType>({
  method,
  path,
  schema,
  query,
  body,
}: RequestOptions<T>): Promise<z.infer<T>> {
  let res: Response
  try {
    res = await fetch(buildUrl(path, query), {
      method,
      headers: body === undefined ? undefined : { 'Content-Type': 'application/json' },
      body: body === undefined ? undefined : JSON.stringify(body),
    })
  } catch {
    throw new ApiError({
      status: 0,
      code: errorCodes.network,
      message: 'Unable to reach the server',
    })
  }

  if (!res.ok) throw await toApiError(res)

  let payload: unknown
  try {
    payload = res.status === 204 ? undefined : await readJson(res)
  } catch {
    throw new ApiError({
      status: res.status,
      code: errorCodes.invalidResponse,
      message: 'Server returned an unreadable response',
    })
  }

  const parsed = schema.safeParse(payload)
  if (!parsed.success) {
    throw new ApiError({
      status: res.status,
      code: errorCodes.invalidResponse,
      message: 'Server returned an unexpected response',
    })
  }
  return parsed.data
}
