import { request } from '../../lib/apiClient'
import { orderSchema, type CreateOrderInput, type Order } from '../../types/api'

export const createOrder = (input: CreateOrderInput): Promise<Order> =>
  request({
    method: 'POST',
    path: '/api/orders',
    body: input,
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
