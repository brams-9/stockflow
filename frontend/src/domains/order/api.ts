import { request } from '../../lib/apiClient'
import { orderSchema, type CreateOrderInput, type Order } from '../../types/api'

export const createOrder = (input: CreateOrderInput, idempotencyKey: string): Promise<Order> =>
  request({
    method: 'POST',
    path: '/api/orders',
    body: input,
    headers: { 'Idempotency-Key': idempotencyKey },
    schema: orderSchema,
  })

export const getOrder = (id: number): Promise<Order> =>
  request({
    method: 'GET',
    path: `/api/orders/${id}`,
    schema: orderSchema,
  })

export const cancelOrder = (id: number): Promise<Order> =>
  request({
    method: 'POST',
    path: `/api/orders/${id}/cancel`,
    schema: orderSchema,
  })
